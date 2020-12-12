package indexer

import (
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/torznab"
	"time"
)

//IteratePages goes over all the pages in an index and returns the results through a channel.
func GetAllPagesFromIndex(facade *Facade, query *torznab.Query) <-chan search.ResultItemBase { //nolint:unused
	outputChan := make(chan search.ResultItemBase)
	if query == nil {
		query = &torznab.Query{}
	}
	go func() {
		var currentSearch search.Instance
		maxPages := facade.Indexer.MaxSearchPages()
		for {
			var err error
			if currentSearch == nil {
				currentSearch, err = facade.Search(nil, query)
			} else {
				currentSearch, err = facade.Search(currentSearch, query)
			}
			if err != nil {
				log.Errorf("Couldn't search: %v", err)
				break
			}
			if currentSearch == nil {
				log.Warningf("Could not fetch page: %d\n", query.Page)
				break
			}
			for _, result := range currentSearch.GetResults() {
				tmpResult := result
				outputChan <- tmpResult
			}
			//Go to the next page
			query.Page += 1
			//If we've reached the end we stop
			if maxPages == query.Page {
				break
			}
		}
		close(outputChan)
	}()
	return outputChan
}

//Watch tracks an index for any new items, through all search pages(or max pages).
//Whenever old results are found, or we've exhausted the number of pages, the search restarts from the start.
//The interval is in seconds, it's used to sleep after each search for new results.
func Watch(facade *Facade, initialQuery *torznab.Query, intervalSec int) <-chan search.ResultItemBase {
	outputChan := make(chan search.ResultItemBase)
	if initialQuery == nil {
		initialQuery = torznab.NewQuery()
	}
	go func() {
		var currentSearch search.Instance
		startingPage := initialQuery.Page
		maxPages := facade.Indexer.MaxSearchPages()
		//Go over all pages
		for {
			var err error
			if currentSearch == nil {
				currentSearch, err = facade.Search(nil, initialQuery)
			} else {
				currentSearch, err = facade.Search(currentSearch, initialQuery)
			}
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
			if currentSearch == nil {
				log.Warningf("Could not fetch page: %d", initialQuery.Page)
				time.Sleep(time.Second * time.Duration(intervalSec))
				continue
			}
			sendSearchResults(currentSearch, outputChan)
			//Parse the currentPage and see if there are any new torrents
			//if there aren't any, sleep the intervalSec
			finished := false
			hasReachedStaleItems := false
			resultItems := currentSearch.GetResults()
			for _, result := range resultItems {
				if finished {
					break
				}
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
				if !result.IsNew() {
					hasReachedStaleItems = true
					finished = true
					break
				}
			}
			//If we have stale torrents we wait some time and try again
			if hasReachedStaleItems || len(resultItems) == 0 {
				log.WithFields(log.Fields{"page": initialQuery.PageCount}).
					Infof("Reached page with 0 results. Search is complete.")
				time.Sleep(time.Second * time.Duration(intervalSec))
				currentSearch = nil
				initialQuery.Page = startingPage
				continue
			}
			//Otherwise we proceed to the next currentPage if there's any
			initialQuery.Page += 1
			//We've exceeded the pages, sleep and go to the start
			if maxPages == initialQuery.Page {
				initialQuery.Page = startingPage
				currentSearch = nil
				time.Sleep(time.Second * time.Duration(intervalSec))
			}
			log.WithFields(log.Fields{"page": initialQuery.Page, "max-pages": maxPages}).
				Debugf("going through next page in query")
		}
	}()
	return outputChan
}

func sendSearchResults(currentSearch search.Instance, outputChan chan search.ResultItemBase) {
	for _, result := range currentSearch.GetResults() {
		tmpResult := result
		outputChan <- tmpResult
	}
}
