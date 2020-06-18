package storage

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/sp0x/torrentd/db"
	"github.com/sp0x/torrentd/indexer/search"
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

//GetNewest gets the CNT latest results.
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

//FindByNameAndIndex finds an item by it's name and index.
func (ts *DBStorage) FindByNameAndIndex(title string, indexerSite string) *search.ExternalResultItem {
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

//DefaultStorage gets the default storage method for results.
func DefaultStorage() *DBStorage {
	return &defaultStorage
}

//GetOlderThanHours gets items that are at least H hours old.
func GetOlderThanHours(h int) []search.ExternalResultItem {
	return defaultStorage.GetOlderThanHours(h)
}

//HandleResultDiscovery handles the discovery of the result, adding additional information like staleness state.
func HandleResultDiscovery(item *search.ExternalResultItem) (bool, bool) {
	var existingResult *search.ExternalResultItem
	if item.LocalId != "" {
		existingResult = defaultStorage.FindByTorrentId(item.LocalId)
	} else {
		existingResult = defaultStorage.FindByNameAndIndex(item.Title, item.Site)
	}

	isNew := existingResult == nil || existingResult.PublishDate != item.PublishDate
	isUpdate := existingResult != nil && (existingResult.PublishDate != item.PublishDate)
	if isNew {
		if isUpdate && existingResult != nil {
			item.Fingerprint = existingResult.Fingerprint
			defaultStorage.UpdateTorrent(existingResult.ID, item)
		} else {
			item.Fingerprint = search.GetResultFingerprint(item)
			defaultStorage.Create(item)
		}
	}
	//We set the result's state so it's known later on whenever it's used.
	item.SetState(isNew, isUpdate)
	return isNew, isUpdate
}
