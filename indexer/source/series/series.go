package series

import (
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/mediareleaseinfo"

	"github.com/sp0x/torrentd/indexer/search"
)

// Checks if the result is a series result, and ignores it if the title of the series is different.
func IsSeriesAndNotMatching(query *search.Query, item *search.TorrentResultItem) bool {
	if query.Series != "" {
		info, err := releaseinfo.Parse(item.Title)
		if err != nil {
			log.
				WithFields(log.Fields{"title": item.Title}).
				WithError(err).
				Warn("Failed to parse show title, skipping")
			return true
		}

		if info != nil && !info.SeriesTitleInfo.Equal(query.Series) {
			log.
				WithFields(log.Fields{"got": info.SeriesTitleInfo.TitleWithoutYear, "expected": query.Series}).
				Debugf("Series search skipping non-matching series")
			return true
		}
	}
	return false
}
