package cache

import (
	"errors"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"

	"github.com/sp0x/torrentd/indexer/cache/mocks"
	"github.com/sp0x/torrentd/indexer/source"
	mocks2 "github.com/sp0x/torrentd/indexer/source/mocks"
)

var (
	optimisticURL, _  = url.Parse("http://example.com")
	optimisticURL2, _ = url.Parse("http://example2.com")
)

func Test_NewOptimisticConnectivityCache(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	fetcher := mocks2.NewMockContentFetcher(ctrl)
	conCache, _ := NewOptimisticConnectivityCache(fetcher)
	br := &mocks.MockedBrowser{}
	br.CanOpen = true

	g.Expect(conCache.IsOk(optimisticURL.String())).To(gomega.BeTrue())

	conCache.Invalidate(optimisticURL.String())

	g.Expect(conCache.IsOk(optimisticURL.String())).To(gomega.BeFalse())
}

func TestOptimisticCache_ShouldReturnTrueFromTheStart(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	fetcher := mocks2.NewMockContentFetcher(ctrl)
	conCache, _ := NewOptimisticConnectivityCache(fetcher)
	br := &mocks.MockedBrowser{}
	br.CanOpen = true

	g.Expect(conCache.IsOk(optimisticURL.String())).To(gomega.BeTrue())
	g.Expect(conCache.IsOk(optimisticURL.String())).To(gomega.BeTrue())
}

func Test_OptimisticCache_ShouldReturnFalseIfInvalidated(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	fetcher := mocks2.NewMockContentFetcher(ctrl)
	conCache, _ := NewOptimisticConnectivityCache(fetcher)
	br := &mocks.MockedBrowser{}
	br.CanOpen = true
	g.Expect(conCache.IsOk(optimisticURL.String())).To(gomega.BeTrue())
	conCache.Invalidate(optimisticURL.String())
	g.Expect(conCache.IsOk(optimisticURL.String())).To(gomega.BeFalse())

	// Second case
	conCache, _ = NewOptimisticConnectivityCache(fetcher)
	conCache.Invalidate(optimisticURL.String())
	g.Expect(conCache.IsOk(optimisticURL.String())).To(gomega.BeFalse())
	g.Expect(conCache.IsOk(optimisticURL2.String())).To(gomega.BeTrue())
}

func Test_OptimisticCache_ShouldTestUrlsOnceTheyWereInvalidated(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	fetcher := mocks2.NewMockContentFetcher(ctrl)
	conCache, _ := NewOptimisticConnectivityCache(fetcher)
	mockGetRequest(fetcher, optimisticURL.String())
	g.Expect(conCache.IsOk(optimisticURL.String())).To(gomega.BeTrue())

	conCache.Invalidate(optimisticURL.String())

	// After this point urls should be tested
	g.Expect(conCache.IsOk(optimisticURL.String())).To(gomega.BeFalse())
	// We will test the url, which would be OK, so the invalidation is removed.
	g.Expect(conCache.Test(optimisticURL.String())).To(gomega.BeNil())
	ok := conCache.IsOk(optimisticURL.String())
	g.Expect(ok).To(gomega.BeTrue())
}

func mockGetRequest(fetcher *mocks2.MockContentFetcher, destURL string) *gomock.Call {
	return fetcher.EXPECT().
		Open(OfRequest("get", destURL)).
		Return(&source.HTMLFetchResult{
			HTTPResult: source.HTTPResult{
				StatusCode: 200,
			},
			DOM: nil,
		}, nil)
}

func mockFailingGetRequest(fetcher *mocks2.MockContentFetcher, destURL string) *gomock.Call {
	return fetcher.EXPECT().
		Open(OfRequest("get", destURL)).
		Return(nil, errors.New("offline"))
}

func Test_OptimisticCache_ShouldWorkWithOkAndSet(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	fetcher := mocks2.NewMockContentFetcher(ctrl)
	conCache, _ := NewOptimisticConnectivityCache(fetcher)
	mockGetRequest(fetcher, optimisticURL.String())

	// Do this like that so it's locked.
	ok := conCache.IsOkAndSet(optimisticURL.String(), func() bool {
		err := conCache.Test(optimisticURL.String())
		return err == nil
	})
	g.Expect(ok).To(gomega.BeTrue())
}

func Test_OptimisticCache_InvalidationShouldWorkWithOkAndSet(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	fetcher := mocks2.NewMockContentFetcher(ctrl)
	conCache, _ := NewOptimisticConnectivityCache(fetcher)
	mockGetRequest(fetcher, optimisticURL.String())

	// Do this like that so it's locked.
	conCache.Invalidate(optimisticURL.String())
	ok := conCache.IsOkAndSet(optimisticURL.String(), func() bool {
		err := conCache.Test(optimisticURL.String())
		return err == nil
	})
	// We should be able to connect
	g.Expect(ok).To(gomega.BeTrue())
}

func Test_OptimisticCache_NonWorkingURLs_Should_BeInvalidated(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	fetcher := mocks2.NewMockContentFetcher(ctrl)
	conCache, _ := NewOptimisticConnectivityCache(fetcher)
	mockFailingGetRequest(fetcher, optimisticURL.String())

	err := conCache.Test(optimisticURL.String())

	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(conCache.IsOk(optimisticURL.String()))
}

func Test_OptimisticCache_BadUrlsShouldStayInvalidated(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	fetcher := mocks2.NewMockContentFetcher(ctrl)
	conCache, _ := NewOptimisticConnectivityCache(fetcher)
	requestMock := mockGetRequest(fetcher, optimisticURL.String()).
		Times(1)

	// Initially the url should seem like it's OK
	ok := conCache.IsOkAndSet(optimisticURL.String(), func() bool {
		err := conCache.Test(optimisticURL.String())
		return err == nil
	})
	// It should seem like we're able to connect
	g.Expect(ok).To(gomega.BeTrue())

	// But if we invalidate it
	conCache.Invalidate(optimisticURL.String())
	requestMock.Return(nil, errors.New("offline"))
	// mockFailingGetRequest(fetcher, optimisticURL.String()).Times(1)

	ok = conCache.IsOkAndSet(optimisticURL.String(), func() bool {
		err := conCache.Test(optimisticURL.String())
		g.Expect(err).ToNot(gomega.BeNil())
		return err == nil
	})
	g.Expect(ok).To(gomega.BeFalse())
}
