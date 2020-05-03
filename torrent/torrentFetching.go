package torrent

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/indexer"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"os"
	"text/tabwriter"
)

//
func GetNewTorrents(client *indexer.IndexerHelper, fetchOptions *indexer.GenericSearchOptions) error {
	log.Info("Searching for new torrents")
	if fetchOptions == nil {
		fetchOptions = client.GetDefaultOptions()
	}

	page := uint(0)
	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)

	var currentSearch search.Instance
	for page = 0; page < fetchOptions.PageCount; page++ {
		log.Infof("Getting page %d\n", page)
		var err error
		if currentSearch == nil {
			currentSearch, err = client.SearchKeywords(nil, "", page)
		} else {
			currentSearch, err = client.SearchKeywords(currentSearch, "", page)
		}
		if err != nil {
			log.Warningf("Could not fetch page %d\n", page)
			switch err.(type) {
			case *indexer.LoginError:
				return err
			}
		}
		/*
			Scan all pages every time. It's not safe to skip them by last torrent ID in the database,
			because some of them might be hidden at the previous run.
		*/
		counter := uint(0)
		finished := false
		for _, torrent := range currentSearch.GetResults() {
			if finished {
				break
			}
			//isNew, isUpdate := HandleTorrentDiscovery(torrent)
			if torrent.IsNew() || torrent.IsUpdate() {
				if torrent.IsNew() && !torrent.IsUpdate() {
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
			_ = tabWr.Flush()
			if !torrent.IsNew() && fetchOptions.StopOnStaleTorrents {
				finished = true
				break
			}
			counter++
		}
		if finished {
			break
		}
		//if counter != client.pageSize {
		//	log.Errorf("No results while parsing page %d: got %d torrents instead of %d\n", page, counter, client.pageSize)
		//}
	}
	return nil
}
