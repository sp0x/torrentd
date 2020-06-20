package torrent

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/indexer/search"
	"os"
	"text/tabwriter"
	"time"
)

//Watch tracks a tracker for any new torrents and records them.
func Watch(facade *indexer.Facade, interval int) {
	//Fetch pages untill we don't see any new torrents
	startingPage := uint(0)
	maxPages := facade.Indexer.MaxSearchPages()
	page := uint(0)
	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)
	ops := facade.GetDefaultOptions()
	ops.StopOnStaleResults = true
	err := GetNewTorrents(facade, ops)
	if err != nil {
		fmt.Println("Could not fetch initial torrents")
		os.Exit(1)
	}
	//facade.clearSearch()
	var currentSearch search.Instance
	for {
		var err error
		if currentSearch == nil {
			currentSearch, err = facade.SearchKeywords(nil, "", page)
		} else {
			currentSearch, err = facade.SearchKeywords(currentSearch, "", page)
		}
		if err != nil {
			time.Sleep(time.Second * time.Duration(interval))
			switch err.(type) {
			case *indexer.LoginError:
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
		for _, torrent := range currentSearch.GetResults() {
			if finished {
				break
			}
			//torrentNumber := page*facade.pageSize + counter + 1
			//isNew, isUpdate := HandleTorrentDiscovery(torrent)
			if torrent.IsNew() || torrent.IsUpdate() {
				if torrent.IsNew() && !torrent.IsUpdate() {
					_, _ = fmt.Fprintf(tabWr, "Found new result #%s:\t%s\t[%s]:\t%s\n",
						torrent.LocalId, torrent.AddedOnStr(), torrent.Fingerprint, torrent.Title)
				} else {
					_, _ = fmt.Fprintf(tabWr, "Updated torrent #%s:\t%s\t[%s]:\t%s\n",
						torrent.LocalId, torrent.AddedOnStr(), torrent.Fingerprint, torrent.Title)
				}
			}
			_ = tabWr.Flush()
			if !torrent.IsNew() {
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

}
