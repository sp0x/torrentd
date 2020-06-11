package storage

import (
	"github.com/onsi/gomega"
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
	defer func() {
		_ = gormDb.Close()
	}()
	gormDb.AutoMigrate(&search.ExternalResultItem{})
	gormDb.AutoMigrate(&db.TorrentCategory{})
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
	setup()
	g := gomega.NewGomegaWithT(t)
	type args struct {
		id string
	}
	tests := []struct {
		name       string
		args       args
		want       *search.ExternalResultItem
		wantNotNil bool
	}{
		{name: tempfile(), args: args{id: "1"}, wantNotNil: true},
	}
	storage.Create(&search.ExternalResultItem{
		LocalId:    "1",
		ResultItem: search.ResultItem{Title: "a"},
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := storage.FindByTorrentId(tt.args.id)
			if tt.wantNotNil {
				g.Expect(got).ShouldNot(gomega.BeNil())
			}
		})
	}
	shutdown()
}

func TestDBStorage_FindNameAndIndexer(t *testing.T) {
	setup()
	g := gomega.NewGomegaWithT(t)
	type args struct {
		title       string
		indexerSite string
	}
	tests := []struct {
		name       string
		args       args
		want       *search.ExternalResultItem
		wantNotNil bool
	}{
		{name: "a", args: args{
			title:       "a",
			indexerSite: "sitea",
		}, wantNotNil: true},
	}
	storage.Create(&search.ExternalResultItem{
		ResultItem: search.ResultItem{
			Site:  "sitea",
			Title: "a",
		},
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := storage.FindNameAndIndexer(tt.args.title, tt.args.indexerSite)
			if tt.wantNotNil {
				g.Expect(got).ShouldNot(gomega.BeNil())
			}
		})
	}
	shutdown()
}

//Temp ignore
//func TestDBStorage_GetCategories(t *testing.T) {
//	setup()
//	g := gomega.NewGomegaWithT(t)
//	tests := []struct {
//		name string
//		want []db.TorrentCategory
//	}{
//		{name: "a", want: []db.TorrentCategory{db.TorrentCategory{
//			CategoryId:   "12",
//			CategoryName: "Localcat",
//		}}},
//	}
//	storage.Create(&search.ExternalResultItem{
//		ResultItem: search.ResultItem{
//			Site:     "sitea",
//			Title:    "a",
//			Category: 12,
//		},
//		LocalCategoryName: "Localcat",
//	})
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got := storage.GetCategories()
//			g.Expect(got).Should(gomega.BeEquivalentTo(tt.want))
//		})
//	}
//	shutdown()
//}

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
