package torrent

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"github.com/sp0x/rutracker-rss/torrent/storage"
	"os"
	"text/tabwriter"
)

//
func GetNewTorrents(client *Rutracker, fetchOptions *GenericSearchOptions) error {
	log.Info("Searching for new torrents")
	if fetchOptions == nil {
		fetchOptions = client.GetDefaultOptions()
	}

	page := uint(0)
	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)

	var currentSearch *search.Search
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
		client.ParseTorrents(currentSearch.DOM, func(i int, torrent *search.ExternalResultItem) {
			if finished || torrent == nil {
				return
			}
			isNew, isUpdate := HandleTorrentDiscovery(torrent)
			if isNew || isUpdate {
				if isNew && !isUpdate {
					_, _ = fmt.Fprintf(tabWr, "Found new torrent #%s:\t%s\t[%s]:\t%s\n",
						torrent.LocalId, torrent.AddedOnStr(), torrent.Fingerprint, torrent.Title)
				} else {
					_, _ = fmt.Fprintf(tabWr, "Updated torrent #%s:\t%s\t[%s]:\t%s\n",
						torrent.LocalId, torrent.AddedOnStr(), torrent.Fingerprint, torrent.Title)
				}
			} else {
				_, _ = fmt.Fprintf(tabWr, "Torrent #%s:\t%s\t[%s]:\t%s\n",
					torrent.LocalId, torrent.AddedOnStr(), "#", torrent.Title)
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
			log.Errorf("No results while parsing page %d: got %d torrents instead of %d\n", page, counter, client.pageSize)
		}
	}
	return nil
}

var defaultStorage = storage.Storage{}

//Handles torrent discovery
func HandleTorrentDiscovery(torrent *search.ExternalResultItem) (bool, bool) {
	var existingTorrent *search.ExternalResultItem
	if torrent.LocalId != "" {
		existingTorrent = defaultStorage.FindByTorrentId(torrent.LocalId)
	} else {
		existingTorrent = defaultStorage.FindNameAndIndexer(torrent.Title, torrent.Site)
	}

	isNew := existingTorrent == nil || existingTorrent.PublishDate != torrent.PublishDate
	isUpdate := existingTorrent != nil && (existingTorrent.PublishDate != torrent.PublishDate)
	if isNew {
		if isUpdate && existingTorrent != nil {
			torrent.Fingerprint = existingTorrent.Fingerprint
			defaultStorage.UpdateTorrent(existingTorrent.ID, torrent)
		} else {
			torrent.Fingerprint = search.GetTorrentFingerprint(torrent)
			defaultStorage.Create(torrent)
		}
	}
	return isNew, isUpdate
}
