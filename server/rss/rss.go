package rss

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/db"
	"net/http"
	"time"
)

func SendRssFeed(hostname, name string, torrents []db.Torrent, c *gin.Context) {
	feed := &feeds.Feed{
		Title:       fmt.Sprintf("%s from Rutracker", name),
		Link:        &feeds.Link{Href: fmt.Sprintf("http://%s/%s", hostname, name)},
		Description: name,
		//Author:      &feeds.Author{},
		Created: time.Now(),
	}
	feed.Items = make([]*feeds.Item, len(torrents), len(torrents))
	for i, torr := range torrents {
		timep := time.Unix(torr.AddedOn, 0)
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
