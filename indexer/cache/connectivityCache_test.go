package cache

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/sp0x/torrentd/indexer/cache/mocks"
)

func TestConnectivityCache_IsOk(t *testing.T) {
	g := gomega.NewWithT(t)
	conCache, _ := NewConnectivityCache()
	g.Expect(conCache.IsOk("http://example.com")).To(gomega.BeFalse())
	br := &mocks.MockedBrowser{}
	br.CanOpen = true
	conCache.SetBrowser(br)
	g.Expect(conCache.IsOkAndSet("http://example.com", func() bool {
		err := conCache.Test("http://example.com")
		return err == nil
	})).To(gomega.BeTrue())

	g.Expect(conCache.IsOk("http://example.com")).To(gomega.BeTrue())
}
