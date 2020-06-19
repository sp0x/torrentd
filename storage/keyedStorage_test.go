package storage

import (
	. "github.com/onsi/gomega"
	"github.com/sp0x/torrentd/indexer/search"
	"testing"
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
}

func TestGetKeyNameFromQuery(t *testing.T) {
	g := NewWithT(t)
	query := Query{}
	query["a"] = "b"
	name := GetIndexNameFromQuery(query)
	g.Expect(name).To(Equal("a"))
}

func TestGetIndexValueFromQuery(t *testing.T) {
	g := NewWithT(t)
	query := Query{}
	query["a"] = "b"
	val := GetIndexValueFromQuery(query)
	g.Expect(val).To(Equal("b"))
}
