package torznab

import (
	"encoding/xml"
	"fmt"
	"github.com/sp0x/rutracker-rss/torrent/search"
	"strconv"
)

const rfc822 = "Mon, 02 Jan 2006 15:04:05 -0700"

func MarshalXML(result search.ResultItem, e *xml.Encoder, start xml.StartElement) error {
	var enclosure = struct {
		URL    string `xml:"url,attr,omitempty"`
		Length uint64 `xml:"length,attr,omitempty"`
		Type   string `xml:"type,attr,omitempty"`
	}{
		URL:    result.Link,
		Length: result.Size,
		Type:   "application/x-bittorrent",
	}

	var itemView = struct {
		XMLName struct{} `xml:"item"`

		// standard rss elements
		Title       string      `xml:"title,omitempty"`
		Description string      `xml:"description,omitempty"`
		GUID        string      `xml:"guid,omitempty"`
		Comments    string      `xml:"comments,omitempty"`
		Link        string      `xml:"link,omitempty"`
		Category    string      `xml:"category,omitempty"`
		Files       int         `xml:"files,omitempty"`
		Grabs       int         `xml:"grabs,omitempty"`
		PublishDate string      `xml:"pubDate,omitempty"`
		Enclosure   interface{} `xml:"enclosure,omitempty"`

		// torznab elements
		Attrs []torznabAttrView
	}{
		Title:       result.Title,
		Description: result.Description,
		GUID:        result.GUID,
		Comments:    result.Comments,
		Link:        result.Link,
		Category:    strconv.Itoa(result.Category),
		Files:       result.Files,
		Grabs:       result.Grabs,
		PublishDate: result.PublishDate.Format(rfc822),
		Enclosure:   enclosure,
		Attrs: []torznabAttrView{
			{Name: "site", Value: result.Site},
			{Name: "seeders", Value: strconv.Itoa(result.Seeders)},
			{Name: "peers", Value: strconv.Itoa(result.Peers)},
			{Name: "minimumratio", Value: fmt.Sprintf("%.2f", result.MinimumRatio)},
			{Name: "minimumseedtime", Value: fmt.Sprintf("%.f", result.MinimumSeedTime.Seconds())},
			{Name: "size", Value: fmt.Sprintf("%d", result.Size)},
			{Name: "downloadvolumefactor", Value: fmt.Sprintf("%.2f", result.DownloadVolumeFactor)},
			{Name: "uploadvolumefactor", Value: fmt.Sprintf("%.2f", result.UploadVolumeFactor)},
		},
	}

	e.Encode(itemView)
	return nil
}

type torznabAttrView struct {
	XMLName struct{} `xml:"torznab:attr"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
}

type ResultFeed struct {
	Info  Info
	Items []search.ResultItem
}

func (rf ResultFeed) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	var channelView = struct {
		XMLName     struct{} `xml:"channel"`
		Title       string   `xml:"title,omitempty"`
		Description string   `xml:"description,omitempty"`
		Link        string   `xml:"link,omitempty"`
		Language    string   `xml:"language,omitempty"`
		Category    string   `xml:"category,omitempty"`
		Items       []search.ResultItem
	}{
		Title:       rf.Info.Title,
		Description: rf.Info.Description,
		Link:        rf.Info.Link,
		Language:    rf.Info.Language,
		Category:    rf.Info.Category,
		Items:       rf.Items,
	}

	e.Encode(struct {
		XMLName          struct{}    `xml:"rss"`
		TorznabNamespace string      `xml:"xmlns:torznab,attr"`
		Version          string      `xml:"version,attr,omitempty"`
		Channel          interface{} `xml:"channel"`
	}{
		Version:          "2.0",
		Channel:          channelView,
		TorznabNamespace: "http://torznab.com/schemas/2015/feed",
	})
	return nil
}
