package torrent

import "github.com/sp0x/rutracker-rss/db"

type TorrentStorage struct {
}

func (ts *TorrentStorage) FindByTorrentId(id string) *db.Torrent {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	var torrent db.Torrent
	if gdb.First(&torrent, &db.Torrent{TorrentId: id}).RowsAffected == 0 {
		return nil
	}
	return &torrent
}

func (ts *TorrentStorage) Create(tr *db.Torrent) {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	gdb.Create(tr)
}

func (ts *TorrentStorage) Truncate() {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	gdb.Unscoped().Delete(&db.Torrent{})
}
