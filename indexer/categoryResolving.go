package indexer

import (
	"github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/search"
	"strconv"
)

func (r *Runner) resolveCategory(item *search.ExternalResultItem) {
	if mappedCat, ok := r.definition.Capabilities.CategoryMap[item.LocalCategoryID]; ok {
		item.Category = mappedCat.ID
	} else {
		r.logger.
			WithFields(logrus.Fields{"localId": item.LocalCategoryID, "localName": item.LocalCategoryName}).
			Debug("Unknown local category")

		if intCatId, err := strconv.Atoi(item.LocalCategoryID); err == nil {
			item.Category = intCatId + categories.CustomCategoryOffset
		}
	}
}
