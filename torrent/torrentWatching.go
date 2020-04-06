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
	keepupPagesCount := uint(10)
	startingPage := uint(0)
	maxPages := uint(10)
	totalTorrents := keepupPagesCount * client.pageSize

	page := uint(0)
	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)
	ops := client.getDefaultOptions()
	ops.StopOnStaleTorrents = true
	err := GetNewTorrents(client, ops)
	if err != nil {
		fmt.Println("Could not fetch initial torrents")
		os.Exit(1)
	}
	client.clearSearch()
	for true {
		log.Infof("Reloading tracker on page: %d\n", page)
		pageDoc, err := client.search(page)
		if err != nil {
			time.Sleep(time.Second * time.Duration(interval))
			continue
		}
		if pageDoc == nil {
			log.Warningf("Could not fetch torrent page: %d\n", page)
			time.Sleep(time.Second * time.Duration(interval))
			continue
		}
		//Parse the page and see if there are any new torrents
		//if there aren't any, sleep the interval
		counter := uint(0)
		finished := false
		hasStaleTorrents := false
		client.parseTorrents(pageDoc, func(i int, torrent *db.Torrent) {
			if finished || torrent == nil {
				return
			}
			torrentNumber := page*client.pageSize + counter + 1
			existingTorrent := client.torrentStorage.FindByTorrentId(torrent.TorrentId)
			isNew := existingTorrent == nil || existingTorrent.AddedOn != torrent.AddedOn
			isUpdate := existingTorrent != nil && (existingTorrent.AddedOn != torrent.AddedOn)
			if !isNew {
				hasStaleTorrents = true
				finished = true
				return
			}
			if isNew && torrentNumber >= totalTorrents/2 {
				log.Warningf("Got a new torrent after a half of the search (%d of %d). "+
					"Consider to increase the search page number.\n", torrentNumber, totalTorrents)
			}
			if isNew || (existingTorrent != nil && existingTorrent.Name != torrent.Name) {
				if isUpdate {
					torrent.Fingerprint = existingTorrent.Fingerprint
					_, _ = fmt.Fprintf(tabWr, "Found new torrent #%s:\t%s\t[%s]:\t%s\n",
						torrent.TorrentId, torrent.AddedOn, torrent.Fingerprint, torrent.Name)
					client.torrentStorage.UpdateTorrent(existingTorrent.ID, torrent)
					_, _ = fmt.Fprintf(tabWr, "Updated torrent #%s:\t%s\t[%s]:\t%s\n",
						torrent.TorrentId, torrent.AddedOn, torrent.Fingerprint, torrent.Name)
				} else {
					torrent.Fingerprint = getTorrentFingerprint(torrent)
					_, _ = fmt.Fprintf(tabWr, "Found new torrent #%s:\t%s\t[%s]:\t%s\n",
						torrent.TorrentId, torrent.AddedOn, torrent.Fingerprint, torrent.Name)
					client.torrentStorage.Create(torrent)
				}
			} else {
				_, _ = fmt.Fprintf(tabWr, "Torrent #%s:\t%s\t[%s]:\t%s\n",
					torrent.TorrentId, torrent.AddedOn, "#", torrent.Name)

			}
			_ = tabWr.Flush()
			counter++
		})
		//If we have stale torrents we wait some time and try again
		if hasStaleTorrents {
			time.Sleep(time.Second * time.Duration(interval))
			client.clearSearch()
			page = startingPage
			continue
		}
		//Otherwise we proceed to the next page if there's any
		page += 1
		//We've exceeded the pages, go to the start
		if maxPages == page {
			page = startingPage
			client.clearSearch()
		}
	}

}
