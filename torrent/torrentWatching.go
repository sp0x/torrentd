package torrent

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/db"
	"os"
	"text/tabwriter"
	"time"
)

//Watch tracks a tracker for any new torrents and records them.
func Watch(client *Rutracker, interval int) {
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
	var currentSearch *Search
	for true {
		var err error
		if currentSearch == nil {
			currentSearch, err = client.Search(nil, "", page)
		} else {
			currentSearch, err = client.Search(currentSearch, "", page)
		}
		if err != nil {
			time.Sleep(time.Second * time.Duration(interval))
			continue
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
		client.ParseTorrents(currentSearch.doc, func(i int, torrent *db.Torrent) {
			if finished || torrent == nil {
				return
			}
			//torrentNumber := page*client.pageSize + counter + 1
			isNew, isUpdate := HandleTorrentDiscovery(client, torrent)
			if isNew || isUpdate {
				if isNew && !isUpdate {
					_, _ = fmt.Fprintf(tabWr, "Found new torrent #%s:\t%s\t[%s]:\t%s\n",
						torrent.TorrentId, torrent.AddedOnStr(), torrent.Fingerprint, torrent.Name)
				} else {
					_, _ = fmt.Fprintf(tabWr, "Updated torrent #%s:\t%s\t[%s]:\t%s\n",
						torrent.TorrentId, torrent.AddedOnStr(), torrent.Fingerprint, torrent.Name)
				}
			}
			_ = tabWr.Flush()
			if !isNew {
				hasStaleTorrents = true
				finished = true
				return
			}
			counter++
		})
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
