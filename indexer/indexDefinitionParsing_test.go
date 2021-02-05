package indexer

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestIndexerDefinition_getSearchEntity(t *testing.T) {
	g := gomega.NewWithT(t)
	ixdef := &Definition{}
	ixdef.Search.Key = []string{"time"}
	searchEntity := ixdef.getSearchEntity()
	g.Expect(searchEntity.IndexKey[0]).To(gomega.Equal("ExtraFields.time"))
	ixdef.Search.Key = []string{"LocalID"}
	searchEntity = ixdef.getSearchEntity()
	g.Expect(searchEntity.IndexKey[0]).To(gomega.Equal("LocalID"))
}
