package indexer

import (
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/torznab"
)

//A facade for an indexer/aggregate.
type Facade struct {
	//pageSize uint
	Indexer Indexer
	Config  config.Config
}

type GenericSearchOptions struct {
	PageCount            uint
	StartingPage         uint
	MaxRequestsPerSecond uint
	StopOnStaleTorrents  bool
}

func NewFacade(config config.Config) *Facade {
	rt := Facade{}
	ixr := config.GetString("Indexer")
	if ixr == "" {
		ixr = "rutracker.org"
	}
	ixrObj, err := Lookup(config, ixr)
	if err != nil {
		log.Errorf("Could not find Indexer `%s`.\n", ixr)
		return nil
	}
	rt.Config = config
	rt.Indexer = ixrObj
	return &rt
}

//NewAggregateIndexerHelperWithCategories Finds an indexer from the config, that matches the given categories.
func NewAggregateIndexerHelperWithCategories(config config.Config, cats ...categories.Category) *Facade {
	rt := Facade{}
	indexerName := config.GetString("Indexer")
	ixrObj, err := CreateAggregateForCategories(config, cats)
	if err != nil {
		log.Errorf("Could not find Indexer `%s`.\n", indexerName)
		return nil
	}
	rt.Config = config
	rt.Indexer = ixrObj
	return &rt
}

//Search using a given query
func (th *Facade) Search(searchContext search.Instance, query torznab.Query) (search.Instance, error) {
	srch, err := th.Indexer.Search(query, searchContext)
	if err != nil {
		return nil, err
	}
	return srch, nil
}

//Open the search to a given page.
func (th *Facade) SearchKeywords(searchContext search.Instance, query string, page uint) (search.Instance, error) {
	qrobj := torznab.ParseQueryString(query)
	qrobj.Page = page
	return th.Search(searchContext, qrobj)
}

//SearchKeywordsWithCategory Search for *keywords* matching the needed category.
func (th *Facade) SearchKeywordsWithCategory(searchContext search.Instance, query string, cat categories.Category, page uint) (search.Instance, error) {
	qrobj := torznab.ParseQueryString(query)
	qrobj.Page = page
	qrobj.Categories = []int{cat.ID}
	srch, err := th.Indexer.Search(qrobj, searchContext)
	if err != nil {
		return nil, err
	}
	return srch, nil
}

//
func (th *Facade) GetDefaultOptions() *GenericSearchOptions {
	return &GenericSearchOptions{
		PageCount:            10,
		StartingPage:         0,
		MaxRequestsPerSecond: 1,
	}
}
