package sqlite

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/sp0x/torrentd/db"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage/indexing"
	"strconv"
	"strings"
	"time"
)

type DBStorage struct {
	Path string
}

//Create a parameterized SQL query array. The first element is the query, all the following elements are parameters.
func createQueryArray(query indexing.Query) []interface{} {
	var searchParts []string
	var searchValues []interface{}
	for _, key := range query.Keys() {
		value, _ := query.Get(key)
		searchParts = append(searchParts, fmt.Sprintf("%s = ?", key))
		searchValues = append(searchValues, value)
	}
	fullQuery := []interface{}{strings.Join(searchParts, " AND ")}
	fullQuery = append(fullQuery, searchValues...)
	return fullQuery
}

func (d *DBStorage) Find(query indexing.Query, matchingItem *search.ExternalResultItem) error {
	gdb := db.GetOrmDb(d.Path)
	defer func() {
		_ = gdb.Close()
	}()
	fullQuery := createQueryArray(query)
	if gdb.First(&matchingItem, fullQuery).RowsAffected == 0 {
		return nil
	}
	return nil
}

//Update a result with a matching keyParts.
func (d *DBStorage) Update(query indexing.Query, item *search.ExternalResultItem) error {
	gdb := db.GetOrmDb(d.Path)
	defer func() {
		_ = gdb.Close()
	}()
	fullQuery := createQueryArray(query)
	if gdb.Model(&search.ExternalResultItem{}).Where(fullQuery).
		Update(&item).
		Limit(1).RowsAffected == 0 {
		//Log maybe?
	}
	return nil
}

//Find a result by it's id
func (d *DBStorage) FindById(id string) *search.ExternalResultItem {
	gdb := db.GetOrmDb(d.Path)
	defer func() {
		_ = gdb.Close()
	}()
	var torrent search.ExternalResultItem
	if gdb.First(&torrent, &search.ExternalResultItem{LocalId: id}).RowsAffected == 0 {
		return nil
	}
	return &torrent
}

//Create a new result record.
//In sqlite we're not using the key parts.
func (d *DBStorage) Create(keyParts indexing.Key, tr *search.ExternalResultItem) error {
	gdb := db.GetOrmDb(d.Path)
	defer func() {
		_ = gdb.Close()
	}()
	gdb.Create(tr)
	return nil
}

func (d *DBStorage) Truncate() {
	gdb := db.GetOrmDb(d.Path)
	defer func() {
		_ = gdb.Close()
	}()
	gdb.Unscoped().Delete(&search.ExternalResultItem{})
}

func (d *DBStorage) GetLatest(cnt int) []search.ExternalResultItem {
	gdb := db.GetOrmDb(d.Path)
	defer func() {
		_ = gdb.Close()
	}()
	var items []search.ExternalResultItem
	gdb.Model(&search.ExternalResultItem{}).Find(&items).Order("added_on").Limit(cnt)
	return items
}

func (d *DBStorage) GetTorrentCount() int64 {
	gdb := db.GetOrmDb(d.Path)
	defer func() {
		_ = gdb.Close()
	}()
	var result int64
	gdb.Model(&search.ExternalResultItem{}).Count(&result)
	return result
}

func (d *DBStorage) GetCategories() []db.TorrentCategory {
	gdb := db.GetOrmDb(d.Path)
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

func (d *DBStorage) UpdateResult(id uint, torrent *search.ExternalResultItem) {
	gdb := db.GetOrmDb(d.Path)
	defer func() {
		_ = gdb.Close()
	}()
	gdb.Model(&search.ExternalResultItem{}).Where(id).Update(torrent)
}

func (d *DBStorage) GetTorrentsInCategories(ids []int) []search.ExternalResultItem {
	gdb := db.GetOrmDb(d.Path)
	defer func() {
		_ = gdb.Close()
	}()
	var torrents []search.ExternalResultItem
	gdb.Model(&search.ExternalResultItem{}).Where(" category_id IN (?)", ids).Order("added_on desc").Find(&torrents)
	return torrents
}

func (d *DBStorage) GetOlderThanHours(h int) []search.ExternalResultItem {
	gdb := db.GetOrmDb(d.Path)
	defer func() {
		_ = gdb.Close()
	}()
	var torrents []search.ExternalResultItem
	tm := time.Now().Unix() - int64(60)*int64(60)*int64(h)
	gdb.Model(&search.ExternalResultItem{}).
		Where(fmt.Sprintf("publish_date < %d", tm)).
		Find(&torrents)
	return torrents
}

//GetNewest gets the CNT latest results.
func (d *DBStorage) GetNewest(cnt int) []search.ExternalResultItem {
	gdb := db.GetOrmDb(d.Path)
	defer func() {
		_ = gdb.Close()
	}()
	var torrents []search.ExternalResultItem
	gdb.Model(&search.ExternalResultItem{}).
		Order("publish_date desc").
		Limit(cnt).
		Find(&torrents)
	return torrents
}

//FindByNameAndIndex finds an item by it's name and index.
func (d *DBStorage) FindByNameAndIndex(title string, indexerSite string) *search.ExternalResultItem {
	gdb := db.GetOrmDb(d.Path)
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

func (d *DBStorage) GetDb() *gorm.DB {
	return db.GetOrmDb(d.Path)
}

//var defaultStorage = DBStorage{}
//
////DefaultStorageBacking gets the default storage method for results.
//func DefaultStorageBacking() *DBStorage {
//	return &defaultStorage
//}

//GetOlderThanHours gets items that are at least H hours old.
//func GetOlderThanHours(h int) []search.ExternalResultItem {
//	return defaultStorage.GetOlderThanHours(h)
//}
