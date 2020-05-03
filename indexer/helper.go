package indexer

import (
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/config"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"github.com/sp0x/rutracker-rss/torznab"
)

type IndexerHelper struct {
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

func NewIndexerHelper(config config.Config) *IndexerHelper {
	rt := IndexerHelper{}
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

//Open the search to a given page.
func (th *IndexerHelper) SearchKeywords(searchContext search.Instance, query string, page uint) (search.Instance, error) {
	qrobj := torznab.ParseQueryString(query)
	qrobj.Page = page
	srch, err := th.Indexer.Search(qrobj, searchContext)
	if err != nil {
		return nil, err
	}
	return srch, nil
}

func (th *IndexerHelper) GetDefaultOptions() *GenericSearchOptions {
	return &GenericSearchOptions{
		PageCount:            10,
		StartingPage:         0,
		MaxRequestsPerSecond: 1,
	}
}
