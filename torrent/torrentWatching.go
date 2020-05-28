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
func Watch(client *indexer.IndexerHelper, interval int) {
	//Fetch pages untill we don't see any new torrents
	startingPage := uint(0)
	maxPages := uint(10)
	page := uint(0)
	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)
	ops := client.GetDefaultOptions()
	ops.StopOnStaleTorrents = true
	err := GetNewTorrents(client, ops)
	if err != nil {
		fmt.Println("Could not fetch initial torrents")
		os.Exit(1)
	}
	//client.clearSearch()
	var currentSearch search.Instance
	for true {
		var err error
		if currentSearch == nil {
			currentSearch, err = client.SearchKeywords(nil, "", page)
		} else {
			currentSearch, err = client.SearchKeywords(currentSearch, "", page)
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
			//torrentNumber := page*client.pageSize + counter + 1
			//isNew, isUpdate := HandleTorrentDiscovery(torrent)
			if torrent.IsNew() || torrent.IsUpdate() {
				if torrent.IsNew() && !torrent.IsUpdate() {
					_, _ = fmt.Fprintf(tabWr, "Found new torrent #%s:\t%s\t[%s]:\t%s\n",
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
