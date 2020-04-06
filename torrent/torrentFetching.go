package torrent

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"os"
	"text/tabwriter"
)

func GetNewTorrents(client *Rutracker, fetchOptions *FetchOptions) error {
	log.Info("Searching for new torrents")
	if fetchOptions == nil {
		fetchOptions = client.getDefaultOptions()
	}
	totalTorrents := fetchOptions.PageCount * client.pageSize
	page := uint(0)
	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)

	for page = 0; page < fetchOptions.PageCount; page++ {
		log.Infof("Getting page %d\n", page)
		pageDoc, err := client.search(page)
		if err != nil {
			log.Warning("Could not fetch page %d", page)
			continue
		}
		/*
			Scan all pages every time. It's not safe to skip them by last torrent ID in the database,
			because some of them might be hidden at the previous run.
		*/
		counter := uint(0)
		pageDoc.Find("tr.tCenter.hl-tr").Each(func(i int, s *goquery.Selection) {
			torrentNumber := page*client.pageSize + counter + 1
			torrent := client.parseTorrentRow(s)
			if torrent == nil {
				return
			}
			existingTorrent := client.torrentStorage.FindByTorrentId(torrent.TorrentId)
			isNew := existingTorrent == nil || existingTorrent.AddedOn != torrent.AddedOn
			if isNew && torrentNumber >= totalTorrents/2 {
				log.Warningf("Got a new torrent after a half of the search (%d of %d). "+
					"Consider to increase the search page number.\n", torrentNumber, totalTorrents)
			}
			if isNew || (existingTorrent != nil && existingTorrent.Name != torrent.Name) {
				newTorrent := torrent
				newTorrent.Fingerprint = getTorrentFingerprint(torrent)
				_, _ = fmt.Fprintf(tabWr, "Found new torrent #%s:\t%s\t[%s]:\t%s\n",
					newTorrent.TorrentId, newTorrent.AddedOn, newTorrent.Fingerprint, newTorrent.Name)
				client.torrentStorage.Create(newTorrent)
			} else {
				_, _ = fmt.Fprintf(tabWr, "Torrent #%s:\t%s\t[%s]:\t%s\n",
					torrent.TorrentId, torrent.AddedOn, "#", torrent.Name)

			}
			_ = tabWr.Flush()
			counter++
		})
		if counter != client.pageSize {
			log.Errorf("Error while parsing page %s: got %s torrents instead of %s\n", page, counter, client.torrentStorage)
		}
	}
	return nil
}
