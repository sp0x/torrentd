package bolt_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/sp0x/torrentd/bots"
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/search"
	. "github.com/sp0x/torrentd/storage/bolt"
	"github.com/sp0x/torrentd/storage/indexing"
)

var _ = Describe("Bolt storage", func() {
	It("Should be able to open a db", func() {
		db, err := GetBoltDB(tempfile())
		if err != nil {
			Fail(fmt.Sprintf("Couldn't open a db: %v", err))
			return
		}
		if db == nil {
			Fail(fmt.Sprintf("Nil db: %v", err))
			return
		}
		defer func() {
			_ = db.Close()
		}()
	})
	Context("with a database", func() {
		var db *bolt.DB
		bstore := &Storage{}
		key := indexing.NewKey("ChatID")
		// Init db
		BeforeEach(func() {
			tmpBstore, err := NewBoltDbStorage(tempfile(), &bots.Chat{})
			if err != nil {
				Fail(fmt.Sprintf("Couldn't open a db: %v", err))
				return
			}
			if tmpBstore == nil {
				Fail(fmt.Sprintf("Nil db: %v", err))
				return
			}
			db = tmpBstore.Database
			bstore = tmpBstore
		})
		// Teardown db
		AfterEach(func() {
			if db != nil {
				err := bstore.Truncate()
				if err != nil {
					Fail("couldn't teardown bolt storage")
				}
			}
		})
		It("Should be able to store chats", func() {
			newchat := &bots.Chat{Username: "tester", InitialText: "", ChatID: 12}
			err := bstore.Create(newchat, key)
			if err != nil {
				Fail(fmt.Sprintf("Couldn't store chat: %v", err))
				return
			}
			chat := bots.Chat{}
			query := indexing.NewQuery()
			query.Put("ChatID", 12)
			err = bstore.Find(query, &chat)
			if err != nil {
				Fail(fmt.Sprintf("couldn't get chat from storage: %v", err))
				return
			}
			if chat.ChatID != newchat.ChatID || chat.Username != newchat.Username || chat.InitialText != newchat.InitialText {
				Fail("couldn't properly restore chat from storage")
				return
			}
		})

		It("shouldn't return nil if a chat isn't found", func() {
			chat := bots.Chat{ChatID: 12}
			query := indexing.NewQuery()
			query.Put("ChatID", 12)
			err := bstore.Find(query, &chat)
			if err == nil {
				Fail(fmt.Sprintf("Error fetching non-existing chat: %v", err))
				return
			}
		})

		It("Should be able to iterate over chats", func() {
			c1, c2 := &bots.Chat{Username: "a", InitialText: "", ChatID: 1}, &bots.Chat{Username: "b", InitialText: "", ChatID: 2}
			err := bstore.Create(c1, key)
			if err != nil {
				Fail("couldn't store chat 1")
			}
			err = bstore.Create(c2, key)
			if err != nil {
				Fail("couldn't store chat 2")
			}
			cnt := 0
			bstore.ForEach(func(obj search.Record) {
				chat := obj.(*bots.Chat)
				if chat.ChatID == c1.ChatID || chat.ChatID == c2.ChatID {
					cnt++
				}
			})
			if cnt != 2 {
				Fail("Couldn't correctly iterate over chats, got the wrong results.")
			}
		})

		It("Should be able to store multiple search results and fetch them by a category id", func() {
			items := []search.Record{
				&search.TorrentResultItem{
					Title: "a", Category: categories.CategoryBooks.ID,
				},
				&search.TorrentResultItem{
					Title: "b", Category: categories.CategoryBooks.ID,
				},
			}
			items[0].SetUUID("a")
			items[1].SetUUID("b")

			err := bstore.AddAll(items)
			if err != nil {
				Fail(fmt.Sprintf("failed to save multiple search results: %v", err))
				return
			}
			itemsRestored, err := bstore.GetSearchResults(categories.CategoryBooks.ID)
			if err != nil {
				Fail("error while fetching stored items")
				return
			}
			if len(itemsRestored) != len(items) {
				Fail("mismatch in restoring search results")
			}
		})

		It("Should be able to store multiple uncategorized search results", func() {
			items := []search.Record{
				&search.TorrentResultItem{
					Title: "az", Category: -100,
				},
				&search.TorrentResultItem{
					Title: "bz", Category: -100,
				},
			}
			items[0].SetUUID("ag")
			items[1].SetUUID("bg")
			err := bstore.AddAll(items)
			if err != nil {
				Fail(fmt.Sprintf("failed to save multiple search results: %v", err))
				return
			}
			itemsRestored, err := bstore.GetSearchResults(-100)
			if err != nil {
				Fail("error while fetching stored items")
				return
			}
			if len(itemsRestored) != len(items) {
				Fail("mismatch in restoring search results")
			}
		})
	})
})

// tempfile returns a temporary file path.
func tempfile() string {
	f, err := ioutil.TempFile("", "bolt-")
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		panic(err)
	}
	return f.Name()
}

func TestNewBoltStorage(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	tests := []struct {
		name    string
		want    *Storage
		wantErr bool
	}{
		{"", nil, false},
	}
	for _, tt := range tests {
		// Run as a subtest
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewBoltDbStorage(tempfile(), &bots.Chat{})
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
			g.Expect(got).ShouldNot(gomega.BeNil())
		})
	}
}

func Test_getItemKey(t *testing.T) {
	type args struct {
		item search.Record
	}
	g := gomega.NewGomegaWithT(t)
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
		notNil  bool
	}{
		{name: "1", args: args{&search.TorrentResultItem{Title: "a"}}, wantErr: false},
		{name: "1", args: args{&search.TorrentResultItem{Title: "b"}}, wantErr: false},
		{name: "1", args: args{&search.TorrentResultItem{Title: "a"}}, wantErr: true},
	}
	tests[0].args.item.SetUUID("x")
	tests[1].args.item.SetUUID("y")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPKValueFromRecord(tt.args.item)
			if tt.wantErr {
				g.Expect(err).ShouldNot(gomega.BeNil())
			} else if tt.notNil {
				g.Expect(got).ShouldNot(gomega.BeNil())
			}
		})
	}
}

func TestBoltStorage_GetBucket(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	storage, err := NewBoltDbStorage(tempfile(), &bots.Chat{})
	if err != nil {
		t.Fatal(err)
	}
	db := storage.Database
	tx, err := db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}
	g.Expect(storage.GetBucket(tx, "none")).To(gomega.BeNil())
	bucket, err := tx.CreateBucketIfNotExists([]byte("newbucket"))
	g.Expect(bucket).Should(gomega.BeNil())
	g.Expect(err).ToNot(gomega.BeNil())
	_ = tx.Rollback()
	tx, err = db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}
	g.Expect(storage.GetBucket(tx, "none")).To(gomega.BeNil())
	bucket, err = tx.CreateBucketIfNotExists([]byte("newbucket"))
	g.Expect(bucket).ToNot(gomega.BeNil())
	g.Expect(err).To(gomega.BeNil())
	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}
	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}
	g.Expect(storage.GetBucket(tx, "newbucket")).ToNot(gomega.BeNil())
}

func TestBoltStorage_Find(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	storage, err := NewBoltDbStorage(tempfile(), &bots.Chat{})
	if err != nil {
		t.Fatal(err)
	}
	item := &search.ScrapeResultItem{}
	item.ModelData = make(map[string]interface{})
	item.ModelData["a"] = "b"
	item.ModelData["c"] = "b"
	// We create an item that would be indexed only by UUIDValue
	err = storage.Create(item, nil)
	if err != nil {
		t.Fatal(err)
	}
	g.Expect(item.UUIDValue != "").To(gomega.BeTrue())

	query := indexing.NewQuery()
	searchResult := search.ScrapeResultItem{}

	// Should be able to find it by UUIDValue, since it's the ID.
	query.Put("UUID", item.UUIDValue)
	g.Expect(storage.Find(query, &searchResult)).To(gomega.BeNil())

	// Should not find an item that's not indexed in that way.
	query = indexing.NewQuery()
	query.Put("a", "b")
	g.Expect(storage.Find(query, &searchResult)).ToNot(gomega.BeNil())

	// Should be able to create a new item with a custom ID
	item = &search.ScrapeResultItem{}
	item.ModelData = make(map[string]interface{})
	item.ModelData["a"] = "b"
	item.ModelData["c"] = "b"
	err = storage.CreateWithID(indexing.NewKey("a"), item, nil)
	g.Expect(err).To(gomega.BeNil())
	// it shouldn't use the UUIDValue
	g.Expect(item.UUIDValue != "").To(gomega.BeFalse())
	// and find it after that, using that custom ID
	query = indexing.NewQuery()
	query.Put("a", "b")
	g.Expect(storage.Find(query, &searchResult)).To(gomega.BeNil())
	g.Expect(len(searchResult.ModelData)).To(gomega.Equal(2))

	// Should be able to create records by UUIDValue
	// and index them with another key field
	item = &search.ScrapeResultItem{}
	item.ModelData = make(map[string]interface{})
	item.ModelData["x"] = "b"
	item.ModelData["c"] = "b"
	err = storage.Create(item, indexing.NewKey("x")) // Create it with UUID
	g.Expect(err).To(gomega.BeNil())
	query = indexing.NewQuery()
	query.Put("x", "b") // We're indexed under UUID, but we can also use the `x` key.
	searchResult = search.ScrapeResultItem{}
	g.Expect(storage.Find(query, &searchResult)).To(gomega.BeNil())
	g.Expect(len(searchResult.ModelData)).To(gomega.Equal(2))

	// Should be able to update records with a custom key as an additional index
	query = indexing.NewQuery()
	query.Put("x", "b")
	updateItem := &search.ScrapeResultItem{}
	updateItem.ModelData = make(map[string]interface{})
	updateItem.ModelData["x"] = "b"
	updateItem.ModelData["c"] = "b"
	updateItem.ModelData["d"] = "ddb"
	searchResult = search.ScrapeResultItem{}
	g.Expect(storage.Update(query, updateItem)).To(gomega.BeNil())
	g.Expect(storage.Find(query, &searchResult)).To(gomega.BeNil())
	g.Expect(len(searchResult.ModelData)).To(gomega.Equal(3))
}
