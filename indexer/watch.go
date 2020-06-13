package indexer

import (
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/torznab"
	"time"
)

//Watch tracks a tracker for any new torrents and records them.
//The interval is in seconds
func Watch(helper *Facade, initialQuery torznab.Query, interval int) <-chan search.ExternalResultItem {
	//Fetch pages until we don't see any new torrents

	outputChan := make(chan search.ExternalResultItem)
	go func() {
		var currentSearch search.Instance
		startingPage := uint(0)
		maxPages := uint(10)
		page := uint(0)
		for true {
			var err error
			if currentSearch == nil {
				currentSearch, err = helper.Search(nil, initialQuery)
			} else {
				currentSearch, err = helper.Search(currentSearch, initialQuery)
			}
			if err != nil {
				time.Sleep(time.Second * time.Duration(interval))
				switch err.(type) {
				case *LoginError:
					return
				}
			}
			if currentSearch == nil {
				log.Warningf("Could not fetch torrent page: %d\n", page)
				time.Sleep(time.Second * time.Duration(interval))
				continue
			}
			//Parse the page and see if there are any new torrents
			//if there aren't any, sleep the interval
			counter := uint(0)
			finished := false
			hasStaleTorrents := false
			for _, result := range currentSearch.GetResults() {
				outputChan <- result
			}
			for _, result := range currentSearch.GetResults() {
				if finished {
					break
				}
				//isNew, isUpdate := HandleTorrentDiscovery(torrent)
				if result.IsNew() || result.IsUpdate() {
					if result.IsNew() && !result.IsUpdate() {
						log.WithFields(log.Fields{"id": result.LocalId, "name": result.Title, "pub": result.PublishDate}).
							Info("Found new torrent")
					} else {
						log.WithFields(log.Fields{"id": result.LocalId, "name": result.Title, "pub": result.PublishDate}).
							Info("Updated torrent")
					}

				}
				if !result.IsNew() {
					hasStaleTorrents = true
					finished = true
					break
				}
				counter++
			}
			//If we have stale torrents we wait some time and try again
			if hasStaleTorrents {
				time.Sleep(time.Second * time.Duration(interval))
				currentSearch = nil
				page = startingPage
				continue
			}
			//Otherwise we proceed to the next page if there's any
			page += 1
			//We've exceeded the pages, go to the start
			if maxPages == page {
				page = startingPage
				currentSearch = nil
			}
		}
	}()
	return outputChan
}
