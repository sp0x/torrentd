package indexer

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"github.com/yosssi/gohtml"
	"strconv"
	"strings"
	"time"
)

//Extract the actual result item from it's row/col
func (r *Runner) extractItem(rowIdx int, selection *goquery.Selection) (extractedItem, error) {
	row := map[string]string{}

	html, _ := goquery.OuterHtml(selection)
	r.logger.WithFields(logrus.Fields{"html": gohtml.Format(html)}).Debug("Processing row")

	for _, item := range r.definition.Search.Fields {
		r.logger.
			WithFields(logrus.Fields{"row": rowIdx, "block": item.Block.String()}).
			Debugf("Processing field %q", item.Field)
		val, err := item.Block.MatchText(selection)
		if err != nil {
			return extractedItem{}, err
		}

		r.logger.
			WithFields(logrus.Fields{"row": rowIdx, "output": val}).
			Debugf("Finished processing field %q", item.Field)

		row[item.Field] = val
	}

	//Evaluate pattern items
	for _, item := range r.definition.Search.Fields {
		if item.Block.Pattern == "" {
			continue
		}
		val := row[item.Field]
		templateData := row
		updated, err := r.applyTemplate("result_template", val, templateData)
		if err != nil {
			return extractedItem{}, err
		}
		val = updated
		row[item.Field] = val
	}

	item := extractedItem{
		ResultItem: search.ResultItem{
			Site: r.definition.Site,
		},
	}

	r.logger.
		WithFields(logrus.Fields{"row": rowIdx, "data": row}).
		Debugf("Finished row %d", rowIdx)

	for key, val := range row {
		switch key {
		case "download":
			u, err := r.resolvePath(val)
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable url %q in %s", rowIdx, val, key)
				continue
			}
			item.Link = u
		case "link":
			u, err := r.resolvePath(val)
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable url %q in %s", rowIdx, val, key)
				continue
			}
			item.SourceLink = u
		case "details":
			u, err := r.resolvePath(val)
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
			u, err := r.resolvePath(val)
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable url %q in %s", rowIdx, val, key)
				continue
			}
			item.Comments = u
		case "title":
			item.Title = val
		case "description":
			item.Description = val
		case "category":
			item.LocalCategoryID = val
		case "categoryName":
			item.LocalCategoryName = val
		case "size":
			bytes, err := humanize.ParseBytes(strings.Replace(val, ",", "", -1))
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable size %q: %v", rowIdx, val, err.Error())
				continue
			}
			r.logger.Debugf("After parsing, size is %v", bytes)
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
		case "date":
			t, err := parseFuzzyTime(val, time.Now())
			if err != nil {
				r.logger.Warnf("Row #%d has unparseable time %q in %s", rowIdx, val, key)
				continue
			}
			item.PublishDate = t
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
		default:
			r.logger.Warnf("Row #%d has unknown field %s", rowIdx, key)
			continue
		}
	}

	if item.GUID == "" && item.Link != "" {
		item.GUID = item.Link
	}

	if r.hasDateHeader() {
		date, err := r.extractDateHeader(selection)
		if err != nil {
			return extractedItem{}, err
		}

		item.PublishDate = date
	}

	return item, nil
}
