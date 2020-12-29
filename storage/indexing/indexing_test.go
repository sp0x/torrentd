package indexing

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/sp0x/torrentd/indexer/search"
)

func TestGetIndexValueFromItem(t *testing.T) {
	g := gomega.NewWithT(t)
	item := &search.ScrapeResultItem{}
	item.ModelData = make(map[string]interface{})
	item.ModelData["time"] = "33"
	k := NewKey("ExtraFields.time")
	val := GetIndexValueFromItem(k, item)
	g.Expect(val).ToNot(gomega.BeNil())
	g.Expect(string(val)).To(gomega.Equal("33"))
}
