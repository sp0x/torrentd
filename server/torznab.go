package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/indexer/cache"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/torrent"
	"github.com/sp0x/torrentd/torznab"
)

func (s *Server) aggregatesStatus(c *gin.Context) {
	indexes, err := s.indexerFacade.IndexScope.Lookup(s.config, "all")
	if err != nil {
		_ = c.Error(err)
		return
	}

	statusObj := struct {
		ActiveIndexers []string `json:"latest"`
	}{}
	for _, ixr := range indexes {
		ixrInfo := ixr.Info()
		statusObj.ActiveIndexers = append(statusObj.ActiveIndexers, ixrInfo.GetTitle())
	}
	c.JSON(200, statusObj)
}

var searchCache, _ = cache.NewTTL(100, 1*time.Hour)

// torznabIndexCapabilities godoc
// @Summary      Torznab capabilities
// @Description  Get the index(es) capabilities in torznab format
// @Tags         torznab
// @Accept       */*
// @param        indexes path string false "Index name(s) to get the capabilities for"
// @Router       /torznab/caps/{indexes} [get]
// @Produce      xml
// @Success      200  {object}  torznab.Capabilities
// @Failure 	 404  {object}  torznab.err
func (s *Server) torznabIndexCapabilities(c *gin.Context) {
	indexerID := indexer.ResolveIndexID(s.indexerFacade.IndexScope, c.Param("indexes"))
	searchIndexes, err := s.indexerFacade.IndexScope.Lookup(s.config, indexerID)
	if err != nil {
		torznab.Error(c, err.Error(), torznab.ErrIncorrectParameter)
		return
	}

	searchIndex := searchIndexes[0]

	searchIndex.Capabilities().ServeHTTP(c.Writer, c.Request)
}

// torznabHandler godoc
// @Summary      Torznab handler
// @Description  Query indexes in torznab format
// @Tags         torznab
// @Accept       */*
// @param        indexes path string false "Index name(s) to search through"
// @param        t query string false "Type of search. Can be caps, search, tvsearch, tv-search, movie, movie-search, moviesearch. Defaults to caps, returning the capabilities."
// @param 	  	 q query string false "Search query"
// @param 	  	 cat query string false "Category"
// @param 	  	 format query string false "The output format to use"
// @param 	  	 imdbid query string false "IMDB ID"
// @param 	  	 tmdbid query string false "TMDB ID"
// @param 	  	 rid query string false "TVDB ID"
// @param 	  	 season query string false "Season number"
// @param 	  	 ep query string false "Episode number"
// @param 	  	 limit query string false "Limit the number of results, defaults to 20"
// @param 	  	 offset query string false "Offset the results"
// @param 	  	 minage query string false "Minimum age of the torrent"
// @param 	  	 maxage query string false "Maximum age of the torrent"
// @param 	  	 minsize query string false "Minimum size of the torrent"
// @param 	  	 maxsize query string false "Maximum size of the torrent"
// @param 	  	 apikey query string false "API key"
// @Produce      xml,json
// @Success      200  {object}  torznab.ResultFeed
// @Failure 	 404 {type} string "404 page not found"
// @Router       /torznab/{indexes} [get]
func (s *Server) torznabHandler(c *gin.Context) {
	_ = c.Params

	t := c.Query("t")
	if t == "" {
		// Redirect to capabilities
		http.Redirect(c.Writer, c.Request, "/torznab/caps/"+c.Param("indexes"), http.StatusTemporaryRedirect)
		return
	} else if t == "caps" {
		s.torznabIndexCapabilities(c)
		return
	}

	indexerID := indexer.ResolveIndexID(s.indexerFacade.IndexScope, c.Param("indexes"))
	searchIndexes, err := s.indexerFacade.IndexScope.Lookup(s.config, indexerID)
	if err != nil {
		torznab.Error(c, err.Error(), torznab.ErrIncorrectParameter)
		return
	}

	searchIndex := searchIndexes[0]
	if !s.checkAPIKey(c.Query("apikey")) {
		torznab.Error(c, "Invalid apikey parameter", torznab.ErrInsufficientPrivs)
		return
	}

	switch t {
	case "search", "tvsearch", "tv-search", "movie", "movie-search", "moviesearch":
		query, err := search.NewQueryFromUrl(c.Request.URL.Query())
		if err != nil {
			torznab.Error(c, "Invalid query", torznab.ErrInsufficientPrivs)
			return
		}
		err = torrent.EnrichMovieAndShowData(query)
		if err != nil {
			torznab.Error(c, "Invalid query", torznab.ErrInsufficientPrivs)
			return
		}

		var feed *torznab.ResultFeed
		if cachedFeed, ok := searchCache.Get(query.UniqueKey()); ok {
			feed = cachedFeed.(*torznab.ResultFeed)
		} else {
			feed, err = s.torznabSearch(c.Request, query, s.indexerFacade)
			searchCache.Add(query.UniqueKey(), feed)
		}
		if err != nil {
			torznab.Error(c, err.Error(), torznab.ErrUnknownError)
			return
		}
		switch c.Query("format") {
		case "atom":
			atomOutput(c, feed)
		case "", "xml":
			xmlOutput(c, feed, searchIndex.GetEncoding())
		case "json":
			jsonOutput(c.Writer, feed, searchIndex.GetEncoding())
		}

	default:
		torznab.Error(c, "Unknown type parameter", torznab.ErrIncorrectParameter)
	}
}

func formatEncoding(nm string) string {
	nm = strings.ReplaceAll(nm, "ows12", "ows-12")
	nm = strings.Title(nm)
	return nm
}

func (s *Server) torznabSearch(r *http.Request, query *search.Query, indexFacade *indexer.Facade) (*torznab.ResultFeed, error) {
	resultsChan, err := indexFacade.Search(query)
	if err != nil {
		return nil, err
	}
	nfo := indexFacade.Indexes.Info()
	var results []search.ResultItemBase
	for resultPage := range resultsChan {
		results = append(results, resultPage...)
	}

	feed := &torznab.ResultFeed{
		Info: torznab.Info{
			ID:          nfo.GetID(),
			Title:       nfo.GetTitle(),
			Description: "",
			Link:        nfo.GetLink(),
			Language:    nfo.GetLanguage(),
			Category:    "",
		},
		Items: results,
	}
	feed.Info.Category = query.Type

	rewritten, err := s.rewriteLinks(r, results)
	if err != nil {
		return nil, err
	}

	feed.Items = rewritten

	return feed, err
}

// Rewrites the download links so that the download goes through us.
// This is required since only we can access the torrent ( the site might need authorization )
func (s *Server) rewriteLinks(r *http.Request, items []search.ResultItemBase) ([]search.ResultItemBase, error) {
	baseURL, err := s.baseURL(r, "/d")
	if err != nil {
		return nil, err
	}
	apiKey := s.sharedKey()
	// rewrite non-magnet links to use the server
	for idx, item := range items {
		scrapeItem := item.AsScrapeItem()
		if strings.HasPrefix(scrapeItem.Link, "magnet:") {
			continue
		}
		// itemTmp := item
		tokenValue, err := getTokenValue(scrapeItem, apiKey)
		if err != nil {
			return nil, err
		}
		filename := strings.ReplaceAll(item.UUID(), "/", "-")
		items[idx].AsScrapeItem().Link = fmt.Sprintf("%s/%s/%s.torrent", baseURL.String(), tokenValue, url.QueryEscape(filename))
	}

	return items, nil
}

func getTokenValue(item *search.ScrapeResultItem, apiKey []byte) (string, error) {
	sourceLink := item.SourceLink
	if sourceLink == "" {
		sourceLink = item.Link
	}
	indexerName := ""
	if item.Indexer != nil {
		indexerName = item.Indexer.Name
	} else {
		indexerName = item.Site
	}
	// Encode the site and source of the torrent as a JWT token
	t := &token{
		IndexName: indexerName,
		Link:      sourceLink,
	}

	te, err := t.Encode(apiKey)
	if err != nil {
		log.Debugf("Error encoding token: %v", err)
		return "", err
	}
	return te, nil
}
