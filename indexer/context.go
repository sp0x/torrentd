package indexer

import (
	"github.com/sp0x/torrentd/indexer/source"
)

// Read anything from the content that's needed
// so we can extract info about our run
func updateSearchDataFromScrapeItem(r *Runner, srch *workerJob, dom source.RawScrapeItem) {
	for _, item := range r.definition.Search.Context {
		val, err := item.Block.Match(dom)
		if err != nil {
			continue
		}
		if item.Field == "searchId" {
			srch.SetID(val.(string))
		}
	}
}
