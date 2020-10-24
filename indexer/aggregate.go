package indexer

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage"
	"github.com/sp0x/torrentd/torznab"
	"io"
	"net/http"
	"strings"

	"golang.org/x/sync/errgroup"
)

type Aggregate struct {
	Indexers []Indexer
	Storage  storage.ItemStorage
}

func (ag *Aggregate) SetStorage(s storage.ItemStorage) {
	ag.Storage = s
}

func (ag *Aggregate) GetDefinition() *IndexerDefinition {
	definition := &IndexerDefinition{}
	definition.Site = "aggregate"
	var indexerNames []string
	for _, ixr := range ag.Indexers {
		indexerNames = append(indexerNames, ixr.GetDefinition().Name)
	}
	definition.Name = strings.Join(indexerNames, ",")
	return definition
}

func (ag *Aggregate) Open(s *search.ExternalResultItem) (io.ReadCloser, error) {
	//Find the Indexer
	for _, ixr := range ag.Indexers {
		nfo := ixr.Info()
		if nfo.GetTitle() == s.Site {
			return ixr.Open(s)
		}
	}
	return nil, errors.New("couldn't find Indexer")
}

//MaxSearchPages returns the maximum number of pages that this aggregate can search, this is using the maximum paged index in the aggregate.
func (ag *Aggregate) MaxSearchPages() uint {
	maxValue := uint(0)
	for _, index := range ag.Indexers {
		if index.MaxSearchPages() > maxValue {
			maxValue = index.MaxSearchPages()
		}
	}
	return maxValue
}

//SearchIsSinglePaged this is true only if all indexes inside the aggregate are single paged.
func (ag *Aggregate) SearchIsSinglePaged() bool {
	//For this, all indexes must be single paged
	for _, index := range ag.Indexers {
		if !index.SearchIsSinglePaged() {
			return false
		}
	}
	return true
}

func (ag *Aggregate) ProcessRequest(req *http.Request) (*http.Response, error) {
	for _, indexer := range ag.Indexers {
		return indexer.ProcessRequest(req)
	}
	return nil, nil
}

func (ag *Aggregate) GetEncoding() string {
	for _, indexer := range ag.Indexers {
		return indexer.GetEncoding()
	}
	return "utf-8"
}

//Check checks all indexes, if they can be searched.
func (ag *Aggregate) Check() error {
	g := errgroup.Group{}
	for _, ixr := range ag.Indexers {
		indexerID := ixr.Info().GetId()
		//Run the Indexer in a goroutine
		g.Go(func() error {
			_, err := ixr.Search(&torznab.Query{}, nil)
			if err != nil {
				log.Warnf("Indexer %q failed: %s", indexerID, err)
				return nil
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Warn(err)
		return err
	}
	return nil
}

func (ag *Aggregate) Search(query *torznab.Query, srch search.Instance) (search.Instance, error) {
	g := errgroup.Group{}
	allResults := make([][]search.ExternalResultItem, len(ag.Indexers))
	maxLength := 0
	if srch == nil {
		srch = search.NewAggregatedSearch()
	}
	switch srch.(type) {
	case *search.AggregatedSearch:
	default:
		return nil, errors.New("can't use normal search on an aggregate")
	}
	aggSearch := srch.(*search.AggregatedSearch)
	//indexerSearches := make(map[int]*search.SearchKeywords)
	// fetch all results
	if ag.Indexers == nil {
		log.Warn("aggregate has no indexes")
	}
	for idx, pIndexer := range ag.Indexers {
		//Run the Indexer in a goroutine
		idx, pIndexer := idx, pIndexer
		g.Go(func() error {
			indexerID := pIndexer.Info().GetId()
			ixrSearch := aggSearch.SearchContexts[&pIndexer]
			//log.WithFields(log.Fields{"Indexer": indexerID}).
			//	Info("Aggregate index search")
			srchRes, err := pIndexer.Search(query, ixrSearch)
			if err != nil {
				log.Warnf("Indexer %q failed: %s", indexerID, err)
				return nil
			}
			aggSearch.SearchContexts[&pIndexer] = srchRes
			allResults[idx] = srchRes.GetResults()
			if l := len(srchRes.GetResults()); l > maxLength {
				maxLength = l
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Warn(err)
		return nil, err
	}
	var results []search.ExternalResultItem

	// interleave search results to preserve ordering
	for i := 0; i <= maxLength; i++ {
		for _, r := range allResults {
			if len(r) > i {
				results = append(results, r[i])
			}
		}
	}

	if query.Limit > 0 && len(results) > query.Limit {
		results = results[:query.Limit]
	}
	aggSearch.SetResults(results)
	return aggSearch, nil
}

func (ag *Aggregate) Capabilities() torznab.Capabilities {
	return torznab.Capabilities{
		SearchModes: []search.SearchMode{
			{Key: "movie-search", Available: true, SupportedParams: []string{"q", "imdbid"}},
			{Key: "tv-search", Available: true, SupportedParams: []string{"q", "season", "ep"}},
			{Key: "search", Available: true, SupportedParams: []string{"q"}},
		},
	}
}

func (ag *Aggregate) Download(string) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

type AggregateInfo struct{}

func (a *AggregateInfo) GetLanguage() string {
	return "en-US"
}
func (a *AggregateInfo) GetLink() string {
	return ""
}
func (a *AggregateInfo) GetTitle() string {
	return "Aggregated Indexer"
}
func (a *AggregateInfo) GetId() string {
	return "aggregate"
}

func (ag *Aggregate) Info() Info {
	return &AggregateInfo{}
}
