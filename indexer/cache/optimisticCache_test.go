package cache

import (
	"github.com/golang/mock/gomock"
	mocks2 "github.com/sp0x/torrentd/indexer/source/mocks"
	"net/url"
	"testing"

	"github.com/onsi/gomega"

	"github.com/sp0x/torrentd/indexer/cache/mocks"
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

	g.Expect(conCache.IsOk(optimisticURL)).To(gomega.BeTrue())

	conCache.Invalidate(optimisticURL.String())

	g.Expect(conCache.IsOk(optimisticURL)).To(gomega.BeFalse())
}

func TestOptimisticCache_ShouldReturnTrueFromTheStart(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	fetcher := mocks2.NewMockContentFetcher(ctrl)
	conCache, _ := NewOptimisticConnectivityCache(fetcher)
	br := &mocks.MockedBrowser{}
	br.CanOpen = true

	g.Expect(conCache.IsOk(optimisticURL)).To(gomega.BeTrue())
	g.Expect(conCache.IsOk(optimisticURL)).To(gomega.BeTrue())
}

func Test_OptimisticCache_ShouldReturnFalseIfInvalidated(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	fetcher := mocks2.NewMockContentFetcher(ctrl)
	conCache, _ := NewOptimisticConnectivityCache(fetcher)
	br := &mocks.MockedBrowser{}
	br.CanOpen = true
	g.Expect(conCache.IsOk(optimisticURL)).To(gomega.BeTrue())
	conCache.Invalidate(optimisticURL.String())
	g.Expect(conCache.IsOk(optimisticURL)).To(gomega.BeFalse())

	// Second case
	conCache, _ = NewOptimisticConnectivityCache(fetcher)
	conCache.Invalidate(optimisticURL.String())
	g.Expect(conCache.IsOk(optimisticURL)).To(gomega.BeFalse())
	g.Expect(conCache.IsOk(optimisticURL2)).To(gomega.BeTrue())
}

func Test_OptimisticCache_ShouldTestUrlsOnceTheyWereInvalidated(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	fetcher := mocks2.NewMockContentFetcher(ctrl)
	conCache, _ := NewOptimisticConnectivityCache(fetcher)
	br := &mocks.MockedBrowser{}
	br.CanOpen = true
	g.Expect(conCache.IsOk(optimisticURL)).To(gomega.BeTrue())

	conCache.Invalidate(optimisticURL.String())

	// After this point urls should be tested
	g.Expect(conCache.IsOk(optimisticURL)).To(gomega.BeFalse())
	// We will test the url, which would be OK, so the invalidation is removed.
	g.Expect(conCache.Test(optimisticURL)).To(gomega.BeNil())
	g.Expect(conCache.IsOk(optimisticURL)).To(gomega.BeTrue())
}

func Test_OptimisticCache_ShouldWorkWithOkAndSet(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	fetcher := mocks2.NewMockContentFetcher(ctrl)
	conCache, _ := NewOptimisticConnectivityCache(fetcher)
	br := &mocks.MockedBrowser{}
	br.CanOpen = true

	// Do this like that so it's locked.
	ok := conCache.IsOkAndSet(optimisticURL, func() bool {
		err := conCache.Test(optimisticURL)
		return err == nil
	})
	g.Expect(ok).To(gomega.BeTrue())
}

func Test_OptimisticCache_InvalidationShouldWorkWithOkAndSet(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	fetcher := mocks2.NewMockContentFetcher(ctrl)
	conCache, _ := NewOptimisticConnectivityCache(fetcher)
	br := &mocks.MockedBrowser{}
	br.CanOpen = true

	// Do this like that so it's locked.
	conCache.Invalidate(optimisticURL.String())
	ok := conCache.IsOkAndSet(optimisticURL, func() bool {
		err := conCache.Test(optimisticURL)
		return err == nil
	})
	// We should be able to connect
	g.Expect(ok).To(gomega.BeTrue())
}

func Test_OptimisticCache_BadUrlsShouldBeInvalidated(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	fetcher := mocks2.NewMockContentFetcher(ctrl)
	conCache, _ := NewOptimisticConnectivityCache(fetcher)
	br := &mocks.MockedBrowser{}
	br.CanOpen = false

	err := conCache.Test(optimisticURL)

	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(conCache.IsOk(optimisticURL))
}

func Test_OptimisticCache_BadUrlsShouldStayInvalidated(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	fetcher := mocks2.NewMockContentFetcher(ctrl)
	conCache, _ := NewOptimisticConnectivityCache(fetcher)
	br := &mocks.MockedBrowser{}
	br.CanOpen = false

	// Initially the url should seem like it's OK
	ok := conCache.IsOkAndSet(optimisticURL, func() bool {
		err := conCache.Test(optimisticURL)
		return err == nil
	})
	// It should seem like we're able to connect
	g.Expect(ok).To(gomega.BeTrue())
	// But if we invalidate it
	conCache.Invalidate(optimisticURL.String())

	ok = conCache.IsOkAndSet(optimisticURL, func() bool {
		err := conCache.Test(optimisticURL)
		g.Expect(err).ToNot(gomega.BeNil())
		return err == nil
	})
	g.Expect(ok).To(gomega.BeFalse())
}
