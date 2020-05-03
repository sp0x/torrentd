package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/indexer"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"github.com/sp0x/rutracker-rss/torznab"
	"net/http"
	"net/url"
	"strings"
)

func (s *Server) aggregatesStatus(c *gin.Context) {
	aggregate, err := indexer.Lookup(s.config, "all")
	if err != nil {
		c.Error(err)
		return
	}

	statusObj := struct {
		ActiveIndexers []string `json:"latest"`
	}{}
	for _, ixr := range aggregate.(*indexer.Aggregate).Indexers {
		ixrInfo := ixr.Info()
		statusObj.ActiveIndexers = append(statusObj.ActiveIndexers, ixrInfo.GetTitle())
	}
	c.JSON(200, statusObj)
}

func (s *Server) torznabHandler(c *gin.Context) {
	_ = c.Params
	indexerID := c.Param("indexer")
	//Type of operation
	t := c.Query("t")
	indexer, err := indexer.Lookup(s.config, indexerID)
	if err != nil {
		torznab.Error(c, err.Error(), torznab.ErrIncorrectParameter)
		return
	}
	switch t {
	case "caps":
		indexer.Capabilities().ServeHTTP(c.Writer, c.Request)
		return
	}

	apiKey := c.Query("apikey")
	if !s.checkAPIKey(apiKey) {
		torznab.Error(c, "Invalid apikey parameter", torznab.ErrInsufficientPrivs)
		return
	}

	if t == "" {
		http.Redirect(c.Writer, c.Request, c.Request.URL.Path+"?t=caps", http.StatusTemporaryRedirect)
		return
	}

	switch t {
	case "caps":
		indexer.Capabilities().ServeHTTP(c.Writer, c.Request)

	case "search", "tvsearch", "tv-search", "movie", "movie-search", "moviesearch":
		feed, err := s.torznabSearch(c.Request, indexer, indexerID)
		if err != nil {
			torznab.Error(c, err.Error(), torznab.ErrUnknownError)
			return
		}
		switch c.Query("format") {
		case "atom":
			atomOutput(c, feed, indexer.GetEncoding())
		case "", "xml":
			xmlOutput(c, feed, indexer.GetEncoding())
		case "json":
			jsonOutput(c.Writer, feed, indexer.GetEncoding())
		}

	default:
		torznab.Error(c, "Unknown type parameter", torznab.ErrIncorrectParameter)
	}
}

func formatEncoding(nm string) string {
	nm = strings.Replace(nm, "ows12", "ows-12", -1)
	nm = strings.Title(nm)
	return nm
}

func (s *Server) torznabSearch(r *http.Request, indexer indexer.Indexer, siteKey string) (*torznab.ResultFeed, error) {
	query, err := torznab.ParseQuery(r.URL.Query())
	if err != nil {
		return nil, err
	}

	srch, err := indexer.Search(query, nil)
	if err != nil {
		return nil, err
	}
	nfo := indexer.Info()

	feed := &torznab.ResultFeed{
		Info: torznab.Info{
			ID:          nfo.GetId(),
			Title:       nfo.GetTitle(),
			Description: "",
			Link:        nfo.GetLink(),
			Language:    nfo.GetLanguage(),
			Category:    "",
		},
		Items: srch.GetResults(),
	}
	feed.Info.Category = query.Type

	rewritten, err := s.rewriteLinks(r, srch.GetResults())
	if err != nil {
		return nil, err
	}

	feed.Items = rewritten

	return feed, err
}

//Rewrites the download links so that the download goes through us.
//This is required since only we can access the torrent ( the site might need authorization )
func (s *Server) rewriteLinks(r *http.Request, items []search.ExternalResultItem) ([]search.ExternalResultItem, error) {
	baseURL, err := s.baseURL(r, "/d")
	if err != nil {
		return nil, err
	}
	k, err := s.sharedKey()
	if err != nil {
		return nil, err
	}
	// rewrite non-magnet links to use the server
	for idx, item := range items {
		if strings.HasPrefix(item.Link, "magnet:") {
			continue
		}
		sourceLink := item.SourceLink
		if sourceLink == "" {
			sourceLink = item.Link
		}
		//Encode the site and source of the torrent as a JWT token
		t := &token{
			Site: item.Site,
			Link: sourceLink,
		}

		te, err := t.Encode(k)
		if err != nil {
			log.Debugf("Error encoding token: %v", err)
			return nil, err
		}

		filename := strings.Replace(item.Title, "/", "-", -1)
		items[idx].Link = fmt.Sprintf("%s/%s/%s.torrent", baseURL.String(), te, url.QueryEscape(filename))
	}

	return items, nil
}
