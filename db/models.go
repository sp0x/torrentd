package db

import (
	"github.com/jinzhu/gorm"
)

type Torrent struct {
	gorm.Model
	Name         string
	TorrentId    string
	AddedOn      string // *time.Time
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
}

type TorrentCategory struct {
	CategoryId   string
	CategoryName string
}
