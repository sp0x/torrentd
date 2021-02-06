package indexer

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"

	"github.com/sp0x/torrentd/indexer/search"
)

func Get(facade *Facade, query *search.Query) error {
	if query == nil {
		query = search.NewQuery()
	}
	page := uint(0)
	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)

	pagesToFetch := query.PageCount
	if pagesToFetch == 0 {
		pagesToFetch = 10
	}
	searchInstance := search.NewSearch(query)
	if searchInstance == nil {
		return fmt.Errorf("couldn't search for page %d", page)
	}

	for page = 0; page < pagesToFetch; page++ {
		log.Infof("Fetching page %d", page)
		var err error
		searchInstance, err = facade.Search(searchInstance, query)
		if searchInstance == nil {
			return fmt.Errorf("couldn't search for page %d", page)
		}
		if err != nil {
			log.Warningf("Could not fetch page %d", page)
			switch err.(type) {
			case *LoginError:
				return err
			}
		}
		/*
			Scan all pages every time. It's not safe to skip them by last scrapeItem ID in the database,
			because some of them might be hidden at the previous run.
		*/
		finished := false
		for _, scrapeItem := range searchInstance.GetResults() {
			if finished {
				break
			}

			logFetchedResult(scrapeItem, tabWr)
			if !scrapeItem.IsNew() {
				finished = true
				break
			}
		}
		if finished {
			break
		}
	}
	return nil
}

func logFetchedResult(scrapeItem search.ResultItemBase, tabWr *tabwriter.Writer) {
	if scrapeItem.IsNew() || scrapeItem.IsUpdate() {
		if scrapeItem.IsNew() && !scrapeItem.IsUpdate() {
			serialized, _ := json.Marshal(scrapeItem)
			_, _ = fmt.Fprintf(tabWr, "Found new result #%s:\t%s\n",
				scrapeItem.UUID(), string(serialized))
		} else {
			_, _ = fmt.Fprintf(tabWr, "Updated result #%s:\t%s\n",
				scrapeItem.UUID(), scrapeItem.String())
		}
	} else {
		_, _ = fmt.Fprintf(tabWr, "Result #%s:\t%s\n",
			scrapeItem.UUID(), scrapeItem.String())
	}
	_ = tabWr.Flush()
}
