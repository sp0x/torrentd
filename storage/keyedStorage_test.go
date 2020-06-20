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
	g.Expect(string(val)).To(Equal("b"))
	//Should work with multiple query fields
	query = Query{}
	query["a"] = "b"
	query["d"] = "x"
	val = GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal("b\000x"))

	//Should work with ints
	query = Query{}
	query["a"] = "b"
	query["d"] = 3
	val = GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal("b\0003"))

	//Should work with floats
	query = Query{}
	query["a"] = "b"
	query["d"] = 3.5
	val = GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal("b\0003.5"))

	//Should work with dates
	tm := time.Now()
	query = Query{}
	query["a"] = "b"
	query["d"] = tm
	val = GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal(fmt.Sprintf("b\000%v", tm.Unix())))

	//Should work with dates
	query = Query{}
	query["a"] = "b"
	query["d"] = true
	val = GetIndexValueFromQuery(query)
	g.Expect(string(val)).To(Equal(fmt.Sprintf("b\000%v", true)))
}
