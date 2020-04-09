package torrent

import "github.com/sp0x/rutracker-rss/db"

type Storage struct {
}

func (ts *Storage) FindByTorrentId(id string) *db.Torrent {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	var torrent db.Torrent
	if gdb.First(&torrent, &db.Torrent{TorrentId: id}).RowsAffected == 0 {
		return nil
	}
	return &torrent
}

func (ts *Storage) Create(tr *db.Torrent) {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	gdb.Create(tr)
}

func (ts *Storage) Truncate() {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	gdb.Unscoped().Delete(&db.Torrent{})
}

func (ts *Storage) GetLatest(cnt int) []db.Torrent {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	var items []db.Torrent
	gdb.Model(&db.Torrent{}).Find(&items).Order("added_on").Limit(cnt)
	return items
}

func (ts *Storage) GetTorrentCount() int64 {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	var result int64
	gdb.Model(&db.Torrent{}).Count(&result)
	return result
}

func (ts *Storage) GetCategories() []db.TorrentCategory {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	var categories []db.TorrentCategory
	gdb.Model(&db.Torrent{}).Select("category_name, category_id").Group("category_id").Scan(&categories)
	return categories
}

func (ts *Storage) UpdateTorrent(id uint, torrent *db.Torrent) {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	gdb.Model(&db.Torrent{}).Where(id).Update(torrent)
}

func (ts *Storage) GetTorrentsInCategories(ids []int) []db.Torrent {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	var torrents []db.Torrent
	gdb.Model(&db.Torrent{}).Where(" category_id IN (?)", ids).Order("added_on desc").Find(&torrents)
	return torrents
}

func (ts *Storage) GetOlderThanHours(h int) []db.Torrent {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	var torrents []db.Torrent
	gdb.Model(&db.Torrent{}).Where("added_on").Find(&torrents)
	return torrents
}
