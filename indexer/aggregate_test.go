package indexer

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/sp0x/torrentd/config"
)

func TestGetDefinition(t *testing.T) {
	g := gomega.NewWithT(t)
	cfg := &config.ViperConfig{}
	indexLoader := CreateEmbeddedDefinitionSource([]string{"index1", "index2", "c.yml"}, func(key string) ([]byte, error) {
		if key == "index1" {
			return []byte("name: index1"), nil
		} else {
			return []byte("name: index2"), nil
		}
	})
	setConfiguredIndexLoader(indexLoader, cfg)
	indexScope := NewScope(indexLoader)

	indexes, err := indexScope.LookupAll(cfg, newIndexSelector("index1, index2"))

	g.Expect(err).To(gomega.BeNil())
	g.Expect(indexes).ToNot(gomega.BeEmpty())
	g.Expect(len(indexes)).To(gomega.Equal(2))
	g.Expect(indexes[0].GetDefinition().Name).To(gomega.Equal("index1"))
	g.Expect(indexes[1].GetDefinition().Name).To(gomega.Equal("index2"))
	g.Expect(indexes[0].GetDefinition()).ToNot(gomega.BeNil())
	g.Expect(indexes[1].GetDefinition()).ToNot(gomega.BeNil())
	g.Expect(indexes.Name()).To(gomega.Equal("index1,index2"))
	g.Expect(indexes).ToNot(gomega.BeNil())
}
