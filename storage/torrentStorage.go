package storage

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/sp0x/rutracker-rss/db"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"strconv"
	"time"
)

type DBStorage struct {
	Path string
}

func (ts *DBStorage) FindByTorrentId(id string) *search.ExternalResultItem {
	gdb := db.GetOrmDb(ts.Path)
	defer gdb.Close()
	var torrent search.ExternalResultItem
	if gdb.First(&torrent, &search.ExternalResultItem{LocalId: id}).RowsAffected == 0 {
		return nil
	}
	return &torrent
}

func (ts *DBStorage) Create(tr *search.ExternalResultItem) {
	gdb := db.GetOrmDb(ts.Path)
	defer func() {
		_ = gdb.Close()
	}()
	gdb.Create(tr)
}

func (ts *DBStorage) Truncate() {
	gdb := db.GetOrmDb(ts.Path)
	defer gdb.Close()
	gdb.Unscoped().Delete(&search.ExternalResultItem{})
}

func (ts *DBStorage) GetLatest(cnt int) []search.ExternalResultItem {
	gdb := db.GetOrmDb(ts.Path)
	defer gdb.Close()
	var items []search.ExternalResultItem
	gdb.Model(&search.ExternalResultItem{}).Find(&items).Order("added_on").Limit(cnt)
	return items
}

func (ts *DBStorage) GetTorrentCount() int64 {
	gdb := db.GetOrmDb(ts.Path)
	defer gdb.Close()
	var result int64
	gdb.Model(&search.ExternalResultItem{}).Count(&result)
	return result
}

func (ts *DBStorage) GetCategories() []db.TorrentCategory {
	gdb := db.GetOrmDb(ts.Path)
	defer func() {
		_ = gdb.Close()
	}()
	var categories []db.TorrentCategory
	var rawCats []search.ExternalResultItem
	if gdb.Model(&search.ExternalResultItem{}).Group("local_category_id").
		Scan(&rawCats).RowsAffected == 0 {
		return nil
	}
	for _, rc := range rawCats {
		categories = append(categories, db.TorrentCategory{
			CategoryId:   strconv.Itoa(rc.Category),
			CategoryName: rc.LocalCategoryName,
		})
	}
	return categories
}

func (ts *DBStorage) UpdateTorrent(id uint, torrent *search.ExternalResultItem) {
	gdb := db.GetOrmDb(ts.Path)
	defer gdb.Close()
	gdb.Model(&search.ExternalResultItem{}).Where(id).Update(torrent)
}

func (ts *DBStorage) GetTorrentsInCategories(ids []int) []search.ExternalResultItem {
	gdb := db.GetOrmDb(ts.Path)
	defer gdb.Close()
	var torrents []search.ExternalResultItem
	gdb.Model(&search.ExternalResultItem{}).Where(" category_id IN (?)", ids).Order("added_on desc").Find(&torrents)
	return torrents
}

func (ts *DBStorage) GetOlderThanHours(h int) []search.ExternalResultItem {
	gdb := db.GetOrmDb(ts.Path)
	defer gdb.Close()
	var torrents []search.ExternalResultItem
	tm := time.Now().Unix() - int64(60)*int64(60)*int64(h)
	gdb.Model(&search.ExternalResultItem{}).
		Where(fmt.Sprintf("publish_date < %d", tm)).
		Find(&torrents)
	return torrents
}

func (ts *DBStorage) GetNewest(cnt int) []search.ExternalResultItem {
	gdb := db.GetOrmDb(ts.Path)
	defer gdb.Close()
	var torrents []search.ExternalResultItem
	gdb.Model(&search.ExternalResultItem{}).
		Order("publish_date desc").
		Limit(cnt).
		Find(&torrents)
	return torrents
}

func (ts *DBStorage) FindNameAndIndexer(title string, indexerSite string) *search.ExternalResultItem {
	gdb := db.GetOrmDb(ts.Path)
	defer func() {
		_ = gdb.Close()
	}()
	var torrent search.ExternalResultItem
	srch := &search.ExternalResultItem{}
	srch.Title = title
	srch.Site = indexerSite
	if gdb.First(&torrent, srch).RowsAffected == 0 {
		return nil
	}
	return &torrent
}

func (ts *DBStorage) GetDb() *gorm.DB {
	return db.GetOrmDb(ts.Path)
}

var defaultStorage = DBStorage{}

func DefaultStorage() *DBStorage {
	return &defaultStorage
}

func GetOlderThanHours(h int) []search.ExternalResultItem {
	return defaultStorage.GetOlderThanHours(h)
}

//Handles torrent discovery
func HandleTorrentDiscovery(item *search.ExternalResultItem) (bool, bool) {
	var existingTorrent *search.ExternalResultItem
	if item.LocalId != "" {
		existingTorrent = defaultStorage.FindByTorrentId(item.LocalId)
	} else {
		existingTorrent = defaultStorage.FindNameAndIndexer(item.Title, item.Site)
	}

	isNew := existingTorrent == nil || existingTorrent.PublishDate != item.PublishDate
	isUpdate := existingTorrent != nil && (existingTorrent.PublishDate != item.PublishDate)
	if isNew {
		if isUpdate && existingTorrent != nil {
			item.Fingerprint = existingTorrent.Fingerprint
			defaultStorage.UpdateTorrent(existingTorrent.ID, item)
		} else {
			item.Fingerprint = search.GetTorrentFingerprint(item)
			defaultStorage.Create(item)
		}
	}
	item.SetState(isNew, isUpdate)
	return isNew, isUpdate
}
