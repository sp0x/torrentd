package indexer

import (
	"github.com/onsi/gomega"
	"testing"
)

func TestIndexerDefinition_getSearchEntity(t *testing.T) {
	g := gomega.NewWithT(t)
	ixdef := &IndexerDefinition{}
	ixdef.Search.Key = []string{"time"}
	searchEntity := ixdef.getSearchEntity()
	g.Expect(searchEntity.IndexKey[0]).To(gomega.Equal("ExtraFields.time"))
	ixdef.Search.Key = []string{"LocalId"}
	searchEntity = ixdef.getSearchEntity()
	g.Expect(searchEntity.IndexKey[0]).To(gomega.Equal("LocalId"))
}
