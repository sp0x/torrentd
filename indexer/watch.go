package indexer

import (
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/torznab"
	"time"
)

//Watch tracks an index for any new items and records them.
//The interval is in seconds
func Watch(helper *Facade, initialQuery torznab.Query, intervalSec int) <-chan search.ExternalResultItem {
	outputChan := make(chan search.ExternalResultItem)
	go func() {
		var currentSearch search.Instance
		startingPage := uint(0)
		maxPages := helper.Indexer.MaxSearchPages()
		currentPage := uint(0)
		for {
			var err error
			if currentSearch == nil {
				currentSearch, err = helper.Search(nil, initialQuery)
			} else {
				currentSearch, err = helper.Search(currentSearch, initialQuery)
			}
			if err != nil {
				time.Sleep(time.Second * time.Duration(intervalSec))
				switch err.(type) {
				case *LoginError:
					return
				}
			}
			if currentSearch == nil {
				log.Warningf("Could not fetch torrent currentPage: %d\n", currentPage)
				time.Sleep(time.Second * time.Duration(intervalSec))
				continue
			}
			//Parse the currentPage and see if there are any new torrents
			//if there aren't any, sleep the intervalSec
			counter := uint(0)
			finished := false
			hasReachedStaleItems := false
			for _, result := range currentSearch.GetResults() {
				outputChan <- result
			}
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
				counter++
			}
			//If we have stale torrents we wait some time and try again
			if hasReachedStaleItems {
				time.Sleep(time.Second * time.Duration(intervalSec))
				currentSearch = nil
				currentPage = startingPage
				continue
			}
			//Otherwise we proceed to the next currentPage if there's any
			currentPage += 1
			//We've exceeded the pages, sleep and go to the start
			if maxPages == currentPage {
				currentPage = startingPage
				currentSearch = nil
				time.Sleep(time.Second * time.Duration(intervalSec))
			}
		}
	}()
	return outputChan
}
