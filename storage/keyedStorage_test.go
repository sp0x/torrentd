package storage

import (
	"fmt"
	. "github.com/onsi/gomega"
	"github.com/sp0x/torrentd/indexer/search"
	"testing"
	"time"
)

func TestKeyedStorage_Add(t *testing.T) {
	g := NewWithT(t)
	bolts, _ := NewBoltStorage(tempfile())
	storage := NewKeyedStorageWithBacking(NewKey("a"), bolts)
	//storage := NewKeyedStorage(NewKey("a"))
	item := &search.ExternalResultItem{}
	item.ExtraFields = make(map[string]interface{})
	item.ExtraFields["a"] = "b"
	storage.Add(item)
	g.Expect(item.IsNew()).To(BeTrue())

	item = &search.ExternalResultItem{}
	item.ExtraFields = make(map[string]interface{})
	item.ExtraFields["a"] = "b"
	item.ExtraFields["c"] = "b"
	storage.Add(item)
	g.Expect(item.IsNew()).To(BeFalse())
	g.Expect(item.IsUpdate()).To(BeTrue())
}

func TestGetKeyNameFromQuery(t *testing.T) {
	g := NewWithT(t)
	query := NewQuery()
	query.Put("a", "b")
	name := GetIndexNameFromQuery(query)
	g.Expect(name).To(Equal("a"))
}

func TestGetIndexValueFromQuery(t *testing.T) {
	g := NewWithT(t)
	query := NewQuery()
	query.Put("a", "b")
	val := GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal("b"))
	//Should work with multiple query fields
	query = NewQuery()
	query.Put("a", "b")
	query.Put("d", "x")
	val = GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal("b\000x"))

	//Should work with ints
	query = NewQuery()
	query.Put("a", "b")
	query.Put("d", 3)
	val = GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal("b\0003"))

	//Should work with floats
	query = NewQuery()
	query.Put("a", "b")
	query.Put("d", 3.5)
	val = GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal("b\0003.5"))

	//Should work with dates
	tm := time.Now()
	query = NewQuery()
	query.Put("a", "b")
	query.Put("d", tm)
	val = GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal(fmt.Sprintf("b\000%v", tm.Unix())))

	//Should work with dates
	query = NewQuery()
	query.Put("a", "b")
	query.Put("d", true)
	val = GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal(fmt.Sprintf("b\000%v", true)))
}

func TestGetIndexValueFromItem(t *testing.T) {
	g := NewWithT(t)
	//Should use key values only, when generating the index value
	key := Key{}
	key = append(key, "a")
	item := &search.ExternalResultItem{}
	item.ExtraFields = make(map[string]interface{})
	item.ExtraFields["a"] = "asdasd123"
	item.ExtraFields["ab"] = "2"
	item.ExtraFields["ax"] = time.Now()
	item.ExtraFields["55"] = time.Now().Unix()
	indexValue := GetIndexValueFromItem(key, item)
	g.Expect(string(indexValue)).To(Equal("asdasd123"))
}

func TestKeyedStorage_NewWithKey(t *testing.T) {
	g := NewWithT(t)
	bolts, _ := NewBoltStorage(tempfile())
	storage := NewKeyedStorageWithBacking(NewKey("a"), bolts)
	otherStorage := storage.NewWithKey(NewKey("keyb"))
	//The storage backing in the second storage should be the same as in the first one.
	g.Expect(otherStorage.(*KeyedStorage).backing).To(Equal(bolts))
}
