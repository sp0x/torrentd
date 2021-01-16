package indexer

import (
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/search"
)

func (r *Runner) populateCategory(item interface{}) {
	switch item.(type) {
	case *search.TorrentResultItem:
		torrentItem := item.(*search.TorrentResultItem)
		r.resolveCategoryForTorrent(torrentItem)
		break
	}
}

func (r *Runner) resolveCategoryForTorrent(torrentItem *search.TorrentResultItem) {
	if mappedCat, ok := r.definition.Capabilities.CategoryMap[torrentItem.LocalCategoryID]; ok {
		torrentItem.Category = mappedCat.ID
	} else {
		r.logger.
			WithFields(logrus.Fields{"localId": torrentItem.LocalCategoryID, "localName": torrentItem.LocalCategoryName}).
			Debug("Unknown local category")

		if intCatID, err := strconv.Atoi(torrentItem.LocalCategoryID); err == nil {
			torrentItem.Category = intCatID + categories.CustomCategoryOffset
		}
	}
}
