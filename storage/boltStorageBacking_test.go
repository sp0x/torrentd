package storage

import (
	"fmt"
	"github.com/boltdb/bolt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/search"
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
		//Init db
		BeforeEach(func() {
			tempDb, err := GetBoltDb(tempfile())
			if err != nil {
				Fail(fmt.Sprintf("Couldn't open a db: %v", err))
				return
			}
			if tempDb == nil {
				Fail(fmt.Sprintf("Nil db: %v", err))
				return
			}
			db = tempDb
			bstore.Database = tempDb
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
			newchat := &Chat{"tester", "", 12}
			err := bstore.StoreChat(newchat)
			if err != nil {
				Fail("Couldn't store chat")
				return
			}
			chat, err := bstore.GetChat(12)
			if err != nil {
				Fail("couldn't get chat from storage")
				return
			}
			if chat.ChatId != newchat.ChatId || chat.Username != newchat.Username || chat.InitialText != newchat.InitialText {
				Fail("couldn't properly restore chat from storage")
				return
			}
		})

		It("should return nil if a chat isn't found", func() {
			chat, err := bstore.GetChat(12)
			if err != nil {
				Fail(fmt.Sprintf("Error fetching non-existing chat: %v", err))
				return
			}
			if chat != nil {
				Fail("chat was not nil")
			}
		})

		It("Should be able to iterate over chats", func() {
			c1, c2 := &Chat{"a", "", 1}, &Chat{"b", "", 2}
			_ = bstore.StoreChat(c1)
			_ = bstore.StoreChat(c2)
			cnt := 0
			err := bstore.ForChat(func(chat *Chat) {
				if chat.ChatId == c1.ChatId || chat.ChatId == c2.ChatId {
					cnt += 1
				}
			})
			if err != nil {
				Fail(fmt.Sprintf("failed iterating over chats: %v", err))
				return
			}
			if cnt != 2 {
				Fail("Couldn't correctly iterate over chats, got the wrong results.")
			}
		})

		It("Should be able to store multiple search results", func() {
			items := []search.ExternalResultItem{
				{ResultItem: search.ResultItem{
					Title: "a", Category: categories.CategoryBooks.ID, GUID: "a",
				}},
				{ResultItem: search.ResultItem{
					Title: "b", Category: categories.CategoryBooks.ID, GUID: "b",
				}},
			}
			err := bstore.StoreSearchResults(items)
			if err != nil {
				Fail(fmt.Sprintf("failed to save multiple search results: %v", err))
				return
			}
			itemsRestored, err := bstore.GetSearchResults(categories.CategoryBooks.ID)
			if err != nil {
				Fail(fmt.Sprintf("error while fetching stored items"))
				return
			}
			if len(itemsRestored) != len(items) {
				Fail("mismatch in restoring search results")
			}
		})

		It("Should be able to store multiple uncategorized search results", func() {
			items := []search.ExternalResultItem{
				{ResultItem: search.ResultItem{
					Title: "az", Category: -100, GUID: "ag",
				}},
				{ResultItem: search.ResultItem{
					Title: "bz", Category: -100, GUID: "bg",
				}},
			}
			err := bstore.StoreSearchResults(items)
			if err != nil {
				Fail(fmt.Sprintf("failed to save multiple search results: %v", err))
				return
			}
			itemsRestored, err := bstore.GetSearchResults(-100)
			if err != nil {
				Fail(fmt.Sprintf("error while fetching stored items"))
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
			got, err := NewBoltStorage(tempfile())
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
			ResultItem: search.ResultItem{Title: "a", GUID: "x"},
		}}, wantErr: false},
		{name: "1", args: args{item: search.ExternalResultItem{
			ResultItem: search.ResultItem{Title: "b", GUID: "y"},
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
