package search

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"time"
)

const rfc822 = "Mon, 02 Jan 2006 15:04:05 -0700"

type torznabAttribute struct {
	XMLName struct{} `xml:"torznab:attr"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
}

type Model struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

type ExternalResultItem struct {
	Model
	ResultItem
	LocalCategoryID   string
	LocalCategoryName string
	LocalId           string
	Announce          string
	Publisher         string
	isNew             bool
	isUpdate          bool
	PublishedWith     string
}

//SetState sets the staleness state of this result.
func (i *ExternalResultItem) SetState(isNew bool, update bool) {
	i.isNew = isNew
	i.isUpdate = update
}

//IsNew whether the result is new to us.
func (i *ExternalResultItem) IsNew() bool {
	return i.isNew
}

func (i *ExternalResultItem) Id() uint32 {
	return i.ID
}
func (i *ExternalResultItem) SetId(id uint32) {
	i.ID = id
}

func (i *ExternalResultItem) UUID() string {
	return i.GUID
}

func (i *ExternalResultItem) SetUUID(u string) {
	i.GUID = u
}

//IsUpdate whether the result is an update to an existing one.
func (i *ExternalResultItem) IsUpdate() bool {
	return i.isUpdate
}

//SetField sets the value of an extra fields
func (i *ExternalResultItem) SetField(key string, val interface{}) {
	i.ExtraFields[key] = val
}

//GetField by a key, use this for extra fields.
func (i *ExternalResultItem) GetField(key string) interface{} {
	val, ok := i.ExtraFields[key]
	if !ok {
		return ""
	}
	return val
}

//Equals checks if this item matches the other one exactly(excluding the ID)
//TODO: refactor this to reduce #complexity
func (i *ExternalResultItem) Equals(other interface{}) bool {
	item, isOkType := other.(*ExternalResultItem)
	if !isOkType {
		return false
	}
	if (item.ExtraFields == nil && i.ExtraFields != nil) ||
		(i.ExtraFields == nil && item.ExtraFields != nil) ||
		(len(i.ExtraFields) != len(item.ExtraFields)) {
		return false
	}
	//Check the extra fields
	for key, val := range i.ExtraFields {
		otherVal, contained := item.ExtraFields[key]
		if !contained {
			return false
		}
		if val != otherVal {
			return false
		}
	}
	//Doing this in this way because it's more performant
	if i.IsMagnet != item.IsMagnet {
		return false
	} else if i.Size != item.Size {
		return false
	} else if i.Banner != item.Banner {
		return false
	} else if i.Site != item.Site {
		return false
	} else if i.Link != item.Link {
		return false
	} else if i.Category != item.Category {
		return false
	} else if i.Title != item.Title {
		return false
	} else if i.Seeders != item.Seeders {
		return false
	} else if i.PublishDate != item.PublishDate {
		return false
	} else if i.LocalId != item.LocalId {
		return false
	} else if i.MagnetLink != item.MagnetLink {
		return false
	} else if i.SourceLink != item.SourceLink {
		return false
	} else if i.DownloadVolumeFactor != item.DownloadVolumeFactor {
		return false
	} else if i.ShortTitle != item.ShortTitle {
		return false
	} else if i.Author != item.Author {
		return false
	} else if i.AuthorId != item.AuthorId {
		return false
	} else if i.LocalCategoryID != item.LocalCategoryID {
		return false
	} else if i.LocalCategoryName != item.LocalCategoryName {
		return false
	} else if i.Grabs != item.Grabs {
		return false
	} else if i.OriginalTitle != item.OriginalTitle {
		return false
	} else if i.Fingerprint != item.Fingerprint {
		return false
	} else if i.Publisher != item.Publisher {
		return false
	} else if i.PublishedWith != item.PublishedWith {
		return false
	} else if i.Peers != item.Peers {
		return false
	} else if i.Comments != item.Comments {
		return false
	} else if i.MinimumSeedTime != item.MinimumSeedTime {
		return false
	} else if i.MinimumRatio != item.MinimumRatio {
		return false
	} else if i.Description != item.Description {
		return false
	} else if i.Files != item.Files {
		return false
	}
	return true
}

type ResultItem struct {
	Site          string
	Title         string
	OriginalTitle string
	ShortTitle    string
	Description   string
	GUID          string
	Comments      string
	Link          string
	Fingerprint   string
	Banner        string
	IsMagnet      bool

	SourceLink  string
	MagnetLink  string
	Category    int
	Size        uint32
	Files       int
	Grabs       int
	PublishDate int64

	Seeders              int
	Peers                int
	MinimumRatio         float64
	MinimumSeedTime      time.Duration
	DownloadVolumeFactor float64
	UploadVolumeFactor   float64
	Author               string
	AuthorId             string
	Indexer              *ResultIndexer
	ExtraFields          map[string]interface{} `gorm:"-"` // Ignored in gorm
}

type ResultIndexer struct {
	Id   string `xml:"id,attr"`
	Name string `xml:",chardata"` //make the name the value
}

//AddedOnStr gets the publish date of this result as a string
func (ri *ResultItem) AddedOnStr() string {
	tm := time.Unix(ri.PublishDate, 0)
	return tm.String()
}

func (ri ResultItem) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
	//The info view enclosure
	var enclosure = struct {
		URL    string `xml:"url,attr,omitempty"`
		Length uint32 `xml:"length,attr,omitempty"`
		Type   string `xml:"type,attr,omitempty"`
	}{
		URL:    ri.Link,
		Length: ri.Size,
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
		Title:       ri.Title,
		Description: ri.Description,
		Indexer:     ri.Indexer,
		GUID:        ri.GUID,
		Comments:    ri.Comments,
		Link:        ri.Link,
		Category:    strconv.Itoa(ri.Category),
		Files:       ri.Files,
		Grabs:       ri.Grabs,
		PublishDate: time.Unix(ri.PublishDate, 0).Format(rfc822),
		Enclosure:   enclosure,
		AtomLink:    atomLink,
		Size:        ri.Size,
		Banner:      ri.Banner,
	}
	attribs := itemView.TorznabAttributes
	attribs = append(attribs, torznabAttribute{Name: "category", Value: strconv.Itoa(ri.Category)})
	attribs = append(attribs, torznabAttribute{Name: "seeders", Value: strconv.Itoa(ri.Seeders)})
	attribs = append(attribs, torznabAttribute{Name: "peers", Value: strconv.Itoa(ri.Peers)})
	attribs = append(attribs, torznabAttribute{Name: "minimumratio", Value: fmt.Sprint(ri.MinimumRatio)})
	attribs = append(attribs, torznabAttribute{Name: "minimumseedtime", Value: fmt.Sprint(ri.MinimumSeedTime)})
	attribs = append(attribs, torznabAttribute{Name: "downloadvolumefactor", Value: fmt.Sprint(ri.DownloadVolumeFactor)})
	attribs = append(attribs, torznabAttribute{Name: "uploadvolumefactor", Value: fmt.Sprint(ri.UploadVolumeFactor)})

	itemView.TorznabAttributes = attribs
	_ = e.Encode(itemView)
	return nil
}
