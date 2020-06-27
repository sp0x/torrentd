package cache

import (
	"github.com/onsi/gomega"
	"testing"
)

func TestConnectivityCache_IsOk(t *testing.T) {
	g := gomega.NewWithT(t)
	conCache, _ := NewConnectivityCache()
	g.Expect(conCache.IsOk("http://example.com")).To(gomega.BeFalse())
	br := &MockedBrowser{}
	br.CanOpen = true
	conCache.SetBrowser(br)
	g.Expect(conCache.IsOkAndSet("http://example.com", func() bool {
		err := conCache.Test("http://example.com")
		return err == nil
	})).To(gomega.BeTrue())

	g.Expect(conCache.IsOk("http://example.com")).To(gomega.BeTrue())
}
