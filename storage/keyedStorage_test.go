package storage

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage/bolt"
	"github.com/sp0x/torrentd/storage/indexing"
)

func TestKeyedStorage_Add(t *testing.T) {
	g := NewWithT(t)
	bolts, _ := bolt.NewBoltDbStorage(tempfile(), &search.ScrapeResultItem{})
	// We'll use `a` as a primary key
	storage := NewBuilder().
		WithPK(indexing.NewKey("a")).
		BackedBy(bolts).
		WithRecord(&search.ScrapeResultItem{}).
		Build()
	// We'll also define an index `ix`
	storage.AddUniqueIndex(indexing.NewKey("ix"))
	item := &search.ScrapeResultItem{}

	item.ModelData = make(map[string]interface{})
	item.ModelData["a"] = "b"
	err := storage.Add(item)
	g.Expect(item.IsNew()).To(BeTrue())
	// Since we're using a custom key, UUIDValue should be nil
	g.Expect(item.UUIDValue != "").To(BeFalse())
	g.Expect(err).To(BeNil())

	// Shouldn't be able to add a new record since IX is a unique index and we'll be breaking that rule
	item = &search.ScrapeResultItem{}
	item.ModelData = make(map[string]interface{})
	item.ModelData["a"] = "bx"
	item.ModelData["c"] = "b"
	err = storage.Add(item)
	g.Expect(item.IsNew()).To(BeFalse())
	g.Expect(item.IsUpdate()).To(BeFalse())
	g.Expect(item.UUIDValue != "").To(BeFalse())
	g.Expect(err).ToNot(BeNil())

	// Should be able to add a new record with an unique IX
	item = &search.ScrapeResultItem{}
	item.ModelData = make(map[string]interface{})
	item.ModelData["a"] = "bx"
	item.ModelData["c"] = "b"
	item.ModelData["ix"] = "bbbb"
	err = storage.Add(item)
	g.Expect(item.IsNew()).To(BeTrue())
	g.Expect(item.IsUpdate()).To(BeFalse())
	g.Expect(item.UUIDValue != "").To(BeFalse())
	g.Expect(err).To(BeNil())

	// Should create a new item if the key field is not set.
	item = &search.ScrapeResultItem{}
	item.ModelData = make(map[string]interface{})
	item.ModelData["c"] = "b"
	item.ModelData["ix"] = "1"
	err = storage.Add(item)
	g.Expect(item.IsNew()).To(BeTrue())
	g.Expect(item.IsUpdate()).To(BeFalse())
	g.Expect(item.UUIDValue != "").To(BeTrue())
	g.Expect(err).To(BeNil())
}

func TestGetKeyNameFromQuery(t *testing.T) {
	g := NewWithT(t)
	query := indexing.NewQuery()
	query.Put("a", "b")
	name := indexing.GetIndexNameFromQuery(query)
	g.Expect(name).To(Equal("a"))
}

func TestGetIndexValueFromQuery(t *testing.T) {
	g := NewWithT(t)
	query := indexing.NewQuery()
	query.Put("a", "b")
	val := indexing.GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal("b"))
	// Should work with multiple query fields
	query = indexing.NewQuery()
	query.Put("a", "b")
	query.Put("d", "x")
	val = indexing.GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal("b\000x"))

	// Should work with ints
	query = indexing.NewQuery()
	query.Put("a", "b")
	query.Put("d", 3)
	val = indexing.GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal("b\0003"))

	// Should work with floats
	query = indexing.NewQuery()
	query.Put("a", "b")
	query.Put("d", 3.5)
	val = indexing.GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal("b\0003.5"))

	// Should work with dates
	tm := time.Now()
	query = indexing.NewQuery()
	query.Put("a", "b")
	query.Put("d", tm)
	val = indexing.GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal(fmt.Sprintf("b\000%v", tm.Unix())))

	// Should work with dates
	query = indexing.NewQuery()
	query.Put("a", "b")
	query.Put("d", true)
	val = indexing.GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal(fmt.Sprintf("b\000%v", true)))
}

func TestGetIndexValueFromItem(t *testing.T) {
	g := NewWithT(t)
	// Should use key values only, when generating the index value
	key := indexing.NewKey("a")
	item := &search.ScrapeResultItem{}
	item.ModelData = make(map[string]interface{})
	item.ModelData["a"] = "asdasd123"
	item.ModelData["ab"] = "2"
	item.ModelData["ax"] = time.Now()
	item.ModelData["55"] = time.Now().Unix()
	indexValue := indexing.GetIndexValueFromItem(key, item)
	g.Expect(string(indexValue)).To(Equal("asdasd123"))
}

func TestKeyedStorage_NewWithKey(t *testing.T) {
	g := NewWithT(t)
	bolts, _ := bolt.NewBoltDbStorage(tempfile(), &search.ScrapeResultItem{})
	storage := NewBuilder().
		WithPK(indexing.NewKey("a")).
		BackedBy(bolts).
		WithRecord(&search.ScrapeResultItem{}).
		Build()
	otherStorage := storage.NewWithKey(indexing.NewKey("keyb"))
	// The storage backing in the second storage should be the same as in the first one.
	g.Expect(otherStorage.(*KeyedStorage).backing).To(Equal(bolts))
}
