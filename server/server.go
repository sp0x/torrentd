package server

import (
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/server/rss"
	"github.com/sp0x/torrentd/storage"
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
	tracker   *indexer.Facade
	tabWriter *tabwriter.Writer
	//Params    Params
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

func (s *Server) Listen(tracker *indexer.Facade) error {
	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)

	s.tracker = tracker
	s.tabWriter = tabWr
	r := gin.Default()
	//Register pprof so we can profile our app.
	pprof.Register(r)
	s.setupRoutes(r)
	log.Info("Starting server...")
	key := s.sharedKey()
	log.Infof("API Key: %s", key)
	err := r.Run(fmt.Sprintf(":%d", s.Port))
	return err
}

var hostname string

func (s *Server) serveAllTorrents(c *gin.Context) {
	torrents := storage.DefaultStorage().GetTorrentsInCategories([]int{})
	rss.SendRssFeed(hostname, "torrents", torrents, c)
}

func (s *Server) baseURL(r *http.Request, appendPath string) (*url.URL, error) {
	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}
	return url.Parse(fmt.Sprintf("%s://%s%s", proto, r.Host,
		path.Join(s.Params.PathPrefix, appendPath)))
}
