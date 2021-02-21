package indexer

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestNewFsLoader(t *testing.T) {
	g := gomega.NewWithT(t)

	loader := NewFsLoader("appx")

	g.Expect(len(loader.Directories)).To(gomega.Equal(2))
}
