package indexer

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer/search"
	"strconv"
	"strings"
	"time"
)

//Extracts a field's value from the given selection
func (r *Runner) extractField(selection *goquery.Selection, field *fieldBlock) (interface{}, error) {
	if field == nil {
		return "", errors.New("no field given")
	}
	r.logger.
		WithFields(logrus.Fields{"block": field.Block.String()}).
		Debugf("Processing field %q", field.Field)
	val, err := field.Block.Match(selection)
	return val, err
}

//formatValues formats a field's value (singular or multiple)
func formatValues(field *fieldBlock, value interface{}, values map[string]interface{}) interface{} {
	if value == nil && field == nil {
		return nil
	} else if value == nil && field.Block.TextVal == "" {
		return value
	} else if value == nil && field.Block.TextVal != "" {
		value = field.Block.TextVal
	}

	if _, ok := value.([]string); ok {
		valueArray := value.([]string)
		for ix, subValue := range valueArray {
			newValue := formatValues(field, subValue, values).(string)
			valueArray[ix] = newValue
		}
		return value
	}
	if value == nil {
		value = ""
	}
	strValue := value.(string)
	if strings.Contains(strValue, "{{") || (field != nil && field.Block.Pattern != "") {
		templateData := values
		updated, err := applyTemplate("result_template", strValue, templateData)
		if err != nil {
			return strValue
		}
		strValue = updated
	} else {
		//Don't format non-patterns
		return strValue
	}
	if field != nil {
		updated, err := field.Block.ApplyFilters(strValue)
		if err != nil || updated == "" {
			return strValue
		}
		strValue = updated
	}
	return strValue
}

//Extract the actual result item from it's row/col
//TODO: refactor this to reduce #complexity
func (r *Runner) extractItem(rowIdx int, selection *goquery.Selection, context *rowContext) (search.ResultItemBase, error) {
	row := map[string]interface{}{}
	nonFilteredRow := map[string]string{}
	//html, _ := goquery.OuterHtml(selection)
	r.logger.WithFields(logrus.Fields{}).Debug("Processing row")

	for _, item := range r.definition.Search.Fields {
		r.logger.
			WithFields(logrus.Fields{"row": rowIdx, "block": item.Block.String()}).
			Debugf("Processing field %q", item.Field)
		val, err := item.Block.Match(selection)
		if item.Field == "title" {
			valRaw, err := item.Block.MatchRawText(selection)
			if err == nil {
				nonFilteredRow[item.Field] = valRaw
			}
		}

		if err != nil {
			r.logger.WithFields(logrus.Fields{"error": err, "selector": item.Field}).
				Debugf("Couldn't process selector")
			r.failingSearchFields[item.Field] = item
			continue
		}

		row[item.Field] = val
	}

	//Evaluate pattern items
	for _, item := range r.definition.Search.Fields {
		value := row[item.Field]
		currentItem := item
		value = formatValues(&currentItem, value, row)
		row[item.Field] = value
	}

	item := r.definition.getNewResultItem()
	item.SetSite(r.definition.Site)
	item.SetIndexer(r.getIndexer())

	r.logger.WithFields(logrus.Fields{"row": rowIdx, "data": row}).
		Debugf("Finished row %d", rowIdx)

	//Fill in the extracted values
	for key, val := range row {
		formatValues(nil, val, row)

		if r.definition.Scheme == "torrent" {
			r.populateTorrentItemField(item, key, val, row, nonFilteredRow, rowIdx)
		} else {
			r.populateScrapeItemField(item, key, val, row, rowIdx)
		}
	}

	if r.hasDateHeader() {
		date, err := r.extractDateHeader(selection)
		if err != nil {
			return &search.ScrapeResultItem{}, err
		}

		item.SetPublishDate(date.Unix())
	}

	if r.definition.Scheme == "torrent" {
		r.populateTorrentData(item, context)
	}

	return item, nil
}

func (r *Runner) populateTorrentData(resultItem search.ResultItemBase, context *rowContext) {
	//Maybe don't do that always?
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
		u, err := r.resolveIndexerPath(firstString(val))
		if err != nil {
			r.logger.Warnf("Row #%d has unparseable url %q in %s", rowIdx, val, key)
			return false
		}
		//item.UUIDValue = u
		item.Description = u
		// comments is used by Sonarr for linking to
		if item.Comments == "" {
			item.Comments = u
		}
	case "comments":
		u, err := r.resolveIndexerPath(firstString(val))
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
		murl, err := r.resolveIndexerPath(firstString(val))
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
		//r.logger.Debugf("After parsing, size is %v", bytes)
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
		item.AuthorId = firstString(val)
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
		banner, err := r.resolveIndexerPath(firstString(val))
		if err != nil {
			item.Banner = firstString(val)
		} else {
			item.Banner = banner
		}
	default:
		populatedOk := r.populateScrapeItemField(&item.ScrapeResultItem, key, val, row, rowIdx)
		if !populatedOk {
			return false
		}
	}
	return true
}

func (r *Runner) populateScrapeItemField(item search.ResultItemBase, key string, val interface{}, row map[string]interface{}, rowIdx int) bool {
	scrapeItem := item.(*search.ScrapeResultItem)
	switch key {
	case "id":
		scrapeItem.SetLocalId(firstString(val))
	case "download":
		u, err := r.resolveIndexerPath(firstString(val))
		if err != nil {
			r.logger.Warnf("Row #%d has unparseable url %q in %s", rowIdx, val, key)
			return false
		}
		//item.Link = u
		scrapeItem.SourceLink = u
	case "link":
		u, err := r.resolveIndexerPath(firstString(val))
		if err != nil {
			r.logger.Warnf("Row #%d has unparseable url %q in %s", rowIdx, val, key)
			return false
		}
		scrapeItem.Link = u
	default:
		scrapeItem.SetField(key, val)
	}
	return true
}
