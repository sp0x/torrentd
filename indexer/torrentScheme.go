package indexer

import "github.com/sp0x/torrentd/indexer/search"

func itemMatchesScheme(scheme string, item search.ResultItemBase) bool {
	if scheme == "torrent" {
		_, ok := item.(*search.TorrentResultItem)
		return ok
	}
	_, ok := item.(*search.ScrapeResultItem)
	return ok
}

func (r *Runner) itemMatchesLocalCategories(localCats []string, item *search.TorrentResultItem) bool {
	for _, catId := range localCats {
		if catId == item.LocalCategoryID {
			return true
		}
	}
	return false
}
