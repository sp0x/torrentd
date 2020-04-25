package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/indexer"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"github.com/sp0x/rutracker-rss/server/rss"
	"net/url"
)

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
	var srch *search.Search
	ops := s.tracker.GetDefaultOptions()
	currentPage := uint(0)
	name := c.Param("name")
	name = url.QueryEscape(name)
	var items []search.ExternalResultItem
	for true {
		var err error
		if srch == nil {
			srch, err = s.tracker.Search(nil, name, 0)
		} else {
			srch, err = s.tracker.Search(srch, name, currentPage)
		}
		if err != nil {
			log.Warningf("Error while searching for torrent: %s . %s", name, err)
			switch err.(type) {
			case *indexer.LoginError:
				break
			}
		}
		if currentPage >= ops.PageCount {
			break
		}
		for _, torrent := range srch.Results {
			//isNew, isUpdate := torrent.HandleTorrentDiscovery(tr)
			if torrent.IsNew() || torrent.IsUpdate() {
				if torrent.IsNew() && !torrent.IsUpdate() {
					_, _ = fmt.Fprintf(s.tabWriter, "Found new torrent #%s:\t%s\t[%s]:\t%s\n",
						torrent.LocalId, torrent.AddedOnStr(), torrent.Fingerprint, torrent.Title)
				} else {
					_, _ = fmt.Fprintf(s.tabWriter, "Updated torrent #%s:\t%s\t[%s]:\t%s\n",
						torrent.LocalId, torrent.AddedOnStr(), torrent.Fingerprint, torrent.Title)
				}
			} else {
				_, _ = fmt.Fprintf(s.tabWriter, "Torrent #%s:\t%s\t[%s]:\t%s\n",
					torrent.LocalId, torrent.AddedOnStr(), "#", torrent.Title)
			}
			items = append(items, torrent)
			s.tabWriter.Flush()
		}

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
