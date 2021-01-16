package indexer

import (
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
	var currentSearch search.Instance
	pagesToFetch := query.PageCount
	if pagesToFetch == 0 {
		pagesToFetch = 10
	}
	for page = 0; page < pagesToFetch; page++ {
		log.Infof("Getting page %d", page)
		var err error
		if currentSearch == nil {
			currentSearch, err = facade.Search(nil, query)
		} else {
			currentSearch, err = facade.Search(currentSearch, query)
		}
		if currentSearch == nil {
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
		counter := uint(0)
		finished := false
		for _, scrapeItem := range currentSearch.GetResults() {
			if finished {
				break
			}

			if scrapeItem.IsNew() || scrapeItem.IsUpdate() {
				if scrapeItem.IsNew() && !scrapeItem.IsUpdate() {
					_, _ = fmt.Fprintf(tabWr, "Found new result #%s:\t%s\n",
						scrapeItem.UUID(), scrapeItem.String())
				} else {
					_, _ = fmt.Fprintf(tabWr, "Updated result #%s:\t%s\n",
						scrapeItem.UUID(), scrapeItem.String())
				}
			} else {
				_, _ = fmt.Fprintf(tabWr, "Result #%s:\t%s\n",
					scrapeItem.UUID(), scrapeItem.String())
			}
			_ = tabWr.Flush()
			if !scrapeItem.IsNew() {
				finished = true
				break
			}
			counter++
		}
		if finished {
			break
		}
		//if counter != facade.pageSize {
		//	log.Errorf("No results while parsing page %d: got %d torrents instead of %d\n", page, counter, facade.pageSize)
		//}
	}
	return nil
}
