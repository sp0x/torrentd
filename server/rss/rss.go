package rss

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"text/tabwriter"
	"time"

	"github.com/gorilla/feeds"
	log "github.com/sirupsen/logrus"

	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/indexer/search"
	http2 "github.com/sp0x/torrentd/server/http"
)

func ServerAll(c http2.Context) {
	// keyedStorage := storage.NewKeyedStorage(nil)
	// torrents := sqlite.DefaultStorageBacking().GetTorrentsInCategories([]int{})
	// SendRssFeed("", "torrents", torrents, c)
}

func SearchAndServe(indexFacade *indexer.Facade, options *indexer.GenericSearchOptions, c http2.Context) {
	tabWriter := new(tabwriter.Writer)
	tabWriter.Init(os.Stdout, 0, 8, 0, '\t', 0)

	currentPage := uint(0)
	name := c.Param("name")
	name = url.QueryEscape(name)
	var items []*search.TorrentResultItem

	//for {
	resultsChannel, err := indexFacade.SearchWithKeywords(name, 1, 1)
	if err != nil {
		log.Warningf("Error while searching for torrent: %s . %s", name, err)
		switch err.(type) {
		case *indexer.LoginError:
			break
		}
	}
	//if currentPage >= options.NumberOfPagesToFetch {
	//	break
	//}
	for resultGroup := range resultsChannel {
		for _, result := range resultGroup {
			torrent := result.(*search.TorrentResultItem)
			if torrent.IsNew() || torrent.IsUpdate() {
				if torrent.IsNew() && !torrent.IsUpdate() {
					_, _ = fmt.Fprintf(tabWriter, "Found new result #%s:\t%s\t[%s]:\t%s\n",
						torrent.LocalID, torrent.AddedOnStr(), torrent.Fingerprint, torrent.Title)
				} else {
					_, _ = fmt.Fprintf(tabWriter, "Updated result #%s:\t%s\t[%s]:\t%s\n",
						torrent.LocalID, torrent.AddedOnStr(), torrent.Fingerprint, torrent.Title)
				}
			} else {
				_, _ = fmt.Fprintf(tabWriter, "Result #%s:\t%s\t[%s]:\t%s\n",
					torrent.LocalID, torrent.AddedOnStr(), "#", torrent.Title)
			}
			items = append(items, torrent)
			_ = tabWriter.Flush()
		}
	}

	currentPage++
	//}
	SendRssFeed("", name, items, c)
}

func ServeShows(c http2.Context) {
	// torrents := sqlite.DefaultStorageBacking().GetTorrentsInCategories([]int{
	//	189,  //Foreign shows
	//	2366, //Foreign shows in HD
	//	2100, //Asian shows
	// })
	// SendRssFeed("", "shows", torrents, c)
}

func ServeMusic(c http2.Context) {
	// torrents := sqlite.DefaultStorageBacking().GetTorrentsInCategories([]int{
	//	409,  // Classical and modern academic music
	//	1125, // Folklore, national and ethnical music
	//	1849, //New age, relax, meditative and flamenco
	//	408,  //Rap, hip-hop and rnb
	//	1760, //Raggae, ska, dub
	//	416,  // OST, karaoke and musicals
	//	413,  //Other music
	//	2497, //Popular foreign music
	// })
	// SendRssFeed("", "music", torrents, c)
}

func ServeAnime(c http2.Context) {
	// torrents := sqlite.DefaultStorageBacking().GetTorrentsInCategories([]int{
	//	33, // Anime
	// })
	// SendRssFeed("", "anime", torrents, c)
}

func ServeMovies(c http2.Context) {
	// torrents := sqlite.DefaultStorageBacking().GetTorrentsInCategories([]int{
	//	7,    //foreign films
	//	124,  //art-house and author movies
	//	93,   //DVD
	//	2198, //HD Video
	//	4,    //Multifilms
	//	352,  //3d/stereo movies, video, tv and sports
	// })
	// SendRssFeed("", "movies", torrents, c)
}

func SendRssFeed(hostname, name string, torrents []*search.TorrentResultItem, c http2.Context) {
	feed := &feeds.Feed{
		Title:       fmt.Sprintf("%s from Rutracker", name),
		Link:        &feeds.Link{Href: fmt.Sprintf("http://%s/%s", hostname, name)},
		Description: name,
		// Author:      &feeds.Author{},
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
