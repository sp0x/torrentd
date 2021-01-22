package torrent

import (
	"fmt"
	"strings"

	imdbscraper "github.com/cardigann/go-imdb-scraper"
	"github.com/mrobinsn/go-tvmaze/tvmaze"

	"github.com/sp0x/torrentd/indexer/search"
)

// EnrichMovieAndShowData If the query is for an item id from an external service, then resolve the query keywords.
func EnrichMovieAndShowData(query *search.Query) error {
	var show *tvmaze.Show
	var movie *imdbscraper.Movie
	var err error

	// convert show identifiers to season parameter
	switch {
	case query.TVDBID != "" && query.TVDBID != "0":
		show, err = tvmaze.DefaultClient.GetShowWithTVDBID(query.TVDBID)
		query.TVDBID = "0"
	case query.TVMazeID != "":
		show, err = tvmaze.DefaultClient.GetShowWithID(query.TVMazeID)
		query.TVMazeID = "0"
	case query.TVRageID != "":
		show, err = tvmaze.DefaultClient.GetShowWithTVRageID(query.TVRageID)
		query.TVRageID = ""
	case query.IMDBID != "":
		imdbid := query.IMDBID
		if !strings.HasPrefix(imdbid, "tt") {
			imdbid = "tt" + imdbid
		}
		movie, err = imdbscraper.FindByID(imdbid)
		if err != nil {
			err = fmt.Errorf("imdb error. %s", err)
		}
		query.IMDBID = ""
	}

	if err != nil {
		return err
	}

	if show != nil {
		query.Series = show.Name
	}

	if movie != nil {
		if movie.Title == "" {
			return fmt.Errorf("movie title was blank")
		}
		query.Movie = movie.Title
		query.Year = movie.Year
	}

	return nil
}
