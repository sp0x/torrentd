package db

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Torrent struct {
	gorm.Model
	Name         string
	TorrentId    string
	AddedOn      int64 // *time.Time
	Link         string
	Fingerprint  string
	AuthorName   string
	AuthorId     string
	CategoryName string
	CategoryId   string
	Size         uint64
	Seeders      int
	Leachers     int
	Downloaded   int
	DownloadLink string
	IsMagnet     bool
	Announce     string
	Publisher    string
	AltName      string
}

func (t Torrent) AddedOnStr() interface{} {
	tm := time.Unix(t.AddedOn, 0)
	return tm.String()
}

type TorrentCategory struct {
	CategoryId   string
	CategoryName string
}
