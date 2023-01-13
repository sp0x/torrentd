package indexer

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage"
	"github.com/sp0x/torrentd/storage/indexing"
)

const (
	noIndexError           = "no index configured. you have to pick an index"
	saveResultsOnDiscovery = true
)

// Facade for an indexer/aggregate, helps manage the scope of the index, it's configuration and the index itself.
// All results are also stored after fetching them.
type Facade struct {
	Indexes     IndexCollection
	IndexScope  Scope
	Config      config.Config
	workerCount int
	storage     storage.ItemStorage
	logger      *log.Logger
}

// GenericSearchOptions options for the search.
type GenericSearchOptions struct {
	// The count of pages we can fetch
	PageCount uint
	// The initial search page
	StartingPage         uint
	MaxRequestsPerSecond uint
	StopOnStaleResults   bool
}

type workerJob struct {
	Fields   map[string]interface{}
	Page     uint
	id       string
	Index    Indexer
	Iterator *search.SearchStateIterator
}

// NewFacadeFromConfiguration Creates a new facade using the configuration
func NewFacadeFromConfiguration(cfg config.Config) (*Facade, error) {
	facade := NewEmptyFacade(cfg)
	indexName := cfg.GetString("index")
	if indexName == "" {
		return nil, errors.New(noIndexError)
	}
	log.WithFields(log.Fields{"name": indexName}).
		Debug("Creating new facade from configuration")
	indexes, err := facade.IndexScope.Lookup(cfg, indexName)
	if err != nil {
		log.WithFields(log.Fields{"name": indexName}).
			Error("Could not find indexes.")
		return nil, err
	}
	facade.Indexes = indexes
	return facade, nil
}

// NewEmptyFacade creates a new indexer facade with its own scope and config.
func NewEmptyFacade(cfg config.Config) *Facade {
	facade := &Facade{}
	facade.IndexScope = NewScope(getConfiguredIndexLoader(cfg).(DefinitionLoader))
	facade.Config = cfg
	facade.workerCount = cfg.GetInt("workerCount")
	facade.logger = log.New()
	facade.logger.SetLevel(config.GetMinLogLevel(cfg))
	facade.Indexes = IndexCollection{}
	if facade.workerCount < 1 {
		facade.workerCount = 1
	}
	return facade
}

// NewFacade Creates a new facade for an indexer with the given name and config.
// If any indexCategories are given, the facade must be for an indexer that supports these indexCategories.
// If you don't provide a name or name is `all`, an aggregate is used.
func NewFacade(indexNames string, cfg config.Config, cats ...categories.Category) (*Facade, error) {
	if newIndexSelector(indexNames).isAggregate() {
		return NewAggregateFacadeWithCategories(indexNames, cfg, cats...)
	}
	facade := NewEmptyFacade(cfg)
	indexes, err := facade.IndexScope.Lookup(cfg, indexNames)
	if err != nil {
		log.Errorf("Could not find Indexes `%s`.\n", indexNames)
		return nil, errors.New("indexer wasn't found")
	}
	if len(cats) > 0 {
		if !indexes.HasCategories(cats) {
			return nil, errors.New("indexer doesn't support the needed indexCategories")
		}
	}
	facade.Indexes = indexes
	return facade, nil
}

// NewAggregateFacadeWithCategories Finds an indexer from the config, that matches the given indexCategories.
func NewAggregateFacadeWithCategories(indexNames string, cfg config.Config, cats ...categories.Category) (*Facade, error) {
	facade := NewEmptyFacade(cfg)
	if indexNames == "" {
		indexNames = cfg.GetString("index")
	}
	if indexNames == "" {
		return nil, errors.New(noIndexError)
	}
	selector := newIndexSelector(indexNames)
	indexes, err := facade.IndexScope.LookupWithCategories(cfg, selector, cats)
	if err != nil {
		return nil, fmt.Errorf("Could not find Indexes `%s`.", indexNames)
	}
	facade.Indexes = indexes
	return facade, nil
}

func (f *Facade) OpenStorage() storage.ItemStorage {
	return getMultiIndexDatabase(f.Indexes, f.Config)
}

func (f *Facade) ensureDatabaseConnection() {
	if f.storage != nil {
		return
	}
	f.storage = f.OpenStorage()
}

func (f *Facade) Search(query *search.Query) (chan []search.ResultItemBase, error) {
	f.ensureDatabaseConnection()
	itemKey := indexing.NewKey("LocalID")
	err := f.storage.SetKey(itemKey)
	if err != nil {
		log.WithFields(log.Fields{}).Errorf("Couldn't get item index: %s\n", err)
		return nil, err
	}

	workerPool := f.createWorkerPool(f.Indexes, f.storage, query, f.workerCount)

	f.feedWorkerPool(workerPool)

	return workerPool.resultsChannel, nil
}

// SearchWithKeywords performs a search for a given page
func (f *Facade) SearchWithKeywords(query string, startingPage uint, pageCount uint) (chan []search.ResultItemBase, error) {
	queryObj, err := search.NewQueryFromQueryString(query)
	if err != nil {
		return nil, err
	}
	queryObj.Page = startingPage
	queryObj.NumberOfPagesToFetch = pageCount
	return f.Search(queryObj)
}

// SearchKeywordsWithCategory Search for *keywords* matching the needed category.
func (f *Facade) SearchKeywordsWithCategory(query string, page uint, cat categories.Category) (chan []search.ResultItemBase, error) {
	queryObj, err := search.NewQueryFromQueryString(query)
	if err != nil {
		return nil, err
	}
	queryObj.Page = page
	queryObj.Categories = []int{cat.ID}
	return f.Search(queryObj)
}

// GetDefaultSearchOptions gets the default search options
func (f *Facade) GetDefaultSearchOptions() *GenericSearchOptions {
	return &GenericSearchOptions{
		PageCount:            10,
		StartingPage:         0,
		MaxRequestsPerSecond: 1,
	}
}
