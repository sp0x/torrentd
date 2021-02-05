package indexer

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage"
	"github.com/sp0x/torrentd/torznab"
)

const aggregateSiteName = "aggregate"

type Aggregate struct {
	Indexers []Indexer
	Storage  storage.ItemStorage
	selector *Selector
}

func (ag *Aggregate) SetStorage(s storage.ItemStorage) {
	ag.Storage = s
}

func (ag *Aggregate) GetStorage() storage.ItemStorage {
	return ag.Storage
}

func (ag *Aggregate) Errors() []string {
	var errs []string
	for _, index := range ag.Indexers {
		indexErrors := index.Errors()
		site := index.GetDefinition().Site
		for _, indexError := range indexErrors {
			errs = append(errs, fmt.Sprintf("%s: %s", site, indexError))
		}
	}
	return errs
}

func (ag *Aggregate) GetDefinition() *Definition {
	definition := &Definition{}
	definition.Site = aggregateSiteName
	indexerNames := make([]string, len(ag.Indexers))
	for _, ixr := range ag.Indexers {
		indexerNames = append(indexerNames, ixr.GetDefinition().Name)
	}
	definition.Name = strings.Join(indexerNames, ",")
	return definition
}

func (ag *Aggregate) Open(scrapeItem search.ResultItemBase) (*ResponseProxy, error) {
	// Find the Indexer
	scrapeItemRoot := scrapeItem.AsScrapeItem()
	for _, ixr := range ag.Indexers {
		nfo := ixr.Info()
		if nfo.GetTitle() == scrapeItemRoot.Site {
			return ixr.Open(scrapeItem)
		}
	}
	return nil, errors.New("couldn't find Indexer")
}

// MaxSearchPages returns the maximum number of pages that this aggregate can search, this is using the maximum paged index in the aggregate.
func (ag *Aggregate) MaxSearchPages() uint {
	maxValue := uint(0)
	for _, index := range ag.Indexers {
		if index.MaxSearchPages() > maxValue {
			maxValue = index.MaxSearchPages()
		}
	}
	return maxValue
}

// SearchIsSinglePaged this is true only if all indexesCollection inside the aggregate are single paged.
func (ag *Aggregate) SearchIsSinglePaged() bool {
	// For this, all indexesCollection must be single paged
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

func (ag *Aggregate) Site() string {
	sites := ""
	for _, indx := range ag.Indexers {
		sites += indx.Site() + ","
	}
	sites = strings.TrimRight(sites, ",")
	return sites
}

// HealthCheck checks all indexesCollection, if they can be searched.
func (ag *Aggregate) HealthCheck() error {
	errorGroup := errgroup.Group{}
	for _, ixr := range ag.Indexers {
		indexerID := ixr.Info().GetID()
		// Run the Indexer in a goroutine
		errorGroup.Go(func() error {
			err := ixr.HealthCheck()
			if err != nil {
				log.Warnf("Indexer %q failed: %s", indexerID, err)
				return nil
			}
			return nil
		})
	}
	if err := errorGroup.Wait(); err != nil {
		log.Warn(err)
		return err
	}
	return nil
}

func (ag *Aggregate) Search(query *search.Query, searchInstance search.Instance) (search.Instance, error) {
	errorGroup := errgroup.Group{}
	allResults := make([][]search.ResultItemBase, len(ag.Indexers))
	maxLength := 0
	if searchInstance == nil {
		searchInstance = search.NewAggregatedSearch()
	}
	switch searchInstance.(type) {
	case *search.AggregatedSearch:
	default:
		return nil, errors.New("can't use normal search on an aggregate")
	}
	aggregatedSearch := searchInstance.(*search.AggregatedSearch)
	// indexerSearches := make(map[int]*search.SearchKeywords)
	// fetch all results
	if ag.Indexers == nil {
		log.Warnf("searching an aggregate[%s] that has no indexesCollection", ag.selector)
		return nil, errors.New("no indexesCollection are set for this aggregate")
	}
	for idx, pIndexer := range ag.Indexers {
		// Run the Indexer in a goroutine
		i, pIndex := idx, pIndexer
		if len(pIndexer.Errors()) > 0 {
			log.WithFields(log.Fields{"index": pIndexer}).Debug("Skipping index because it has errors")
			continue
		}
		errorGroup.Go(func() error {
			indexID := pIndex.Info().GetID()
			indexSearch := aggregatedSearch.SearchContexts[&pIndex]
			log.WithFields(log.Fields{"Indexer": indexID}).
				Debug("Aggregate index search")
			resultingSearch, err := pIndex.Search(query, indexSearch)
			if err != nil {
				log.Warnf("Indexer %q failed: %s", indexID, err)
				return nil
			}
			aggregatedSearch.SearchContexts[&pIndex] = resultingSearch
			allResults[i] = resultingSearch.GetResults()
			if l := len(resultingSearch.GetResults()); l > maxLength {
				maxLength = l
			}
			return nil
		})
	}
	if err := errorGroup.Wait(); err != nil {
		log.Warn(err)
		return nil, err
	}
	var results []search.ResultItemBase

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
	aggregatedSearch.SetResults(results)
	return aggregatedSearch, nil
}

func (ag *Aggregate) Capabilities() torznab.Capabilities {
	return torznab.Capabilities{
		SearchModes: []search.Mode{
			{Key: "movie-search", Available: true, SupportedParams: []string{"q", "imdbid"}},
			{Key: "tv-search", Available: true, SupportedParams: []string{"q", "season", "ep"}},
			{Key: "search", Available: true, SupportedParams: []string{"q"}},
		},
	}
}

func (ag *Aggregate) Download(string) (*ResponseProxy, error) {
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

func (a *AggregateInfo) GetID() string {
	return aggregateSiteName
}

func (ag *Aggregate) Info() Info {
	return &AggregateInfo{}
}
