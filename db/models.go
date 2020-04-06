package db

import "github.com/jinzhu/gorm"

type Torrent struct {
	gorm.Model
	Name         string
	TorrentId    string
	AddedOn      string
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
}

type TorrentCategory struct {
	CategoryId   string
	CategoryName string
}
