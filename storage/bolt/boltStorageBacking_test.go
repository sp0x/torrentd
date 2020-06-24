package bolt

import (
	"fmt"
	"github.com/boltdb/bolt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage/indexing"
	"io/ioutil"
	"os"
	"testing"
)

var _ = Describe("Bolt storage", func() {

	It("Should be able to open a db", func() {
		db, err := GetBoltDb(tempfile())
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
		bstore := &BoltStorage{}
		key := indexing.NewKey("ChatId")
		//Init db
		BeforeEach(func() {
			tmpBstore, err := NewBoltStorage(tempfile(), &Chat{})
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
		//Teardown db
		AfterEach(func() {
			if db != nil {
				err := bstore.Truncate()
				if err != nil {
					Fail("couldn't teardown bolt storage")
				}
			}
		})
		It("Should be able to store chats", func() {
			newchat := &Chat{Username: "tester", InitialText: "", ChatId: 12}
			err := bstore.Create(newchat, key)
			if err != nil {
				Fail(fmt.Sprintf("Couldn't store chat: %v", err))
				return
			}
			chat := Chat{}
			query := indexing.NewQuery()
			query.Put("ChatId", 12)
			err = bstore.Find(query, &chat)
			if err != nil {
				Fail(fmt.Sprintf("couldn't get chat from storage: %v", err))
				return
			}
			if chat.ChatId != newchat.ChatId || chat.Username != newchat.Username || chat.InitialText != newchat.InitialText {
				Fail("couldn't properly restore chat from storage")
				return
			}
		})

		It("shouldn't return nil if a chat isn't found", func() {
			chat := Chat{ChatId: 12}
			query := indexing.NewQuery()
			query.Put("ChatId", 12)
			err := bstore.Find(query, &chat)
			if err == nil {
				Fail(fmt.Sprintf("Error fetching non-existing chat: %v", err))
				return
			}
		})

		It("Should be able to iterate over chats", func() {
			c1, c2 := &Chat{Username: "a", InitialText: "", ChatId: 1}, &Chat{Username: "b", InitialText: "", ChatId: 2}
			err := bstore.Create(c1, key)
			if err != nil {
				Fail("couldn't store chat 1")
			}
			err = bstore.Create(c2, key)
			if err != nil {
				Fail("couldn't store chat 2")
			}
			cnt := 0
			bstore.ForEach(func(obj interface{}) {
				chat := obj.(*Chat)
				if chat.ChatId == c1.ChatId || chat.ChatId == c2.ChatId {
					cnt += 1
				}
			})
			if cnt != 2 {
				Fail("Couldn't correctly iterate over chats, got the wrong results.")
			}
		})

		It("Should be able to store multiple search results", func() {
			items := []search.ExternalResultItem{
				{ResultItem: search.ResultItem{
					Title: "a", Category: categories.CategoryBooks.ID, UUIDValue: "a",
				}},
				{ResultItem: search.ResultItem{
					Title: "b", Category: categories.CategoryBooks.ID, UUIDValue: "b",
				}},
			}
			err := bstore.StoreSearchResults(items)
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
			items := []search.ExternalResultItem{
				{ResultItem: search.ResultItem{
					Title: "az", Category: -100, UUIDValue: "ag",
				}},
				{ResultItem: search.ResultItem{
					Title: "bz", Category: -100, UUIDValue: "bg",
				}},
			}
			err := bstore.StoreSearchResults(items)
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
	g := NewGomegaWithT(t)
	tests := []struct {
		name    string
		want    *BoltStorage
		wantErr bool
	}{
		{"", nil, false},
	}
	for _, tt := range tests {
		//Run as a subtest
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewBoltStorage(tempfile(), &Chat{})
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(got).ShouldNot(BeNil())
		})
	}
}

func Test_getItemKey(t *testing.T) {
	type args struct {
		item search.ExternalResultItem
	}
	g := NewGomegaWithT(t)
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
		notNil  bool
	}{
		{name: "1", args: args{item: search.ExternalResultItem{
			ResultItem: search.ResultItem{Title: "a", UUIDValue: "x"},
		}}, wantErr: false},
		{name: "1", args: args{item: search.ExternalResultItem{
			ResultItem: search.ResultItem{Title: "b", UUIDValue: "y"},
		}}, wantErr: false},
		{name: "1", args: args{item: search.ExternalResultItem{
			ResultItem: search.ResultItem{Title: "a"},
		}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getItemKey(tt.args.item)
			if tt.wantErr {
				g.Expect(err).ShouldNot(BeNil())
			} else {
				if tt.notNil {
					g.Expect(got).ShouldNot(BeNil())
				}

			}
		})
	}
}

func TestBoltStorage_GetBucket(t *testing.T) {
	g := NewGomegaWithT(t)
	storage, err := NewBoltStorage(tempfile(), &Chat{})
	if err != nil {
		t.Fatal(err)
	}
	db := storage.Database
	tx, err := db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}
	g.Expect(storage.GetBucket(tx, "none")).To(BeNil())
	bucket, err := tx.CreateBucketIfNotExists([]byte("newbucket"))
	g.Expect(bucket).Should(BeNil())
	g.Expect(err).ToNot(BeNil())
	_ = tx.Rollback()
	tx, err = db.Begin(true)
	if err != nil {
		t.Fatal(err)
	}
	g.Expect(storage.GetBucket(tx, "none")).To(BeNil())
	bucket, err = tx.CreateBucketIfNotExists([]byte("newbucket"))
	g.Expect(bucket).ToNot(BeNil())
	g.Expect(err).To(BeNil())
	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}
	tx, err = db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}
	g.Expect(storage.GetBucket(tx, "newbucket")).ToNot(BeNil())
}

func TestBoltStorage_Find(t *testing.T) {
	g := NewGomegaWithT(t)
	storage, err := NewBoltStorage(tempfile(), &Chat{})
	if err != nil {
		t.Fatal(err)
	}
	item := &search.ExternalResultItem{}
	item.ExtraFields = make(map[string]interface{})
	item.ExtraFields["a"] = "b"
	item.ExtraFields["c"] = "b"
	//We create an item that would be indexed only by UUIDValue
	err = storage.Create(item, nil)
	if err != nil {
		t.Fatal(err)
	}
	g.Expect(item.UUIDValue != "").To(BeTrue())

	query := indexing.NewQuery()
	searchResult := search.ExternalResultItem{}

	//Should be able to find it by UUIDValue, since it's the ID.
	query.Put("UUID", item.UUIDValue)
	g.Expect(storage.Find(query, &searchResult)).To(BeNil())

	//Should not find an item that's not indexed in that way.
	query = indexing.NewQuery()
	query.Put("a", "b")
	g.Expect(storage.Find(query, &searchResult)).ToNot(BeNil())

	//Should be able to create a new item with a custom ID
	item = &search.ExternalResultItem{}
	item.ExtraFields = make(map[string]interface{})
	item.ExtraFields["a"] = "b"
	item.ExtraFields["c"] = "b"
	err = storage.CreateWithId(indexing.NewKey("a"), item, nil)
	g.Expect(err).To(BeNil())
	//it shouldn't use the UUIDValue
	g.Expect(item.UUIDValue != "").To(BeFalse())
	//and find it after that, using that custom ID
	query = indexing.NewQuery()
	query.Put("a", "b")
	g.Expect(storage.Find(query, &searchResult)).To(BeNil())
	g.Expect(len(searchResult.ExtraFields)).To(Equal(2))

	//Should be able to create records by UUIDValue
	//and index them with another key field
	item = &search.ExternalResultItem{}
	item.ExtraFields = make(map[string]interface{})
	item.ExtraFields["x"] = "b"
	item.ExtraFields["c"] = "b"
	err = storage.Create(item, indexing.NewKey("x")) // Create it with UUID
	g.Expect(err).To(BeNil())
	query = indexing.NewQuery()
	query.Put("x", "b") //We're indexed under UUID, but we can also use the `x` key.
	searchResult = search.ExternalResultItem{}
	g.Expect(storage.Find(query, &searchResult)).To(BeNil())
	g.Expect(len(searchResult.ExtraFields)).To(Equal(2))

	//Should be able to update records with a custom key as an additional index
	query = indexing.NewQuery()
	query.Put("x", "b")
	updateItem := &search.ExternalResultItem{}
	updateItem.ExtraFields = make(map[string]interface{})
	updateItem.ExtraFields["x"] = "b"
	updateItem.ExtraFields["c"] = "b"
	updateItem.ExtraFields["d"] = "ddb"
	searchResult = search.ExternalResultItem{}
	g.Expect(storage.Update(query, updateItem)).To(BeNil())
	g.Expect(storage.Find(query, &searchResult)).To(BeNil())
	g.Expect(len(searchResult.ExtraFields)).To(Equal(3))
}
