package indexer

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/sp0x/torrentd/indexer/definitions"
)

func TestAssetLoader_List(t *testing.T) {
	g := NewWithT(t)
	ldr := CreateEmbeddedDefinitionSource([]string{"a", "b", "c.yml"}, func(key string) ([]byte, error) {
		return nil, nil
	})

	names, err := ldr.ListAvailableIndexes(nil)

	g.Expect(err).To(BeNil())
	g.Expect(len(names)).To(Equal(3))
	g.Expect(names[0]).To(Equal("a"))
	g.Expect(names[1]).To(Equal("b"))
	g.Expect(names[2]).To(Equal("c"))
}

func TestAssetLoader_Load(t *testing.T) {
	g := NewWithT(t)
	ldr := CreateEmbeddedDefinitionSource([]string{"a", "b", "rutracker.org.yaml"}, func(key string) ([]byte, error) {
		fullname := fmt.Sprintf("definitions/%s.yaml", key)
		data, err := definitions.GzipAsset(fullname)
		if err != nil {
			return nil, err
		}
		data, _ = definitions.UnzipData(data)
		return data, nil
	})
	definition, err := ldr.Load("rutracker.org")
	g.Expect(err).To(BeNil())
	g.Expect(definition).ToNot(BeNil())
	g.Expect(definition.Site).To(Equal("rutracker.org"))
	g.Expect(definition.Name).To(Equal("rutracker.org"))
}

func TestGetDefaultEmbeddedDefinitionSource(t *testing.T) {
	g := NewWithT(t)
	src := getDefaultEmbeddedDefinitionSource()
	names, err := src.ListAvailableIndexes(nil)
	g.Expect(err).To(BeNil())
	g.Expect(len(names)).To(Equal(64))
}

func TestAssetLoader_ShouldWorkWithCommaSelectors(t *testing.T) {
	g := NewWithT(t)
	src := getDefaultEmbeddedDefinitionSource()

	names, err := src.ListAvailableIndexes(newIndexSelector("zamunda,arenabg"))

	g.Expect(err).To(BeNil())
	g.Expect(len(names)).To(Equal(2))
}
