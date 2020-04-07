package rss

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/db"
	"github.com/sp0x/rutracker-rss/torrent"
	"net/http"
	"time"
)

var storage *torrent.Storage
var hostname string

func StartServer(port int) error {
	r := gin.Default()
	r.GET("/all", serveAllTorrents)
	r.GET("/movies", serveMovies)
	r.GET("/shows", serveShows)
	r.GET("/music", serveMusic)
	r.GET("/anime", serveAnime)
	err := r.Run(fmt.Sprintf(":%d", port))
	storage = &torrent.Storage{}
	hostname = "localhost"
	return err
}

func serveMusic(c *gin.Context) {
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
	sendFeed("music", torrents, c)
}

func serveAnime(c *gin.Context) {
	torrents := storage.GetTorrentsInCategories([]int{
		33, // Anime
	})
	sendFeed("anime", torrents, c)
}

func serveShows(c *gin.Context) {
	torrents := storage.GetTorrentsInCategories([]int{
		189,  //Foreign shows
		2366, //Foreign shows in HD
		2100, //Asian shows
	})
	sendFeed("shows", torrents, c)
}

func serveMovies(c *gin.Context) {
	torrents := storage.GetTorrentsInCategories([]int{
		7,    //foreign films
		124,  //art-house and author movies
		93,   //DVD
		2198, //HD Video
		4,    //Multifilms
		352,  //3d/stereo movies, video, tv and sports
	})
	sendFeed("movies", torrents, c)
}

func serveAllTorrents(c *gin.Context) {
	torrents := storage.GetTorrentsInCategories([]int{})
	sendFeed("torrents", torrents, c)
}

func sendFeed(name string, torrents []db.Torrent, c *gin.Context) {
	feed := &feeds.Feed{
		Title:       fmt.Sprintf("%s from Rutracker", name),
		Link:        &feeds.Link{Href: fmt.Sprintf("http://%s/%s", hostname, name)},
		Description: name,
		//Author:      &feeds.Author{},
		Created: time.Now(),
	}
	feed.Items = make([]*feeds.Item, len(torrents), len(torrents))
	for i, torr := range torrents {
		timep, _ := time.Parse("2006-01-02 3:04PM", torr.AddedOn)
		feedItem := &feeds.Item{
			Title:       torr.Name,
			Link:        &feeds.Link{Href: torr.DownloadLink},
			Description: torr.Link,
			Author:      &feeds.Author{Name: torr.AuthorName},
			Created:     timep,
		}
		feed.Items[i] = feedItem
	}
	rss, err := feed.ToRss()
	if err != nil {
		log.Fatal(err)
	}
	c.Header("Content-Type", "application/xml;")
	c.String(http.StatusOK, rss)
}
