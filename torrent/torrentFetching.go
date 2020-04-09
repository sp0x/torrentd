package torrent

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/db"
	"os"
	"text/tabwriter"
)

func GetNewTorrents(client *Rutracker, fetchOptions *FetchOptions) error {
	log.Info("Searching for new torrents")
	if fetchOptions == nil {
		fetchOptions = client.GetDefaultOptions()
	}
	totalTorrents := fetchOptions.PageCount * client.pageSize
	page := uint(0)
	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)

	var currentSearch *Search
	for page = 0; page < fetchOptions.PageCount; page++ {
		log.Infof("Getting page %d\n", page)
		var err error
		if currentSearch == nil {
			currentSearch, err = client.Search(nil, page)
		} else {
			currentSearch, err = client.Search(currentSearch, page)
		}

		if err != nil {
			log.Warningf("Could not fetch page %d\n", page)
			continue
		}
		/*
			Scan all pages every time. It's not safe to skip them by last torrent ID in the database,
			because some of them might be hidden at the previous run.
		*/
		counter := uint(0)
		finished := false
		client.parseTorrents(currentSearch.doc, func(i int, torrent *db.Torrent) {
			if finished || torrent == nil {
				return
			}
			torrentNumber := page*client.pageSize + counter + 1
			existingTorrent := client.storage.FindByTorrentId(torrent.TorrentId)
			isNew := existingTorrent == nil || existingTorrent.AddedOn != torrent.AddedOn
			isUpdate := existingTorrent != nil && (existingTorrent.AddedOn != torrent.AddedOn)

			if !isNew && fetchOptions.StopOnStaleTorrents {
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
					client.storage.UpdateTorrent(existingTorrent.ID, torrent)
					_, _ = fmt.Fprintf(tabWr, "Updated torrent #%s:\t%s\t[%s]:\t%s\n",
						torrent.TorrentId, torrent.AddedOn, torrent.Fingerprint, torrent.Name)
				} else {
					torrent.Fingerprint = getTorrentFingerprint(torrent)
					_, _ = fmt.Fprintf(tabWr, "Found new torrent #%s:\t%s\t[%s]:\t%s\n",
						torrent.TorrentId, torrent.AddedOnStr(), torrent.Fingerprint, torrent.Name)
					client.storage.Create(torrent)
				}
			} else {
				_, _ = fmt.Fprintf(tabWr, "Torrent #%s:\t%s\t[%s]:\t%s\n",
					torrent.TorrentId, torrent.AddedOn, "#", torrent.Name)

			}
			_ = tabWr.Flush()
			counter++
		})
		if finished {
			break
		}
		if counter != client.pageSize {
			log.Errorf("Error while parsing page %d: got %d torrents instead of %d\n", page, counter, client.pageSize)
		}
	}
	return nil
}
