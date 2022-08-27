package indexer

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/sp0x/torrentd/indexer/search"
)

// IteratePages goes over all the pages in an index and returns the results through a channel.
func GetAllPagesFromIndex(facade *Facade, query *search.Query) <-chan search.ResultItemBase {
	outputChan := make(chan search.ResultItemBase)
	if query == nil {
		query = search.NewQuery()
	}
	maxPages := facade.Indexes.MaxSearchPages()
	query.NumberOfPagesToFetch = maxPages
	query.StopOnStale = true
	resultsChan, _ := facade.Search(query)
	go func() {
		for items := range resultsChan {
			for _, item := range items {
				outputChan <- item
			}
		}
	}()
	return outputChan
}

// Watch tracks an index for any new items, through all search pages(or max pages).
// Whenever old results are found, or we've exhausted the number of pages, the search restarts from the start.
// The interval is in seconds, it's used to sleep after each search for new results.
func Watch(facade *Facade, initialQuery *search.Query, intervalSec int) <-chan search.ResultItemBase {
	outputChan := make(chan search.ResultItemBase)
	if initialQuery == nil {
		initialQuery = search.NewQuery()
	}
	startingPage := initialQuery.Page
	initialQuery.NumberOfPagesToFetch = facade.Indexes.MaxSearchPages()
	initialQuery.StopOnStale = true
	go func() {
		for {
			results, err := facade.Search(initialQuery)
			if err != nil {
				switch err.(type) {
				case *LoginError:
					log.Warnf("search got login error: %v", err)
					return
				default:
					log.Warnf("search error: %v", err)
				}
				time.Sleep(time.Second * time.Duration(intervalSec))
			}
			for resultsPage := range results {
				sendSearchResults(resultsPage, outputChan)
			}

			// If we have stale items we wait some time and try again
			log.WithFields(log.Fields{"page": initialQuery.NumberOfPagesToFetch}).
				Infof("Search is complete.")
			time.Sleep(time.Second * time.Duration(intervalSec))
			initialQuery.Page = startingPage
		}
	}()
	return outputChan
}

func sendSearchResults(results []search.ResultItemBase, outputChan chan search.ResultItemBase) {
	for _, result := range results {
		tmpResult := result
		if result.IsNew() || result.IsUpdate() {
			scrapeItem := result.AsScrapeItem()
			if result.IsNew() && !result.IsUpdate() {
				log.WithFields(log.Fields{"id": result.UUID(), "data": result.String(), "published": scrapeItem.PublishDate}).
					Info("Found new result")
			} else {
				log.WithFields(log.Fields{"id": result.UUID(), "data": result.String(), "published": scrapeItem.PublishDate}).
					Info("Updated result")
			}
		}
		outputChan <- tmpResult
	}
}
