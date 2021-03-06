package cache

import (
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"

	"github.com/sp0x/torrentd/indexer/cache/mocks"
	mocks2 "github.com/sp0x/torrentd/indexer/source/mocks"
)

func TestConnectivityCache_IsOk(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	fetcher := mocks2.NewMockContentFetcher(ctrl)
	conCache, _ := NewConnectivityCache(fetcher)
	exampleURL, _ := url.Parse("http://example.com")
	g.Expect(conCache.IsOk(exampleURL)).To(gomega.BeFalse())
	br := &mocks.MockedBrowser{}
	br.CanOpen = true
	g.Expect(conCache.IsOkAndSet(exampleURL, func() bool {
		err := conCache.Test(exampleURL)
		return err == nil
	})).To(gomega.BeTrue())

	g.Expect(conCache.IsOk(exampleURL)).To(gomega.BeTrue())
}
