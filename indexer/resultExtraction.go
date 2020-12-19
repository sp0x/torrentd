package indexer

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer/search"
	"strings"
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
func (r *Runner) extractItem(rowIdx int, selection RawScrapeItem, context *rowContext) (search.ResultItemBase, error) {
	row := map[string]interface{}{}
	nonFilteredRow := map[string]string{}
	r.logger.WithFields(logrus.Fields{}).
		Debug("Processing row")

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

	//Fill in the extracted values
	for key, val := range row {
		formatValues(nil, val, row)

		if r.definition.Scheme == "torrent" {
			r.populateTorrentItemField(item, key, val, row, nonFilteredRow, rowIdx)
		} else {
			r.populateScrapeItemField(item, key, val, rowIdx)
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

	r.logger.WithFields(logrus.Fields{"row": rowIdx, "data": row}).
		Debugf("Finished row %d", rowIdx)

	return item, nil
}

func (r *Runner) populateScrapeItemField(item search.ResultItemBase, key string, val interface{}, rowIdx int) bool {
	scrapeItem := item.(*search.ScrapeResultItem)
	switch key {
	case "id":
		scrapeItem.SetLocalId(firstString(val))
	case "download":
		u, err := r.resolvePathInIndex(firstString(val))
		if err != nil {
			r.logger.Warnf("Row #%d has unparseable url %q in %s", rowIdx, val, key)
			return false
		}
		//item.Link = u
		scrapeItem.SourceLink = u
	case "link":
		u, err := r.resolvePathInIndex(firstString(val))
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
