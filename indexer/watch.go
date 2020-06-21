package indexer

import (
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/torznab"
	"time"
)

//IteratePages goes over all the pages in an index and returns the results through a channel.
func GetAllPagesFromIndex(facade *Facade, query *torznab.Query) <-chan search.ExternalResultItem {
	outputChan := make(chan search.ExternalResultItem)
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
				break
			}
			if currentSearch == nil {
				log.Warningf("Could not fetch page: %d\n", query.Page)
				break
			}
			for _, result := range currentSearch.GetResults() {
				outputChan <- result
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
func Watch(facade *Facade, initialQuery *torznab.Query, intervalSec int) <-chan search.ExternalResultItem {
	outputChan := make(chan search.ExternalResultItem)
	if initialQuery == nil {
		initialQuery = &torznab.Query{}
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
				time.Sleep(time.Second * time.Duration(intervalSec))
				switch err.(type) {
				case *LoginError:
					return
				}
			}
			if currentSearch == nil {
				log.Warningf("Could not fetch page: %d\n", initialQuery.Page)
				time.Sleep(time.Second * time.Duration(intervalSec))
				continue
			}
			for _, result := range currentSearch.GetResults() {
				outputChan <- result
			}
			//Parse the currentPage and see if there are any new torrents
			//if there aren't any, sleep the intervalSec
			finished := false
			hasReachedStaleItems := false
			for _, result := range currentSearch.GetResults() {
				if finished {
					break
				}
				if result.IsNew() || result.IsUpdate() {
					if result.IsNew() && !result.IsUpdate() {
						log.WithFields(log.Fields{"id": result.LocalId, "name": result.Title, "pub": result.PublishDate}).
							Info("Found new result")
					} else {
						log.WithFields(log.Fields{"id": result.LocalId, "name": result.Title, "pub": result.PublishDate}).
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
			if hasReachedStaleItems {
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
		}
	}()
	return outputChan
}
