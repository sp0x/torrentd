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

	page := uint(0)
	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)

	var currentSearch *Search
	for page = 0; page < fetchOptions.PageCount; page++ {
		log.Infof("Getting page %d\n", page)
		var err error
		if currentSearch == nil {
			currentSearch, err = client.Search(nil, "", page)
		} else {
			currentSearch, err = client.Search(currentSearch, "", page)
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
		client.ParseTorrents(currentSearch.doc, func(i int, torrent *db.Torrent) {
			if finished || torrent == nil {
				return
			}
			isNew, isUpdate := HandleTorrentDiscovery(client, torrent)
			if isNew || isUpdate {
				if isNew {
					_, _ = fmt.Fprintf(tabWr, "Found new torrent #%s:\t%s\t[%s]:\t%s\n",
						torrent.TorrentId, torrent.AddedOnStr(), torrent.Fingerprint, torrent.Name)
				} else {
					_, _ = fmt.Fprintf(tabWr, "Updated torrent #%s:\t%s\t[%s]:\t%s\n",
						torrent.TorrentId, torrent.AddedOnStr(), torrent.Fingerprint, torrent.Name)
				}
			} else {
				_, _ = fmt.Fprintf(tabWr, "Torrent #%s:\t%s\t[%s]:\t%s\n",
					torrent.TorrentId, torrent.AddedOnStr(), "#", torrent.Name)
			}
			if !isNew && fetchOptions.StopOnStaleTorrents {
				finished = true
				return
			}
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

func HandleTorrentDiscovery(client *Rutracker, torrent *db.Torrent) (bool, bool) {
	existingTorrent := client.storage.FindByTorrentId(torrent.TorrentId)
	isNew := existingTorrent == nil || existingTorrent.AddedOn != torrent.AddedOn
	isUpdate := existingTorrent != nil && (existingTorrent.AddedOn != torrent.AddedOn)
	if isNew {
		if isUpdate && existingTorrent != nil {
			torrent.Fingerprint = existingTorrent.Fingerprint
			client.storage.UpdateTorrent(existingTorrent.ID, torrent)
		} else {
			torrent.Fingerprint = getTorrentFingerprint(torrent)
			client.storage.Create(torrent)
		}
	}

	return isNew, isUpdate
}
