package indexer

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"strconv"
	"strings"
	"time"
)

//Extracts a field's value from the given selection
func (r *Runner) extractField(selection *goquery.Selection, field *fieldBlock) (string, error) {
	if field == nil {
		return "", errors.New("no field given")
	}
	r.logger.
		WithFields(logrus.Fields{"block": field.Block.String()}).
		Debugf("Processing field %q", field.Field)
	val, err := field.Block.MatchText(selection)
	return val, err
}

//Extract the actual result item from it's row/col
func (r *Runner) extractItem(rowIdx int, selection *goquery.Selection) (search.ExternalResultItem, error) {
	row := map[string]string{}
	nonFilteredRow := map[string]string{}
	//html, _ := goquery.OuterHtml(selection)
	r.logger.WithFields(logrus.Fields{}).Debug("Processing row")

	for _, item := range r.definition.Search.Fields {
		r.logger.
			WithFields(logrus.Fields{"row": rowIdx, "block": item.Block.String()}).
			Debugf("Processing field %q", item.Field)
		val, err := item.Block.MatchText(selection)
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
		//r.logger.
		//	WithFields(logrus.Fields{"row": rowIdx, "output": val}).
		//	Debugf("Finished processing field %q", item.Field)

		row[item.Field] = val
	}

	//Evaluate pattern items
	for _, item := range r.definition.Search.Fields {
		val := row[item.Field]
		if item.Block.Pattern == "" && !strings.Contains(val, "{{") {
			continue
		}
		templateData := row
		updated, err := applyTemplate("result_template", val, templateData)
		if err != nil {
			return search.ExternalResultItem{}, err
		}
		updated, err = item.Block.ApplyFilters(updated)
		if err != nil {
			return search.ExternalResultItem{}, err
		}
		val = updated
		row[item.Field] = val
	}

	item := search.ExternalResultItem{
		ResultItem: search.ResultItem{
			Site:        r.definition.Site,
			Indexer:     r.getIndexer(),
			ExtraFields: make(map[string]string),
		},
	}

	r.logger.WithFields(logrus.Fields{"row": rowIdx, "data": row}).
		Debugf("Finished row %d", rowIdx)

	//Fill in the extracted values
	for key, val := range row {
		if strings.Contains(val, "{{") {
			//Value is a pattern, we must evaluate it for the last time
			updated, err := applyTemplate("result_template", val, row)
			if err == nil {
				val = updated
			}
		}
		switch key {
		case "id":
			item.LocalId = val
		case "author":
			item.Author = val
		case "download":
			u, err := r.resolveIndexerPath(val)
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable url %q in %s", rowIdx, val, key)
				continue
			}
			//item.Link = u
			item.SourceLink = u
		case "link":
			u, err := r.resolveIndexerPath(val)
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable url %q in %s", rowIdx, val, key)
				continue
			}
			item.Link = u
		case "details":
			u, err := r.resolveIndexerPath(val)
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable url %q in %s", rowIdx, val, key)
				continue
			}
			item.GUID = u
			// comments is used by Sonarr for linking to
			if item.Comments == "" {
				item.Comments = u
			}
		case "comments":
			u, err := r.resolveIndexerPath(val)
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable url %q in %s", rowIdx, val, key)
				continue
			}
			item.Comments = u
		case "title":
			item.Title = val
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
			item.ShortTitle = val
		case "description":
			item.Description = val
		case "category":
			item.LocalCategoryID = val
		case "categoryName":
			item.LocalCategoryName = val
		case "magnet":
			murl, err := r.resolveIndexerPath(val)
			if err != nil {
				r.logger.Warningf("Couldn't resolve magnet url from value %s\n", val)
				continue
			} else {
				item.MagnetLink = murl
			}
		case "size":
			bytes, err := humanize.ParseBytes(strings.Replace(val, ",", "", -1))
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable size %q: %v", rowIdx, val, err.Error())
				continue
			}
			//r.logger.Debugf("After parsing, size is %v", bytes)
			item.Size = bytes
		case "leechers":
			leechers, err := strconv.Atoi(normalizeNumber(val))
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable leechers value %q in %s", rowIdx, val, key)
				continue
			}
			item.Peers += leechers
		case "seeders":
			seeders, err := strconv.Atoi(normalizeNumber(val))
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable seeders value %q in %s", rowIdx, val, key)
				continue
			}
			item.Seeders = seeders
			item.Peers += seeders
		case "authorId":
			item.AuthorId = val
		case "date":
			t, err := parseFuzzyTime(val, time.Now(), true)
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable time %q in %s", rowIdx, val, key)
				continue
			}
			item.PublishDate = t.Unix()
		case "files":
			files, err := strconv.Atoi(normalizeNumber(val))
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable files value %q in %s", rowIdx, val, key)
				continue
			}
			item.Files = files
		case "grabs":
			grabs, err := strconv.Atoi(normalizeNumber(val))
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable grabs value %q in %s", rowIdx, val, key)
				continue
			}
			item.Grabs = grabs
		case "downloadvolumefactor":
			downloadvolumefactor, err := strconv.ParseFloat(normalizeNumber(val), 64)
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable downloadvolumefactor value %q in %s", rowIdx, val, key)
				continue
			}
			item.DownloadVolumeFactor = downloadvolumefactor
		case "uploadvolumefactor":
			uploadvolumefactor, err := strconv.ParseFloat(normalizeNumber(val), 64)
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable uploadvolumefactor value %q in %s", rowIdx, val, key)
				continue
			}
			item.UploadVolumeFactor = uploadvolumefactor
		case "minimumratio":
			minimumratio, err := strconv.ParseFloat(normalizeNumber(val), 64)
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable minimumratio value %q in %s", rowIdx, val, key)
				continue
			}
			item.MinimumRatio = minimumratio
		case "minimumseedtime":
			minimumseedtime, err := strconv.ParseFloat(normalizeNumber(val), 64)
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable minimumseedtime value %q in %s", rowIdx, val, key)
				continue
			}
			item.MinimumSeedTime = time.Duration(minimumseedtime) * time.Second
		case "banner":
			banner, err := r.resolveIndexerPath(val)
			if err != nil {
				banner = val
			} else {
				item.Banner = banner
			}
		default:
			//r.logger.Warnf("Row #%d has unknown field %s", rowIdx, key)
			item.SetField(key, val)
		}
	}

	if item.GUID == "" && item.Link != "" {
		item.GUID = item.Link
	}

	if r.hasDateHeader() {
		date, err := r.extractDateHeader(selection)
		if err != nil {
			return search.ExternalResultItem{}, err
		}

		item.PublishDate = date.Unix()
	}

	return item, nil
}
