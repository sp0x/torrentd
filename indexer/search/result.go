package search

import (
	"fmt"
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
	UUIDValue string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

type ModelData map[string]interface{}

type ScrapeLocalData struct {
	// LocalId the id of the item in the index's local sense
	LocalId string
	// Site is the index's site
	Site string
	// Link the originating link for this result
	Link string
	// SourceLink link for data for this result
	SourceLink string
	// PublishDate is the date on which this item was published
	PublishDate int64
}

type ResultItemBase interface {
	Record
	fmt.Stringer
	SetSite(s string)
	SetIndexer(indexer *ResultIndexer)
	SetLocalId(s string)
	SetPublishDate(unix int64)
	AsScrapeItem() *ScrapeResultItem
}

type ScrapeResultItem struct {
	Model
	ModelData
	ScrapeLocalData
	isNew    bool
	isUpdate bool
	Indexer  *ResultIndexer
}

func (i *ScrapeResultItem) AsScrapeItem() *ScrapeResultItem {
	return i
}

func (i *ScrapeResultItem) SetSite(s string) {
	i.Site = s
}

func (i *ScrapeResultItem) SetIndexer(indexer *ResultIndexer) {
	i.Indexer = indexer
}

func (i *ScrapeResultItem) SetLocalId(s string) {
	i.LocalId = s
}

func (i *ScrapeResultItem) SetPublishDate(unix int64) {
	i.PublishDate = unix
}

func (i *ScrapeResultItem) UUID() string {
	return i.UUIDValue
}

func (i *ScrapeResultItem) SetUUID(u string) {
	i.UUIDValue = u
}

// SetState sets the staleness state of this result.
func (i *ScrapeResultItem) SetState(isNew bool, update bool) {
	i.isNew = isNew
	i.isUpdate = update
}

// IsNew whether the result is new to us.
func (i *ScrapeResultItem) IsNew() bool {
	return i.isNew
}

func (i *ScrapeResultItem) Id() uint32 {
	return i.ID
}

func (i *ScrapeResultItem) SetId(id uint32) {
	i.ID = id
}

// IsUpdate whether the result is an update to an existing one.
func (i *ScrapeResultItem) IsUpdate() bool {
	return i.isUpdate
}

// SetField sets the value of an extra fields
func (i *ScrapeResultItem) SetField(key string, val interface{}) {
	i.ModelData[key] = val
}

// GetField by a key, use this for extra fields.
func (i *ScrapeResultItem) GetField(key string) interface{} {
	val, ok := i.ModelData[key]
	if !ok {
		return ""
	}
	return val
}

func (i *ScrapeResultItem) GetFieldWithDefault(key string, defValue interface{}) interface{} {
	val, ok := i.ModelData[key]
	if !ok {
		return defValue
	}
	return val
}

// Equals checks if this item matches the other one exactly(excluding the ID)
func (i *ScrapeResultItem) Equals(other interface{}) bool {
	item, isOkType := other.(*ScrapeResultItem)
	if !isOkType {
		return false
	}
	if (item.ModelData == nil && i.ModelData != nil) ||
		(i.ModelData == nil && item.ModelData != nil) ||
		(len(i.ModelData) != len(item.ModelData)) {
		return false
	}
	// HealthCheck the extra fields
	for key, val := range i.ModelData {
		otherVal, contained := item.ModelData[key]
		if !contained {
			return false
		}
		if val != otherVal {
			return false
		}
	}

	return true
}

func (i *ScrapeResultItem) String() string {
	return fmt.Sprintf("%s: %s", i.LocalId, i.Link)
}

type ResultIndexer struct {
	Id   string `xml:"id,attr"`
	Name string `xml:",chardata"` // make the name the value
}
