package indexer

import (
	"github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/torznab"
	"strconv"
)

func (r *Runner) resolveCategory(item *extractedItem) {
	if mappedCat, ok := r.definition.Capabilities.CategoryMap[item.LocalCategoryID]; ok {
		item.Category = mappedCat.ID
	} else {
		r.logger.
			WithFields(logrus.Fields{"localId": item.LocalCategoryID, "localName": item.LocalCategoryName}).
			Warn("Unknown local category")

		if intCatId, err := strconv.Atoi(item.LocalCategoryID); err == nil {
			item.Category = intCatId + torznab.CustomCategoryOffset
		}
	}
}
