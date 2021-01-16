package indexer

import (
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/sp0x/torrentd/indexer/search"
)

func (r *Runner) populateTorrentData(resultItem search.ResultItemBase, context *rowContext) {
	// Maybe don't do that always?
	item := resultItem.(*search.TorrentResultItem)

	item.Fingerprint = search.GetResultFingerprint(item)
	if !r.resolveItemCategory(context.query, context.indexCategories, item) {
		_ = context.storage.SetKey(r.getUniqueIndex(&item.ScrapeResultItem))
		err := context.storage.Add(item)
		if err != nil {
			r.logger.Errorf("Found an item that doesn't match our search indexCategories: %s\n", err)
		}
	}
}

func (r *Runner) populateTorrentItemField(
	itemToPopulate search.ResultItemBase,
	key string, val interface{},
	row map[string]interface{},
	nonFilteredRow map[string]string,
	rowIdx int) bool {
	item := itemToPopulate.(*search.TorrentResultItem)

	switch key {
	case "author":
		item.Author = firstString(val)
	case "details":
		u, err := r.getFullURLInIndex(firstString(val))
		if err != nil {
			r.logger.Warnf("Row #%d has unparseable url %q in %s", rowIdx, val, key)
			return false
		}
		// item.UUIDValue = u
		item.Description = u
		// comments is used by Sonarr for linking to
		if item.Comments == "" {
			item.Comments = u
		}
	case "comments":
		u, err := r.getFullURLInIndex(firstString(val))
		if err != nil {
			r.logger.Warnf("Row #%d has unparseable url %q in %s", rowIdx, val, key)
			return false
		}
		item.Comments = u
	case "title":
		item.Title = firstString(val)
		if _, ok := nonFilteredRow["title"]; ok {
			v := nonFilteredRow["title"]
			if strings.Contains(v, "{{") {
				v2, err := applyTemplate("original_title", v, row)
				if err == nil {
					v = v2
				}
			}
			item.OriginalTitle = v
		}
	case "shortTitle":
		item.ShortTitle = firstString(val)
	case "description":
		item.Description = firstString(val)
	case "category":
		item.LocalCategoryID = firstString(val)
	case "categoryName":
		item.LocalCategoryName = firstString(val)
	case "magnet":
		murl, err := r.getFullURLInIndex(firstString(val))
		if err != nil {
			r.logger.Warningf("Couldn't resolve magnet url from value %s\n", val)
			return false
		} else {
			item.MagnetLink = murl
		}
	case "size":
		bytes, err := humanize.ParseBytes(strings.Replace(firstString(val), ",", "", -1))
		if err != nil {
			r.logger.Warnf("Row #%d has unparseable size %q: %v", rowIdx, val, err.Error())
			return false
		}
		// r.logger.Debugf("After parsing, size is %v", bytes)
		item.Size = uint32(bytes)
	case "leechers":
		leechers, err := strconv.Atoi(normalizeNumber(firstString(val)))
		if err != nil {
			r.logger.Warnf("Row #%d has unparseable leechers value %q in %s", rowIdx, val, key)
			return false
		}
		item.Peers += leechers
	case "seeders":
		seeders, err := strconv.Atoi(normalizeNumber(firstString(val)))
		if err != nil {
			r.logger.Debugf("Row #%d has unparseable seeders value %q in %s", rowIdx, val, key)
			return false
		}
		item.Seeders = seeders
		item.Peers += seeders
	case "authorId":
		item.AuthorID = firstString(val)
	case "date":
		t, err := parseFuzzyTime(firstString(val), time.Now(), true)
		if err != nil {
			r.logger.Warnf("Row #%d has unparseable time %q in %s", rowIdx, val, key)
			return false
		}
		item.PublishDate = t.Unix()
	case "files":
		files, err := strconv.Atoi(normalizeNumber(firstString(val)))
		if err != nil {
			r.logger.Warnf("Row #%d has unparseable files value %q in %s", rowIdx, val, key)
			return false
		}
		item.Files = files
	case "grabs":
		grabs, err := strconv.Atoi(normalizeNumber(firstString(val)))
		if err != nil {
			r.logger.Warnf("Row #%d has unparseable grabs value %q in %s", rowIdx, val, key)
			return false
		}
		item.Grabs = grabs
	case "downloadvolumefactor":
		downloadvolumefactor, err := strconv.ParseFloat(normalizeNumber(firstString(val)), 64)
		if err != nil {
			r.logger.Warnf("Row #%d has unparseable downloadvolumefactor value %q in %s", rowIdx, val, key)
			return false
		}
		item.DownloadVolumeFactor = downloadvolumefactor
	case "uploadvolumefactor":
		uploadvolumefactor, err := strconv.ParseFloat(normalizeNumber(firstString(val)), 64)
		if err != nil {
			r.logger.Warnf("Row #%d has unparseable uploadvolumefactor value %q in %s", rowIdx, val, key)
			return false
		}
		item.UploadVolumeFactor = uploadvolumefactor
	case "minimumratio":
		minimumratio, err := strconv.ParseFloat(normalizeNumber(firstString(val)), 64)
		if err != nil {
			r.logger.Warnf("Row #%d has unparseable minimumratio value %q in %s", rowIdx, val, key)
			return false
		}
		item.MinimumRatio = minimumratio
	case "minimumseedtime":
		minimumseedtime, err := strconv.ParseFloat(normalizeNumber(firstString(val)), 64)
		if err != nil {
			r.logger.Warnf("Row #%d has unparseable minimumseedtime value %q in %s", rowIdx, val, key)
			return false
		}
		item.MinimumSeedTime = time.Duration(minimumseedtime) * time.Second
	case "banner":
		banner, err := r.getFullURLInIndex(firstString(val))
		if err != nil {
			item.Banner = firstString(val)
		} else {
			item.Banner = banner
		}
	default:
		populatedOk := r.populateScrapeItemField(&item.ScrapeResultItem, key, val, rowIdx)
		if !populatedOk {
			return false
		}
	}
	return true
}

func itemMatchesScheme(scheme string, item search.ResultItemBase) bool {
	if scheme == "torrent" {
		_, ok := item.(*search.TorrentResultItem)
		return ok
	}
	_, ok := item.(*search.ScrapeResultItem)
	return ok
}

func (r *Runner) itemMatchesLocalCategories(localCats []string, item *search.TorrentResultItem) bool {
	for _, catID := range localCats {
		if catID == item.LocalCategoryID {
			return true
		}
	}
	return false
}
