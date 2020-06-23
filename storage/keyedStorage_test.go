package storage

import (
	"fmt"
	. "github.com/onsi/gomega"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage/bolt"
	"github.com/sp0x/torrentd/storage/indexing"
	"testing"
	"time"
)

func TestKeyedStorage_Add(t *testing.T) {
	g := NewWithT(t)
	bolts, _ := bolt.NewBoltStorage(tempfile())
	//We'll use `a` as a primary key
	storage := NewKeyedStorageWithBacking(indexing.NewKey("a"), bolts)
	//We'll also define an index `ix`
	storage.AddUniqueIndex(indexing.NewKey("ix"))
	item := &search.ExternalResultItem{}

	item.ExtraFields = make(map[string]interface{})
	item.ExtraFields["a"] = "b"
	err := storage.Add(item)
	g.Expect(item.IsNew()).To(BeTrue())
	//Since we're using a custom key, GUID should be nil
	g.Expect(item.GUID != "").To(BeFalse())
	g.Expect(err).To(BeNil())

	//Shouldn't be able to add a new record since IX is a unique index and we'll be breaking that rule
	item = &search.ExternalResultItem{}
	item.ExtraFields = make(map[string]interface{})
	item.ExtraFields["a"] = "bx"
	item.ExtraFields["c"] = "b"
	err = storage.Add(item)
	g.Expect(item.IsNew()).To(BeFalse())
	g.Expect(item.IsUpdate()).To(BeFalse())
	g.Expect(item.GUID != "").To(BeFalse())
	g.Expect(err).ToNot(BeNil())

	//Should be able to add a new record with an unique IX
	item = &search.ExternalResultItem{}
	item.ExtraFields = make(map[string]interface{})
	item.ExtraFields["a"] = "bx"
	item.ExtraFields["c"] = "b"
	item.ExtraFields["ix"] = "bbbb"
	err = storage.Add(item)
	g.Expect(item.IsNew()).To(BeTrue())
	g.Expect(item.IsUpdate()).To(BeFalse())
	g.Expect(item.GUID != "").To(BeFalse())
	g.Expect(err).To(BeNil())

	//Should create a new item if the key field is not set.
	item = &search.ExternalResultItem{}
	item.ExtraFields = make(map[string]interface{})
	item.ExtraFields["c"] = "b"
	item.ExtraFields["ix"] = "1"
	err = storage.Add(item)
	g.Expect(item.IsNew()).To(BeTrue())
	g.Expect(item.IsUpdate()).To(BeFalse())
	g.Expect(item.GUID != "").To(BeTrue())
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
	//Should work with multiple query fields
	query = indexing.NewQuery()
	query.Put("a", "b")
	query.Put("d", "x")
	val = indexing.GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal("b\000x"))

	//Should work with ints
	query = indexing.NewQuery()
	query.Put("a", "b")
	query.Put("d", 3)
	val = indexing.GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal("b\0003"))

	//Should work with floats
	query = indexing.NewQuery()
	query.Put("a", "b")
	query.Put("d", 3.5)
	val = indexing.GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal("b\0003.5"))

	//Should work with dates
	tm := time.Now()
	query = indexing.NewQuery()
	query.Put("a", "b")
	query.Put("d", tm)
	val = indexing.GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal(fmt.Sprintf("b\000%v", tm.Unix())))

	//Should work with dates
	query = indexing.NewQuery()
	query.Put("a", "b")
	query.Put("d", true)
	val = indexing.GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal(fmt.Sprintf("b\000%v", true)))
}

func TestGetIndexValueFromItem(t *testing.T) {
	g := NewWithT(t)
	//Should use key values only, when generating the index value
	key := indexing.NewKey("a")
	item := &search.ExternalResultItem{}
	item.ExtraFields = make(map[string]interface{})
	item.ExtraFields["a"] = "asdasd123"
	item.ExtraFields["ab"] = "2"
	item.ExtraFields["ax"] = time.Now()
	item.ExtraFields["55"] = time.Now().Unix()
	indexValue := indexing.GetIndexValueFromItem(key, item)
	g.Expect(string(indexValue)).To(Equal("asdasd123"))
}

func TestKeyedStorage_NewWithKey(t *testing.T) {
	g := NewWithT(t)
	bolts, _ := bolt.NewBoltStorage(tempfile())
	storage := NewKeyedStorageWithBacking(indexing.NewKey("a"), bolts)
	otherStorage := storage.NewWithKey(indexing.NewKey("keyb"))
	//The storage backing in the second storage should be the same as in the first one.
	g.Expect(otherStorage.(*KeyedStorage).backing).To(Equal(bolts))
}
