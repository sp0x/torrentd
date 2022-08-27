package torrent

import (
	"fmt"
	"os"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"

	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/indexer/search"
)

// GetNewScrapeItems gets the latest torrents.
func GetNewScrapeItems(facade *indexer.Facade, fetchOptions *indexer.GenericSearchOptions) error {
	log.Info("Searching for new torrents")
	if fetchOptions == nil {
		fetchOptions = facade.GetDefaultSearchOptions()
	}

	page := uint(0)
	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)
	query := search.NewQuery()
	query.StopOnStale = true

	resultsChannel, err := facade.Search(query)
	if err != nil {
		log.Warningf("Could not fetch page %d\n", page)
		if _, ok := err.(*indexer.LoginError); ok {
			return err
		}
	}
	/*
		Scan all pages every time. It's not safe to skip them by last scrapeItem ID in the database,
		because some of them might be hidden at the previous run.
	*/
	counter := uint(0)
	for resultsPage := range resultsChannel {
		for _, result := range resultsPage {
			if result.IsNew() || result.IsUpdate() {
				if result.IsNew() && !result.IsUpdate() {
					_, _ = fmt.Fprintf(tabWr, "Found new result #%s:\t%s\n", result.UUID(), result.String())
				} else {
					_, _ = fmt.Fprintf(tabWr, "Updated result #%s:\t%s\n", result.UUID(), result.String())
				}
			} else {
				_, _ = fmt.Fprintf(tabWr, "Result #%s:\t%s\n", result.UUID(), result.String())
			}
			_ = tabWr.Flush()
			counter++
		}
	}
	return nil
}
