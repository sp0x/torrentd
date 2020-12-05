package search

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"time"
)

type TorrentResultItem struct {
	ScrapeResultItem
	Title         string
	OriginalTitle string
	ShortTitle    string
	Description   string
	Comments      string
	Link          string
	Fingerprint   string
	Banner        string
	IsMagnet      bool

	SourceLink string
	MagnetLink string
	Category   int
	Size       uint32
	Files      int
	Grabs      int

	Seeders              int
	Peers                int
	MinimumRatio         float64
	MinimumSeedTime      time.Duration
	DownloadVolumeFactor float64
	UploadVolumeFactor   float64
	Author               string
	AuthorId             string
	ExtraFields          map[string]interface{} `gorm:"-"` // Ignored in gorm

	LocalCategoryID   string
	LocalCategoryName string
	Announce          string
	Publisher         string
	PublishedWith     string
}

func (t *TorrentResultItem) String() string {
	return fmt.Sprintf("[%s]%s", t.LocalId, t.Title)
}

//AddedOnStr gets the publish date of this result as a string
func (t *TorrentResultItem) AddedOnStr() string {
	tm := time.Unix(t.PublishDate, 0)
	return tm.String()
}

//MarshalXML marshals the item to xml
func (t TorrentResultItem) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
	//The info view enclosure
	var enclosure = struct {
		URL    string `xml:"url,attr,omitempty"`
		Length uint32 `xml:"length,attr,omitempty"`
		Type   string `xml:"type,attr,omitempty"`
	}{
		URL:    t.Link,
		Length: t.Size,
		Type:   "application/x-bittorrent",
	}
	var atomLink = struct {
		XMLName string `xml:"atom:link"`
		Href    string `xml:"href,attr"`
		Rel     string `xml:"rel,attr"`
		Type    string `xml:"type,attr"`
	}{
		Href: "", Rel: "self", Type: "application/rss+xml",
	}
	var itemView = struct {
		XMLName  struct{} `xml:"item"`
		AtomLink interface{}
		// standard rss elements
		Title             string         `xml:"title,omitempty"`
		Indexer           *ResultIndexer `xml:"indexer,omitempty"`
		Description       string         `xml:"description,omitempty"`
		GUID              string         `xml:"guid,omitempty"`
		Comments          string         `xml:"comments,omitempty"`
		Link              string         `xml:"link,omitempty"`
		Category          string         `xml:"category,omitempty"`
		Files             int            `xml:"files,omitempty"`
		Grabs             int            `xml:"grabs,omitempty"`
		PublishDate       string         `xml:"pubDate,omitempty"`
		Enclosure         interface{}    `xml:"enclosure,omitempty"`
		Size              uint32         `xml:"size"`
		Banner            string         `xml:"banner"`
		TorznabAttributes []torznabAttribute
	}{
		Title:       t.Title,
		Description: t.Description,
		Indexer:     t.Indexer,
		GUID:        t.UUIDValue,
		Comments:    t.Comments,
		Link:        t.Link,
		Category:    strconv.Itoa(t.Category),
		Files:       t.Files,
		Grabs:       t.Grabs,
		PublishDate: time.Unix(t.PublishDate, 0).Format(rfc822),
		Enclosure:   enclosure,
		AtomLink:    atomLink,
		Size:        t.Size,
		Banner:      t.Banner,
	}
	attribs := itemView.TorznabAttributes
	attribs = append(attribs, torznabAttribute{Name: "category", Value: strconv.Itoa(t.Category)})
	attribs = append(attribs, torznabAttribute{Name: "seeders", Value: strconv.Itoa(t.Seeders)})
	attribs = append(attribs, torznabAttribute{Name: "peers", Value: strconv.Itoa(t.Peers)})
	attribs = append(attribs, torznabAttribute{Name: "minimumratio", Value: fmt.Sprint(t.MinimumRatio)})
	attribs = append(attribs, torznabAttribute{Name: "minimumseedtime", Value: fmt.Sprint(t.MinimumSeedTime)})
	attribs = append(attribs, torznabAttribute{Name: "downloadvolumefactor", Value: fmt.Sprint(t.DownloadVolumeFactor)})
	attribs = append(attribs, torznabAttribute{Name: "uploadvolumefactor", Value: fmt.Sprint(t.UploadVolumeFactor)})

	itemView.TorznabAttributes = attribs
	_ = e.Encode(itemView)
	return nil
}

//Equals checks if the other object is equal.
func (t *TorrentResultItem) Equals(other interface{}) bool {
	otherTItem, isOkType := other.(*TorrentResultItem)
	if !isOkType {
		return false
	}
	otherScrapeItem := otherTItem.ScrapeResultItem
	thisScrapeItem := t.ScrapeResultItem
	if !thisScrapeItem.Equals(otherScrapeItem) {
		return false
	}
	//Doing this in this way because it's more performant
	if t.IsMagnet != otherTItem.IsMagnet {
		return false
	} else if t.Size != otherTItem.Size {
		return false
	} else if t.Banner != otherTItem.Banner {
		return false
	} else if t.Site != otherTItem.Site {
		return false
	} else if t.Link != otherTItem.Link {
		return false
	} else if t.Category != otherTItem.Category {
		return false
	} else if t.Title != otherTItem.Title {
		return false
	} else if t.Seeders != otherTItem.Seeders {
		return false
	} else if t.PublishDate != otherTItem.PublishDate {
		return false
	} else if t.LocalId != otherTItem.LocalId {
		return false
	} else if t.MagnetLink != otherTItem.MagnetLink {
		return false
	} else if t.SourceLink != otherTItem.SourceLink {
		return false
	} else if t.DownloadVolumeFactor != otherTItem.DownloadVolumeFactor {
		return false
	} else if t.ShortTitle != otherTItem.ShortTitle {
		return false
	} else if t.Author != otherTItem.Author {
		return false
	} else if t.LocalCategoryID != otherTItem.LocalCategoryID {
		return false
	} else if t.LocalCategoryName != otherTItem.LocalCategoryName {
		return false
	} else if t.AuthorId != otherTItem.AuthorId {
		return false
	} else if t.Grabs != otherTItem.Grabs {
		return false
	} else if t.OriginalTitle != otherTItem.OriginalTitle {
		return false
	} else if t.Fingerprint != otherTItem.Fingerprint {
		return false
	} else if t.Publisher != otherTItem.Publisher {
		return false
	} else if t.PublishedWith != otherTItem.PublishedWith {
		return false
	} else if t.Peers != otherTItem.Peers {
		return false
	} else if t.Comments != otherTItem.Comments {
		return false
	} else if t.MinimumSeedTime != otherTItem.MinimumSeedTime {
		return false
	} else if t.MinimumRatio != otherTItem.MinimumRatio {
		return false
	} else if t.Description != otherTItem.Description {
		return false
	} else if t.Files != otherTItem.Files {
		return false
	}
	return true
}
