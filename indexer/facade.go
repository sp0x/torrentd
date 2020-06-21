package indexer

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/torznab"
)

//Facade for an indexer/aggregate, helps manage the scope of the index, it's configuration and the index itself.
type Facade struct {
	//The indexer that we're using
	Indexer Indexer
	//Configuration for the indexer.
	Config config.Config
	Scope  Scope
}

//GenericSearchOptions options for the search.
type GenericSearchOptions struct {
	//The count of pages we can fetch
	PageCount uint
	//The initial search page
	StartingPage         uint
	MaxRequestsPerSecond uint
	StopOnStaleResults   bool
}

//NewFacadeFromConfiguration Creates a new facade using the configuration
func NewFacadeFromConfiguration(config config.Config) *Facade {
	facade := NewEmptyFacade(config)
	indexerName := config.GetString("Indexer")
	if indexerName == "" {
		indexerName = "rutracker.org"
	}
	ixrObj, err := facade.Scope.Lookup(config, indexerName)
	if err != nil {
		log.Errorf("Could not find Indexer `%s`.\n", indexerName)
		return nil
	}
	facade.Indexer = ixrObj
	return facade
}

//NewEmptyFacade creates a new indexer facade with it's own scope and config.
func NewEmptyFacade(config config.Config) *Facade {
	facade := &Facade{}
	facade.Scope = NewScope()
	facade.Config = config
	return facade
}

//NewFacade Creates a new facade for an indexer with the given name and config.
//If any categories are given, the facade must be for an indexer that supports these categories.
//If you don't provide a name or name is `all`, an aggregate is used.
func NewFacade(indexerName string, config config.Config, cats ...categories.Category) (*Facade, error) {
	if indexerName == "" || indexerName == "all" {
		return NewAggregateFacadeWithCategories(config, cats...), nil
	}
	facade := NewEmptyFacade(config)
	indexerObj, err := facade.Scope.Lookup(config, indexerName)
	if err != nil {
		log.Errorf("Could not find Indexer `%s`.\n", indexerObj)
		return nil, errors.New("indexer wasn't found")
	}
	if len(cats) > 0 {
		if !indexerObj.Capabilities().HasCategories(cats) {
			return nil, errors.New("indexer doesn't support the needed categories")
		}
	}
	facade.Indexer = indexerObj
	return facade, nil
}

//NewAggregateFacadeWithCategories Finds an indexer from the config, that matches the given categories.
func NewAggregateFacadeWithCategories(config config.Config, cats ...categories.Category) *Facade {
	facade := Facade{}
	facade.Scope = NewScope()
	indexerName := config.GetString("Indexer")
	ixrObj, err := facade.Scope.CreateAggregateForCategories(config, cats)
	if err != nil {
		log.Errorf("Could not find Indexer `%s`.\n", indexerName)
		return nil
	}
	facade.Config = config
	facade.Indexer = ixrObj
	return &facade
}

//Search using a given query
func (th *Facade) Search(searchContext search.Instance, query *torznab.Query) (search.Instance, error) {
	srch, err := th.Indexer.Search(query, searchContext)
	if err != nil {
		return nil, err
	}
	return srch, nil
}

//SearchKeywords performs a search for a given page
func (th *Facade) SearchKeywords(searchContext search.Instance, query string, page uint) (search.Instance, error) {
	qrobj := torznab.ParseQueryString(query)
	qrobj.Page = page
	return th.Search(searchContext, &qrobj)
}

//SearchKeywordsWithCategory Search for *keywords* matching the needed category.
func (th *Facade) SearchKeywordsWithCategory(searchContext search.Instance, query string, cat categories.Category, page uint) (search.Instance, error) {
	qrobj := torznab.ParseQueryString(query)
	qrobj.Page = page
	qrobj.Categories = []int{cat.ID}
	srch, err := th.Indexer.Search(&qrobj, searchContext)
	if err != nil {
		return nil, err
	}
	return srch, nil
}

//GetDefaultOptions gets the default search options
func (th *Facade) GetDefaultOptions() *GenericSearchOptions {
	return &GenericSearchOptions{
		PageCount:            10,
		StartingPage:         0,
		MaxRequestsPerSecond: 1,
	}
}
