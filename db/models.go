package db

import "github.com/jinzhu/gorm"

type Torrent struct {
	gorm.Model
	Name        string
	TorrentId   string
	AddedOn     string
	Link        string
	Fingerprint string
}
