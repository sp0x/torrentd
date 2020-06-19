package storage

import (
	"github.com/onsi/gomega"
	"github.com/sp0x/torrentd/db"
	"github.com/sp0x/torrentd/indexer/search"
	"os"
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
			_ = storage.Create(nil, tt.args.tr)
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
	_ = storage.Create(nil, &search.ExternalResultItem{
		LocalId:    "1",
		ResultItem: search.ResultItem{Title: "a"},
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := storage.FindById(tt.args.id)
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
	_ = storage.Create(nil, &search.ExternalResultItem{
		ResultItem: search.ResultItem{
			Site:  "sitea",
			Title: "a",
		},
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := storage.FindByNameAndIndex(tt.args.title, tt.args.indexerSite)
			if tt.wantNotNil {
				g.Expect(got).ShouldNot(gomega.BeNil())
			}
		})
	}
	shutdown()
}

func TestDBStorage_GetCategories(t *testing.T) {
	setup()
	g := gomega.NewGomegaWithT(t)
	tests := []struct {
		name string
		want []db.TorrentCategory
	}{
		{name: "a", want: []db.TorrentCategory{{
			CategoryId:   "12",
			CategoryName: "Localcat",
		}}},
	}
	_ = storage.Create(nil, &search.ExternalResultItem{
		ResultItem: search.ResultItem{
			Site:     "sitea",
			Title:    "a",
			Category: 12,
		},
		LocalCategoryName: "Localcat",
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := storage.GetCategories()
			g.Expect(got).Should(gomega.BeEquivalentTo(tt.want))
		})
	}
	shutdown()
}

func TestDBStorage_GetLatest(t *testing.T) {
}

func TestDBStorage_GetNewest(t *testing.T) {
}

func TestDBStorage_GetOlderThanHours(t *testing.T) {
}

func TestDBStorage_GetTorrentCount(t *testing.T) {
}

func TestDBStorage_GetTorrentsInCategories(t *testing.T) {
}

func TestDBStorage_Truncate(t *testing.T) {
}

func TestDBStorage_UpdateTorrent(t *testing.T) {
}

func TestDefaultStorage(t *testing.T) {
}

func TestGetOlderThanHours(t *testing.T) {
}

func TestHandleTorrentDiscovery(t *testing.T) {
}
