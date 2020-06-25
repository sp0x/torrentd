package indexing

import (
	"github.com/onsi/gomega"
	"github.com/sp0x/torrentd/indexer/search"
	"testing"
)

func TestGetIndexValueFromItem(t *testing.T) {
	g := gomega.NewWithT(t)
	item := &search.ExternalResultItem{}
	item.ExtraFields = make(map[string]interface{})
	item.ExtraFields["time"] = "33"
	k := NewKey("ExtraFields.time")
	val := GetIndexValueFromItem(k, item)
	g.Expect(val).ToNot(gomega.BeNil())
	g.Expect(string(val)).To(gomega.Equal("33"))
}
