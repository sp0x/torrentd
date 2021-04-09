package indexer

import (
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/search"
)

const (
	noIndexError = "no index configured. you have to pick an index"
)

// Facade for an indexer/aggregate, helps manage the scope of the index, it's configuration and the index itself.
type Facade struct {
	// The indexer that we're using
	Index         Indexer
	LoadedIndexes Scope
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

// NewFacadeFromConfiguration Creates a new facade using the configuration
func NewFacadeFromConfiguration(config config.Config) *Facade {
	facade := NewEmptyFacade(config)
	indexerName := config.GetString("index")
	if indexerName == "" {
		fmt.Print(noIndexError)
		os.Exit(1)
	}
	log.Debugf("Creating new facade from configuration: %v\n", indexerName)
	index, err := facade.LoadedIndexes.Lookup(config, indexerName)
	if err != nil {
		log.Errorf("Could not find Index `%s`.\n", indexerName)
		return nil
	}
	facade.Index = index
	return facade
}

// NewEmptyFacade creates a new indexer facade with it's own scope and config.
func NewEmptyFacade(config config.Config) *Facade {
	facade := &Facade{}
	facade.LoadedIndexes = NewScope()
	return facade
}

// NewFacade Creates a new facade for an indexer with the given name and config.
// If any indexCategories are given, the facade must be for an indexer that supports these indexCategories.
// If you don't provide a name or name is `all`, an aggregate is used.
func NewFacade(indexerName string, config config.Config, cats ...categories.Category) (*Facade, error) {
	if newIndexerSelector(indexerName).isAggregate() {
		return NewAggregateFacadeWithCategories(config, cats...), nil
	}
	facade := NewEmptyFacade(config)
	indexerObj, err := facade.LoadedIndexes.Lookup(config, indexerName)
	if err != nil {
		log.Errorf("Could not find Index `%s`.\n", indexerObj)
		return nil, errors.New("indexer wasn't found")
	}
	if len(cats) > 0 {
		if !indexerObj.Capabilities().HasCategories(cats) {
			return nil, errors.New("indexer doesn't support the needed indexCategories")
		}
	}
	facade.Index = indexerObj
	return facade, nil
}

// NewAggregateFacadeWithCategories Finds an indexer from the config, that matches the given indexCategories.
func NewAggregateFacadeWithCategories(config config.Config, cats ...categories.Category) *Facade {
	facade := Facade{}
	facade.LoadedIndexes = NewScope()
	indexerName := config.GetString("index")
	if indexerName == "" {
		fmt.Print(noIndexError)
		os.Exit(1)
	}
	selector := newIndexerSelector(indexerName)
	ixrObj, err := facade.LoadedIndexes.CreateAggregateForCategories(config, selector, cats)
	if err != nil {
		log.Errorf("Could not find Index `%s`.\n", indexerName)
		return nil
	}
	facade.Index = ixrObj
	return &facade
}

// Search using a given query. The search covers only 1 page.
func (th *Facade) Search(searchInst search.Instance, query *search.Query) (search.Instance, error) {
	srch, err := th.Index.Search(query, searchInst)
	if err != nil {
		return nil, err
	}
	return srch, nil
}

// SearchKeywords performs a search for a given page
func (th *Facade) SearchKeywords(searchContext search.Instance, query string, page uint) (search.Instance, error) {
	qrobj, err := search.NewQueryFromQueryString(query)
	if err != nil {
		return nil, err
	}
	qrobj.Page = page
	return th.Search(searchContext, qrobj)
}

// SearchKeywordsWithCategory Search for *keywords* matching the needed category.
func (th *Facade) SearchKeywordsWithCategory(srch search.Instance, query string, cat categories.Category, page uint) (search.Instance, error) {
	qrobj, err := search.NewQueryFromQueryString(query)
	if err != nil {
		return nil, err
	}
	qrobj.Page = page
	qrobj.Categories = []int{cat.ID}
	return th.Search(srch, qrobj)
}

// GetDefaultSearchOptions gets the default search options
func (th *Facade) GetDefaultSearchOptions() *GenericSearchOptions {
	return &GenericSearchOptions{
		PageCount:            10,
		StartingPage:         0,
		MaxRequestsPerSecond: 1,
	}
}
