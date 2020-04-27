package indexer

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"github.com/sp0x/rutracker-rss/torznab"
	"io"
	"net/http"

	"golang.org/x/sync/errgroup"
)

type Aggregate struct {
	Indexers []Indexer
}

func (ag *Aggregate) Open(s *search.ExternalResultItem) (io.ReadCloser, error) {
	//Find the indexer
	for _, ixr := range ag.Indexers {
		nfo := ixr.Info()
		if nfo.GetTitle() == s.Site {
			return ixr.Open(s)
		}
	}
	return nil, errors.New("couldn't find indexer")
}

func NewAggregate(indexers []Indexer) *Aggregate {
	ag := &Aggregate{}
	ag.Indexers = indexers
	return ag
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

//Check checks all indexers
func (ag *Aggregate) Check() error {
	g := errgroup.Group{}
	for _, ixr := range ag.Indexers {
		indexerID := ixr.Info().GetId()
		//Run the indexer in a goroutine
		g.Go(func() error {
			_, err := ixr.Search(torznab.Query{}, nil)
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

func (ag *Aggregate) Search(query torznab.Query, srch search.Instance) (search.Instance, error) {
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
	//indexerSearches := make(map[int]*search.Search)
	// fetch all results
	for idx, indexer := range ag.Indexers {
		indexerID := indexer.Info().GetId()
		ixrSearch := aggSearch.SearchContexts[&indexer]

		//Run the indexer in a goroutine
		g.Go(func() error {
			srchRes, err := indexer.Search(query, ixrSearch)
			if err != nil {
				log.Warnf("Indexer %q failed: %s", indexerID, err)
				return nil
			}
			aggSearch.SearchContexts[&indexer] = srchRes
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

func (ag *Aggregate) Download(u string) (io.ReadCloser, error) {
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
