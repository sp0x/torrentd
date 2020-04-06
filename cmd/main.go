package main

import (
	"github.com/sp0x/rutracker-rss/db"
	"github.com/sp0x/rutracker-rss/torrent"
	"os"
)

func init() {
	gormDb := db.GetOrmDb()
	defer gormDb.Close()
	gormDb.AutoMigrate(&db.Torrent{})
}

func main() {
	torrent.GetNewTorrents(os.Getenv("RSS_USER"), os.Getenv("RSS_PASS"))
}
