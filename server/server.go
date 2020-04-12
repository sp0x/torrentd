package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/db"
	"github.com/sp0x/rutracker-rss/server/rss"
	"github.com/sp0x/rutracker-rss/torrent"
	"net/url"
	"os"
	"text/tabwriter"
)

//// torznab routes
//subrouter.HandleFunc("/torznab/{indexer}", h.torznabHandler).Methods("GET")
//subrouter.HandleFunc("/torznab/{indexer}/api", h.torznabHandler).Methods("GET")

type Server struct {
	tracker   *torrent.Rutracker
	tabWriter *tabwriter.Writer
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

func (s *Server) serveMusic(c *gin.Context) {
	torrents := storage.GetTorrentsInCategories([]int{
		409,  // Classical and modern academic music
		1125, // Folklore, national and ethnical music
		1849, //New age, relax, meditative and flamenco
		408,  //Rap, hip-hop and rnb
		1760, //Raggae, ska, dub
		416,  // OST, karaoke and musicals
		413,  //Other music
		2497, //Popular foreign music
	})
	rss.SendRssFeed(hostname, "music", torrents, c)
}

func (s *Server) serveAnime(c *gin.Context) {
	torrents := storage.GetTorrentsInCategories([]int{
		33, // Anime
	})
	rss.SendRssFeed(hostname, "anime", torrents, c)
}

func (s *Server) searchAndServe(c *gin.Context) {
	var search *torrent.Search
	ops := s.tracker.GetDefaultOptions()
	currentPage := uint(0)
	name := c.Param("name")
	name = url.QueryEscape(name)
	var items []db.Torrent
	for true {
		var err error
		if search == nil {
			search, err = s.tracker.Search(nil, name, 0)
		} else {
			search, err = s.tracker.Search(search, name, currentPage)
		}
		if err != nil {
			log.Warningf("Error while searching for torrent: %s . %s", name, err)
		}
		if currentPage >= ops.PageCount {
			break
		}
		s.tracker.ParseTorrents(search.GetDocument(), func(i int, tr *db.Torrent) {
			isNew, isUpdate := torrent.HandleTorrentDiscovery(s.tracker, tr)
			if isNew || isUpdate {
				if isNew && !isUpdate {
					_, _ = fmt.Fprintf(s.tabWriter, "Found new torrent #%s:\t%s\t[%s]:\t%s\n",
						tr.TorrentId, tr.AddedOnStr(), tr.Fingerprint, tr.Name)
				} else {
					_, _ = fmt.Fprintf(s.tabWriter, "Updated torrent #%s:\t%s\t[%s]:\t%s\n",
						tr.TorrentId, tr.AddedOnStr(), tr.Fingerprint, tr.Name)
				}
			} else {
				_, _ = fmt.Fprintf(s.tabWriter, "Torrent #%s:\t%s\t[%s]:\t%s\n",
					tr.TorrentId, tr.AddedOnStr(), "#", tr.Name)
			}
			items = append(items, *tr)
			s.tabWriter.Flush()
		})

		currentPage++
	}
	rss.SendRssFeed(hostname, name, items, c)
}

func (s *Server) serveShows(c *gin.Context) {
	torrents := storage.GetTorrentsInCategories([]int{
		189,  //Foreign shows
		2366, //Foreign shows in HD
		2100, //Asian shows
	})
	rss.SendRssFeed(hostname, "shows", torrents, c)
}

func (s *Server) serveMovies(c *gin.Context) {
	torrents := storage.GetTorrentsInCategories([]int{
		7,    //foreign films
		124,  //art-house and author movies
		93,   //DVD
		2198, //HD Video
		4,    //Multifilms
		352,  //3d/stereo movies, video, tv and sports
	})
	rss.SendRssFeed(hostname, "movies", torrents, c)
}

func (s *Server) serveAllTorrents(c *gin.Context) {
	torrents := storage.GetTorrentsInCategories([]int{})
	rss.SendRssFeed(hostname, "torrents", torrents, c)
}
