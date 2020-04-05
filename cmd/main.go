package main

import (
	"github.com/sp0x/rutracker-rss/db"
	"os"
)

func init() {
	gormDb := db.GetOrmDb()
	defer gormDb.Close()
	gormDb.AutoMigrate(&db.Torrent{})
}

func main() {
	getNewTorrents(os.Getenv("RSS_USER"), os.Getenv("RSS_PASS"))
}
