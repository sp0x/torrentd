package storage

import (
	"fmt"
	"github.com/sp0x/rutracker-rss/db"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"time"
)

type Storage struct {
}

func (ts *Storage) FindByTorrentId(id string) *search.ExternalResultItem {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	var torrent search.ExternalResultItem
	if gdb.First(&torrent, &search.ExternalResultItem{LocalId: id}).RowsAffected == 0 {
		return nil
	}
	return &torrent
}

func (ts *Storage) Create(tr *search.ExternalResultItem) {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	gdb.Create(tr)
}

func (ts *Storage) Truncate() {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	gdb.Unscoped().Delete(&search.ExternalResultItem{})
}

func (ts *Storage) GetLatest(cnt int) []search.ExternalResultItem {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	var items []search.ExternalResultItem
	gdb.Model(&search.ExternalResultItem{}).Find(&items).Order("added_on").Limit(cnt)
	return items
}

func (ts *Storage) GetTorrentCount() int64 {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	var result int64
	gdb.Model(&search.ExternalResultItem{}).Count(&result)
	return result
}

func (ts *Storage) GetCategories() []db.TorrentCategory {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	var categories []db.TorrentCategory
	gdb.Model(&search.ExternalResultItem{}).Select("category_name, category_id").Group("category_id").Scan(&categories)
	return categories
}

func (ts *Storage) UpdateTorrent(id uint, torrent *search.ExternalResultItem) {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	gdb.Model(&search.ExternalResultItem{}).Where(id).Update(torrent)
}

func (ts *Storage) GetTorrentsInCategories(ids []int) []search.ExternalResultItem {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	var torrents []search.ExternalResultItem
	gdb.Model(&search.ExternalResultItem{}).Where(" category_id IN (?)", ids).Order("added_on desc").Find(&torrents)
	return torrents
}

func (ts *Storage) GetOlderThanHours(h int) []search.ExternalResultItem {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	var torrents []search.ExternalResultItem
	tm := time.Now().Unix() - int64(60)*int64(60)*int64(h)
	gdb.Model(&search.ExternalResultItem{}).Where(fmt.Sprintf("added_on < %d", tm)).Find(&torrents)
	return torrents
}
