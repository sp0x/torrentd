package indexing_test

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/sp0x/torrentd/bots"
	"github.com/sp0x/torrentd/indexer/search"
	. "github.com/sp0x/torrentd/storage/indexing"
)

func TestKeyHasValue(t *testing.T) {
	g := gomega.NewWithT(t)
	item := &search.ScrapeResultItem{}
	chat := &bots.Chat{}
	item.ModelData = make(map[string]interface{})
	item.ModelData["time"] = "33"
	k := NewKey("ExtraFields.time")
	g.Expect(KeyHasValue(k, item)).To(gomega.BeTrue())

	item = &search.ScrapeResultItem{}
	item.ModelData = make(map[string]interface{})
	item.LocalID = "33"
	k = NewKey("LocalID")
	g.Expect(KeyHasValue(k, item)).To(gomega.BeTrue())

	item = &search.ScrapeResultItem{}
	item.ModelData = make(map[string]interface{})
	item.ModelData["time"] = "33"
	k = NewKey("time")
	g.Expect(KeyHasValue(k, item)).To(gomega.BeTrue())

	item = &search.ScrapeResultItem{}
	item.ModelData = make(map[string]interface{})
	item.ModelData["time"] = ""
	k = NewKey("time")
	g.Expect(KeyHasValue(k, item)).ToNot(gomega.BeTrue())

	// Should work with other types also
	kChat := NewKey("id")
	g.Expect(KeyHasValue(kChat, chat)).ToNot(gomega.BeTrue())
}

func TestGetKeyQueryFromItem(t *testing.T) {
	g := gomega.NewWithT(t)
	item := &search.ScrapeResultItem{}
	item.ModelData = make(map[string]interface{})
	item.ModelData["time"] = "33"
	k := NewKey("ExtraFields.time")
	q := GetKeyQueryFromItem(k, item)
	g.Expect(q).ToNot(gomega.BeNil())
	g.Expect(q.Get("time")).To(gomega.BeNil())
	val, found := q.Get("ExtraFields.time")
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(val).To(gomega.Equal("33"))

	item = &search.ScrapeResultItem{}
	item.ModelData = make(map[string]interface{})
	item.LocalID = "34"
	k = NewKey("LocalID")
	q = GetKeyQueryFromItem(k, item)
	g.Expect(KeyHasValue(k, item)).To(gomega.BeTrue())
	val, found = q.Get("LocalID")
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(val).To(gomega.Equal("34"))
}

func TestKey_AddKeys(t *testing.T) {
	g := gomega.NewWithT(t)
	k := NewKey("a")
	k.AddKeys(NewKey("b"))
	k.AddKeys(NewKey("b", "c", "d"))
	g.Expect(k.IsEmpty()).To(gomega.BeFalse())
	g.Expect(len(k.Fields)).To(gomega.Equal(4))

	k = &Key{Fields: []string{"a"}}
	k.AddKeys(NewKey("b"))
	k.Add("b")
	k.Add("agg")
	g.Expect(k.IsEmpty()).To(gomega.BeFalse())
	g.Expect(len(k.Fields)).To(gomega.Equal(3))
}
