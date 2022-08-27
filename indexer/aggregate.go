package indexer

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage"
	"github.com/sp0x/torrentd/torznab"
)

const aggregateSiteName = "aggregate"

type Aggregate struct {
	Indexes  []Indexer
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
	for _, index := range ag.Indexes {
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
	indexNames := make([]string, len(ag.Indexes))
	for i, ixr := range ag.Indexes {
		indexNames[i] = ixr.GetDefinition().Name
	}
	definition.Name = strings.Join(indexNames, ",")
	return definition
}

// SearchIsSinglePaged this is true only if all indexMap inside the aggregate are single paged.
func (ag *Aggregate) SearchIsSinglePaged() bool {
	// For this, all indexMap must be single paged
	for _, index := range ag.Indexes {
		if !index.SearchIsSinglePaged() {
			return false
		}
	}
	return true
}

func (ag *Aggregate) GetEncoding() string {
	for _, indexer := range ag.Indexes {
		return indexer.GetEncoding()
	}
	return "utf-8"
}

func (ag *Aggregate) Site() string {
	sites := ""
	for _, indx := range ag.Indexes {
		sites += indx.Site() + ","
	}
	sites = strings.TrimRight(sites, ",")
	return sites
}

//
//func (ag *Aggregate) Search(query *search.Query, _ *workerJob) ([]search.ResultItemBase, error) {
//	errorGroup := errgroup.Group{}
//	allResults := make([][]search.ResultItemBase, len(ag.Indexes))
//	maxLength := 0
//
//	if ag.Indexes == nil {
//		log.Warnf("searching an aggregate[%s] that has no indexMap", ag.selector)
//		return nil, errors.New("no indexMap are set for this aggregate")
//	}
//	// TODO: This should be done in the facade instead, since it can use workers
//	aggregate := search.NewAggregatedSearch()
//	for idx, pIndexer := range ag.Indexes {
//		// Run the Indexes in a goroutine
//		i, pIndex := idx, pIndexer
//		if len(pIndexer.Errors()) > 0 {
//			log.WithFields(log.Fields{"index": pIndexer}).Debug("Skipping index because it has errors")
//			continue
//		}
//		errorGroup.Go(func() error {
//			iter := aggregate.SearchIterators[&pIndex]
//			fields, page := iter.Next()
//			results, err := pIndex.Search(query, createWorkerJob(pIndex, fields, page))
//			if err != nil {
//				log.Warnf("Indexes %q failed: %s", pIndex.Info().GetID(), err)
//				return nil
//			}
//			iter.ScanForStaleResults(results)
//			allResults[i] = results
//			if l := len(results); l > maxLength {
//				maxLength = l
//			}
//			return nil
//		})
//	}
//	if err := errorGroup.Wait(); err != nil {
//		log.Warn(err)
//		return nil, err
//	}
//	var results []search.ResultItemBase
//
//	// interleave search results to preserve ordering
//	for i := 0; i <= maxLength; i++ {
//		for _, r := range allResults {
//			if len(r) > i {
//				results = append(results, r[i])
//			}
//		}
//	}
//
//	if query.Limit > 0 && len(results) > query.Limit {
//		results = results[:query.Limit]
//	}
//
//	return results, nil
//}

func (ag *Aggregate) Capabilities() torznab.Capabilities {
	return torznab.Capabilities{
		SearchModes: []search.Capability{
			{Key: "movie-search", Available: true, SupportedParams: []string{"q", "imdbid"}},
			{Key: "tv-search", Available: true, SupportedParams: []string{"q", "season", "ep"}},
			{Key: "search", Available: true, SupportedParams: []string{"q"}},
		},
	}
}

func (ag *Aggregate) Download(string) (*ResponseProxy, error) {
	return nil, errors.New("not implemented")
}

