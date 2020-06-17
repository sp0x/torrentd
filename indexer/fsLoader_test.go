package indexer

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestNewFsLoader(t *testing.T) {
	g := NewWithT(t)
	loader := NewFsLoader("appx")
	g.Expect(len(loader.Directories)).To(Equal(2))
}
