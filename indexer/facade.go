package indexer

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

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
	Storage     storage.ItemStorage
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
func NewFacadeFromConfiguration(cfg config.Config) *Facade {
	facade := NewEmptyFacade(cfg)
	indexerName := cfg.GetString("index")
	if indexerName == "" {
		fmt.Print(noIndexError)
		os.Exit(1)
	}
	log.Debugf("Creating new facade from configuration: %v\n", indexerName)
	indexes, err := facade.IndexScope.Lookup(cfg, indexerName)
	if err != nil {
		log.Errorf("Could not find Indexes `%s`.\n", indexerName)
		return nil
	}
	facade.Indexes = indexes
	return facade
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
		return nil, fmt.Errorf("Could not find Indexes `%s`.\n", indexNames)
	}
	facade.Indexes = indexes
	return facade, nil
}

func (f *Facade) OpenStorage() storage.ItemStorage {
	return getStorageForIndexes(f.Indexes, f.Config)
}

func (f *Facade) Search(query *search.Query) (chan []search.ResultItemBase, error) {
	indexStorage := f.OpenStorage()
	defer indexStorage.Close()
	itemKey := indexing.NewKey("LocalID")
	err := indexStorage.SetKey(itemKey)
	if err != nil {
		log.WithFields(log.Fields{}).Errorf("Couldn't get item index: %s\n", err)
		return nil, err
	}

	workerPool := f.createWorkerPool(f.Indexes, indexStorage, query, f.workerCount)

	f.feedWorkerPool(workerPool)

	return workerPool.resultsChannel, nil
}

// feedWorkerPool Iterate over the index search iterators and add the data to the work channel
func (f *Facade) feedWorkerPool(workerPool *indexWorkerPool) {
	doneIterators := make(map[interface{}]bool)
	for !(len(doneIterators) == len(workerPool.iterators)) {
		for indexForIterator, iterator := range workerPool.iterators {
			if doneIterators[iterator] {
				continue
			}

			fields, page := iterator.Next()
			workerPool.workChannel <- createWorkerJob(iterator, indexForIterator, fields, page)
			if iterator.IsComplete() {
				doneIterators[iterator] = true
				f.logger.Debugf("Completed iterator %p for index %v", iterator, indexForIterator.GetDefinition().Name)
			}
		}
	}
	close(workerPool.workChannel)
}

func createWorkerJob(iterator *search.SearchStateIterator, index Indexer, fields map[string]interface{}, page uint) *workerJob {
	return &workerJob{
		Iterator: iterator,
		Fields:   fields,
		Page:     page,
		id:       "",
		Index:    index,
	}
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

//region Workers

func (f *Facade) runSearchWorker(
	id int,
	query *search.Query,
	wg *sync.WaitGroup,
	resultStorage storage.ItemStorage,
	workChannel <-chan *workerJob,
	resultsChannel chan<- []search.ResultItemBase) {

	for workState := range workChannel {
		searchResults, err := workState.Index.Search(query, workState)
		if err != nil {
			log.WithFields(log.Fields{}).Errorf("Couldn't persist item: %s\n", err)
			continue
		}

		if saveResultsOnDiscovery {
			saveDiscoveredItems(searchResults, resultStorage)
		}

		workState.Iterator.ScanForStaleResults(searchResults)

		if searchResults != nil {
			resultsChannel <- searchResults
		}
	}

	wg.Done()
	f.logger.Debugf("Worker #%v ran out of jobs", id)
}

func saveDiscoveredItems(searchResults []search.ResultItemBase, resultStorage storage.ItemStorage) {
	for _, item := range searchResults {
		err := resultStorage.Add(item)
		if err != nil {
			log.WithFields(log.Fields{}).Errorf("Couldn't persist item: %s\n", err)
		}
	}
}

type indexWorkerPool struct {
	storage             storage.ItemStorage
	completionWaitGroup sync.WaitGroup
	iterators           map[Indexer]*search.SearchStateIterator
	query               *search.Query
	workChannel         chan *workerJob
	resultsChannel      chan []search.ResultItemBase
}

/// createWorkerPool Creates a pool of workers that run in the background using work and results channels
func (f Facade) createWorkerPool(indexes []Indexer, resultStorage storage.ItemStorage, query *search.Query, workerCount int) *indexWorkerPool {
	workerPool := &indexWorkerPool{}
	workerPool.workChannel = make(chan *workerJob, workerCount)
	workerPool.resultsChannel = make(chan []search.ResultItemBase, workerCount)
	workerPool.storage = resultStorage
	workerPool.iterators = make(map[Indexer]*search.SearchStateIterator)

	for w := 0; w < workerCount; w++ {
		workerPool.completionWaitGroup.Add(1)
		go f.runSearchWorker(w, query, &workerPool.completionWaitGroup, resultStorage, workerPool.workChannel, workerPool.resultsChannel)
	}

	for _, index := range indexes {
		workerPool.iterators[index] = search.NewIterator(query)
	}

	go func() {
		// Wait for pool to be complete and close the results channel
		workerPool.completionWaitGroup.Wait()
		close(workerPool.resultsChannel)
	}()

	return workerPool
}

//endregion

//region Worker job

func (s workerJob) String() string {
	output := make([]string, len(s.Fields)+1)
	i := 0
	output[0] = fmt.Sprintf("page: %d", s.Page)
	for fname, fval := range s.Fields {
		val := fmt.Sprintf("{%s: %v}", fname, fval)
		output[i+1] = val
		i++
	}
	return strings.Join(output, ",")
}

func (s *workerJob) SetID(id string) {
	s.id = id
}

//endregion
