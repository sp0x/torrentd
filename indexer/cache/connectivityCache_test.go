package cache

import (
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"

	"github.com/sp0x/torrentd/indexer/source"
	mocks2 "github.com/sp0x/torrentd/indexer/source/mocks"
)

func TestConnectivityCache_IsOk(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	fetcher := mocks2.NewMockContentFetcher(ctrl)
	conCache, _ := NewConnectivityCache(fetcher)
	exampleURL, _ := url.Parse("http://example.com")
	g.Expect(conCache.IsOk(exampleURL.String())).To(gomega.BeFalse())
	fetcher.EXPECT().
		Open(OfRequest("GET", exampleURL.String())).
		Return(&source.HTMLFetchResult{
			HTTPResult: source.HTTPResult{
				StatusCode: 200,
			},
			DOM: nil,
		}, nil)

	result := conCache.IsOkAndSet(exampleURL.String(), func() bool {
		err := conCache.Test(exampleURL.String())
		return err == nil
	})

	g.Expect(result).To(gomega.BeTrue())
	g.Expect(conCache.IsOk(exampleURL.String())).To(gomega.BeTrue())
}
