package rss

import (
	"fmt"
	"github.com/gorilla/feeds"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer/search"
	"net/http"
	"time"
)

func SendRssFeed(hostname, name string, torrents []search.ExternalResultItem, c HttpContext) {
	feed := &feeds.Feed{
		Title:       fmt.Sprintf("%s from Rutracker", name),
		Link:        &feeds.Link{Href: fmt.Sprintf("http://%s/%s", hostname, name)},
		Description: name,
		//Author:      &feeds.Author{},
		Created: time.Now(),
	}
	feed.Items = make([]*feeds.Item, len(torrents))
	for i, torr := range torrents {
		timep := time.Unix(torr.PublishDate, 0)
		feedItem := &feeds.Item{
			Title:       torr.Title,
			Link:        &feeds.Link{Href: torr.SourceLink},
			Description: torr.Link,
			Author:      &feeds.Author{Name: torr.Author},
			Created:     timep,
		}
		feed.Items[i] = feedItem
	}
	rss, err := feed.ToAtom()
	if err != nil {
		log.Fatal(err)
	}
	c.Header("Content-Type", "application/xml;")
	c.String(http.StatusOK, rss)
}
