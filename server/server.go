package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/config"
	"github.com/sp0x/rutracker-rss/indexer"
	"github.com/sp0x/rutracker-rss/server/rss"
	"github.com/sp0x/rutracker-rss/torrent"
	"github.com/sp0x/rutracker-rss/torznab"
	"github.com/spf13/viper"
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
	//Params    Params
	indexers   map[string]torznab.Indexer
	config     config.Config
	Port       int
	Hostname   string
	Params     Params
	PathPrefix string
	Password   string
	version    string
}

type Params struct {
	BaseURL    string
	PathPrefix string
	APIKey     []byte
	Passphrase string
	Version    string
	Config     config.Config
}

func NewServer(conf config.Config) *Server {
	s := &Server{}
	s.indexers = map[string]torznab.Indexer{}
	s.config = conf
	s.Port = conf.GetInt("port")
	s.Hostname = conf.GetString("hostname")
	s.Params = Params{
		BaseURL:    fmt.Sprintf("http://%s:%d%s", s.Hostname, s.Port, s.PathPrefix),
		Passphrase: s.Password,
		PathPrefix: s.PathPrefix,
		Config:     s.config,
		Version:    s.version,
		APIKey:     conf.GetBytes("api_key"),
	}
	return s
}

func (s *Server) Listen(tracker *torrent.Rutracker) error {
	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)

	s.tracker = tracker
	s.tabWriter = tabWr
	r := gin.Default()
	s.setupRoutes(r)

	storage = &torrent.Storage{}
	log.Info("Starting server...")
	key, _ := s.sharedKey()
	log.Infof("API Key: %s", key)
	err := r.Run(fmt.Sprintf(":%d", s.Port))
	return err
}

func (s *Server) setupRoutes(r *gin.Engine) {
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
		ifaceConfig := viper.Get("indexer." + key)
		if ifaceConfig != nil {

			indexer, err := s.lookupIndexer(key)
			if err != nil {
				return nil, err
			}
			agg = append(agg, indexer)
		}
		//if config.IsSectionEnabled(key, s.Params) {
		//	indexer, err := s.lookupIndexer(key)
		//	if err != nil {
		//		return nil, err
		//	}
		//	agg = append(agg, indexer)
		//}
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
