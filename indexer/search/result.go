package search

import (
	"encoding/xml"
	"strconv"
	"time"
)

const rfc822 = "Mon, 02 Jan 2006 15:04:05 -0700"

type ResultItem struct {
	Site        string
	Title       string
	Description string
	GUID        string
	Comments    string
	Link        string

	SourceLink  string
	Category    int
	Size        uint64
	Files       int
	Grabs       int
	PublishDate time.Time

	Seeders              int
	Peers                int
	MinimumRatio         float64
	MinimumSeedTime      time.Duration
	DownloadVolumeFactor float64
	UploadVolumeFactor   float64
}

func (ri ResultItem) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	var enclosure = struct {
		URL    string `xml:"url,attr,omitempty"`
		Length uint64 `xml:"length,attr,omitempty"`
		Type   string `xml:"type,attr,omitempty"`
	}{
		URL:    ri.Link,
		Length: ri.Size,
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
	}{
		Title:       ri.Title,
		Description: ri.Description,
		GUID:        ri.GUID,
		Comments:    ri.Comments,
		Link:        ri.Link,
		Category:    strconv.Itoa(ri.Category),
		Files:       ri.Files,
		Grabs:       ri.Grabs,
		PublishDate: ri.PublishDate.Format(rfc822),
		Enclosure:   enclosure,
	}
	e.Encode(itemView)
	return nil
}
