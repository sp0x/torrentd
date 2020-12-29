package torznab

import (
	"encoding/xml"

	"github.com/sp0x/torrentd/indexer/search"
)

type Info struct {
	ID          string
	Title       string
	Description string
	Link        string
	Language    string
	Category    string
}

type ResultFeed struct {
	Info  Info
	Items []search.ResultItemBase
}

func (rf ResultFeed) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	channelView := struct {
		XMLName     struct{} `xml:"channel"`
		Title       string   `xml:"title,omitempty"`
		Description string   `xml:"description,omitempty"`
		Link        string   `xml:"link,omitempty"`
		Language    string   `xml:"language,omitempty"`
		Category    string   `xml:"category,omitempty"`
		Items       []search.ResultItemBase
	}{
		Title:       rf.Info.Title,
		Description: rf.Info.Description,
		Link:        rf.Info.Link,
		Language:    rf.Info.Language,
		Category:    rf.Info.Category,
		Items:       rf.Items,
	}

	_ = e.Encode(struct {
		XMLName          struct{}    `xml:"rss"`
		TorznabNamespace string      `xml:"xmlns:torznab,attr"`
		AtomNamespace    string      `xml:"xmlns:atom,attr"`
		Version          string      `xml:"version,attr,omitempty"`
		Channel          interface{} `xml:"channel"`
	}{
		Version:          "2.0",
		Channel:          channelView,
		AtomNamespace:    "http://www.w3.org/2005/Atom",
		TorznabNamespace: "http://torznab.com/schemas/2015/feed",
	})
	return nil
}
