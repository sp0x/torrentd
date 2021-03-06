package server

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	log "github.com/sirupsen/logrus"

	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/torznab"
)

func xmlOutput(c *gin.Context, feed *torznab.ResultFeed, encoding string) {
	x, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		torznab.Error(c, err.Error(), torznab.ErrUnknownError)
		return
	}
	if encoding != "" {
		c.Header("Content-Type", fmt.Sprintf("application/rss+xml; charset=%s", formatEncoding(encoding)))
	} else {
		c.Header("Content-Type", "application/rss+xml")
	}
	x = []byte("<?xml version=\"1.0\" encoding=\"" + encoding + "\"?>\n" + string(x))
	_, _ = c.Writer.Write(x)
}

func atomOutput(c *gin.Context, v *torznab.ResultFeed) {
	feed := &feeds.Feed{
		Title:       v.Info.Title,
		Link:        &feeds.Link{Href: v.Info.Link},
		Description: v.Info.Description,
		// Author:      &feeds.Author{},
		Created: time.Now(),
	}
	feed.Items = make([]*feeds.Item, len(v.Items), len(v.Items))
	for i, torrentItem := range v.Items {
		torr := torrentItem.(*search.TorrentResultItem)
		timep := torr.PublishDate
		feedItem := &feeds.Item{
			Title:       torr.Title,
			Link:        &feeds.Link{Href: torr.Link},
			Description: torr.Link,
			Author:      &feeds.Author{Name: torr.Author},
			Created:     time.Unix(timep, 0),
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

func jsonOutput(w http.ResponseWriter, v interface{}, encoding string) {
	if encoding != "" {
		w.Header().Set("Content-Type", fmt.Sprintf("application/json; charset=%s", formatEncoding(encoding)))
	} else {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}

	_, _ = w.Write(append(b, '\n'))
}
