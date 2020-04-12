package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sp0x/rutracker-rss/indexer"
	"github.com/sp0x/rutracker-rss/server/rss"
	"github.com/sp0x/rutracker-rss/torrent"
	"github.com/sp0x/rutracker-rss/torznab"
	"net/http"
	"net/url"
	"os"
	"path"
	"text/tabwriter"
)

//// torznab routes
//subrouter.HandleFunc("/torznab/{indexer}", h.torznabHandler).Methods("GET")
//subrouter.HandleFunc("/torznab/{indexer}/api", h.torznabHandler).Methods("GET")

//
type Server struct {
	tracker   *torrent.Rutracker
	tabWriter *tabwriter.Writer
	Params    Params
	indexers  map[string]torznab.Indexer
}

type Params struct {
	BaseURL    string
	PathPrefix string
	APIKey     []byte
	Passphrase string
	Version    string
}

func NewServer() *Server {
	s := &Server{}
	s.indexers = map[string]torznab.Indexer{}
	return s
}

func (s *Server) StartServer(tracker *torrent.Rutracker, port int) error {
	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)

	s.tracker = tracker
	s.tabWriter = tabWr
	r := gin.Default()
	//Rss
	r.GET("/all", s.serveAllTorrents)
	r.GET("/movies", s.serveMovies)
	r.GET("/shows", s.serveShows)
	r.GET("/music", s.serveMusic)
	r.GET("/anime", s.serveAnime)
	r.GET("/search/:name", s.searchAndServe)

	//Torznab
	r.GET("torznab/:indexer", s.torznabHandler)
	r.GET("torznab/:indexer/api", s.torznabHandler)

	err := r.Run(fmt.Sprintf(":%d", port))
	storage = &torrent.Storage{}
	hostname = "localhost"
	return err
}

var storage *torrent.Storage
var hostname string

func (s *Server) serveAllTorrents(c *gin.Context) {
	torrents := storage.GetTorrentsInCategories([]int{})
	rss.SendRssFeed(hostname, "torrents", torrents, c)
}

func (s *Server) createAggregate() (torznab.Indexer, error) {
	keys, err := indexer.DefaultDefinitionLoader.List()
	if err != nil {
		return nil, err
	}

	agg := indexer.Aggregate{}
	for _, key := range keys {
		if config.IsSectionEnabled(key, s.Params) {
			indexer, err := s.lookupIndexer(key)
			if err != nil {
				return nil, err
			}
			agg = append(agg, indexer)
		}
	}

	return agg, nil
}

func (s *Server) baseURL(r *http.Request, appendPath string) (*url.URL, error) {
	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}
	return url.Parse(fmt.Sprintf("%s://%s%s", proto, r.Host,
		path.Join(s.Params.PathPrefix, appendPath)))
}
