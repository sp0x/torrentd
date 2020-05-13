package storage

import (
	"github.com/sp0x/rutracker-rss/db"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"os"
	"reflect"
	"testing"
)

var storage *DBStorage

func setup() {
	storage = &DBStorage{Path: tempfile()}
	gormDb := storage.GetDb()
	defer gormDb.Close()
	gormDb.AutoMigrate(&search.ExternalResultItem{})
}

func shutdown() {
	storage.Truncate()
	_ = os.Remove(storage.Path)
}

func TestDBStorage_Create(t *testing.T) {
	setup()
	type args struct {
		tr *search.ExternalResultItem
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "", args: args{tr: &search.ExternalResultItem{
			ResultItem: search.ResultItem{
				Title: "a",
			},
		}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.Create(tt.args.tr)
		})
	}
	shutdown()
}

func TestDBStorage_FindByTorrentId(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want *search.ExternalResultItem
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &DBStorage{}
			if got := ts.FindByTorrentId(tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindByTorrentId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBStorage_FindNameAndIndexer(t *testing.T) {
	type args struct {
		title       string
		indexerSite string
	}
	tests := []struct {
		name string
		args args
		want *search.ExternalResultItem
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &DBStorage{}
			if got := ts.FindNameAndIndexer(tt.args.title, tt.args.indexerSite); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindNameAndIndexer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBStorage_GetCategories(t *testing.T) {
	tests := []struct {
		name string
		want []db.TorrentCategory
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &DBStorage{}
			if got := ts.GetCategories(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCategories() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBStorage_GetLatest(t *testing.T) {
	type args struct {
		cnt int
	}
	tests := []struct {
		name string
		args args
		want []search.ExternalResultItem
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &DBStorage{}
			if got := ts.GetLatest(tt.args.cnt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLatest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBStorage_GetNewest(t *testing.T) {
	type args struct {
		cnt int
	}
	tests := []struct {
		name string
		args args
		want []search.ExternalResultItem
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &DBStorage{}
			if got := ts.GetNewest(tt.args.cnt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNewest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBStorage_GetOlderThanHours(t *testing.T) {
	type args struct {
		h int
	}
	tests := []struct {
		name string
		args args
		want []search.ExternalResultItem
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &DBStorage{}
			if got := ts.GetOlderThanHours(tt.args.h); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOlderThanHours() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBStorage_GetTorrentCount(t *testing.T) {
	tests := []struct {
		name string
		want int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &DBStorage{}
			if got := ts.GetTorrentCount(); got != tt.want {
				t.Errorf("GetTorrentCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBStorage_GetTorrentsInCategories(t *testing.T) {
	type args struct {
		ids []int
	}
	tests := []struct {
		name string
		args args
		want []search.ExternalResultItem
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &DBStorage{}
			if got := ts.GetTorrentsInCategories(tt.args.ids); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTorrentsInCategories() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBStorage_Truncate(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//ts := &DBStorage{}
		})
	}
}

func TestDBStorage_UpdateTorrent(t *testing.T) {
	type args struct {
		id      uint
		torrent *search.ExternalResultItem
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//ts := &DBStorage{}
		})
	}
}

func TestDefaultStorage(t *testing.T) {
	tests := []struct {
		name string
		want *DBStorage
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DefaultStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOlderThanHours(t *testing.T) {
	type args struct {
		h int
	}
	tests := []struct {
		name string
		args args
		want []search.ExternalResultItem
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetOlderThanHours(tt.args.h); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOlderThanHours() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleTorrentDiscovery(t *testing.T) {
	type args struct {
		item *search.ExternalResultItem
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := HandleTorrentDiscovery(tt.args.item)
			if got != tt.want {
				t.Errorf("HandleTorrentDiscovery() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("HandleTorrentDiscovery() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
