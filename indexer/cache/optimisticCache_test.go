package cache

import (
	"github.com/onsi/gomega"
	"testing"
)

var optimisticUrl = "http://example.com"
var optimisticUrl2 = "http://example2.com"

func Test_NewOptimisticConnectivityCache(t *testing.T) {
	g := gomega.NewWithT(t)
	conCache, _ := NewOptimisticConnectivityCache()
	br := &MockedBrowser{}
	br.CanOpen = true
	conCache.SetBrowser(br)
	g.Expect(conCache.IsOk(optimisticUrl)).To(gomega.BeTrue())
	conCache.Invalidate(optimisticUrl)
	g.Expect(conCache.IsOk(optimisticUrl)).To(gomega.BeFalse())
}

func TestOptimisticCache_ShouldReturnTrueFromTheStart(t *testing.T) {
	g := gomega.NewWithT(t)
	conCache, _ := NewOptimisticConnectivityCache()
	br := &MockedBrowser{}
	br.CanOpen = true
	conCache.SetBrowser(br)
	g.Expect(conCache.IsOk(optimisticUrl)).To(gomega.BeTrue())
	g.Expect(conCache.IsOk(optimisticUrl)).To(gomega.BeTrue())
}

func Test_OptimisticCache_ShouldReturnFalseIfInvalidated(t *testing.T) {
	g := gomega.NewWithT(t)
	conCache, _ := NewOptimisticConnectivityCache()
	br := &MockedBrowser{}
	br.CanOpen = true
	conCache.SetBrowser(br)
	g.Expect(conCache.IsOk(optimisticUrl)).To(gomega.BeTrue())
	conCache.Invalidate(optimisticUrl)
	g.Expect(conCache.IsOk(optimisticUrl)).To(gomega.BeFalse())

	//Second case
	conCache, _ = NewOptimisticConnectivityCache()
	conCache.Invalidate(optimisticUrl)
	g.Expect(conCache.IsOk(optimisticUrl)).To(gomega.BeFalse())
	g.Expect(conCache.IsOk(optimisticUrl2)).To(gomega.BeTrue())
}

func Test_OptimisticCache_ShouldTestUrlsOnceTheyWereInvalidated(t *testing.T) {
	g := gomega.NewWithT(t)
	conCache, _ := NewOptimisticConnectivityCache()
	br := &MockedBrowser{}
	br.CanOpen = true
	conCache.SetBrowser(br)
	g.Expect(conCache.IsOk(optimisticUrl)).To(gomega.BeTrue())
	conCache.Invalidate(optimisticUrl)
	//After this point urls should be tested
	g.Expect(conCache.IsOk(optimisticUrl)).To(gomega.BeFalse())
	//We will test the url, which would be OK, so the invalidation is removed.
	g.Expect(conCache.Test(optimisticUrl)).To(gomega.BeNil())
	g.Expect(conCache.IsOk(optimisticUrl)).To(gomega.BeTrue())
}

func Test_OptimisticCache_ShouldWorkWithOkAndSet(t *testing.T) {
	g := gomega.NewWithT(t)
	conCache, _ := NewOptimisticConnectivityCache()
	br := &MockedBrowser{}
	br.CanOpen = true
	conCache.SetBrowser(br)

	//Do this like that so it's locked.
	ok := conCache.IsOkAndSet(optimisticUrl, func() bool {
		err := conCache.Test(optimisticUrl)
		return err == nil
	})
	g.Expect(ok).To(gomega.BeTrue())
}

func Test_OptimisticCache_InvalidationShouldWorkWithOkAndSet(t *testing.T) {
	g := gomega.NewWithT(t)
	conCache, _ := NewOptimisticConnectivityCache()
	br := &MockedBrowser{}
	br.CanOpen = true
	conCache.SetBrowser(br)

	//Do this like that so it's locked.
	conCache.Invalidate(optimisticUrl)
	ok := conCache.IsOkAndSet(optimisticUrl, func() bool {
		err := conCache.Test(optimisticUrl)
		return err == nil
	})
	//We should be able to connect
	g.Expect(ok).To(gomega.BeTrue())
}

func Test_OptimisticCache_BadUrlsShouldBeInvalidated(t *testing.T) {
	g := gomega.NewWithT(t)
	conCache, _ := NewOptimisticConnectivityCache()
	br := &MockedBrowser{}
	br.CanOpen = false
	conCache.SetBrowser(br)
	err := conCache.Test(optimisticUrl)
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(conCache.IsOk(optimisticUrl))

}

func Test_OptimisticCache_BadUrlsShouldStayInvalidated(t *testing.T) {
	g := gomega.NewWithT(t)
	conCache, _ := NewOptimisticConnectivityCache()
	br := &MockedBrowser{}
	br.CanOpen = false
	conCache.SetBrowser(br)

	//Initially the url should seem like it's OK
	ok := conCache.IsOkAndSet(optimisticUrl, func() bool {
		err := conCache.Test(optimisticUrl)
		return err == nil
	})
	//It should seem like we're able to connect
	g.Expect(ok).To(gomega.BeTrue())
	//But if we invalidate it
	conCache.Invalidate(optimisticUrl)

	ok = conCache.IsOkAndSet(optimisticUrl, func() bool {
		err := conCache.Test(optimisticUrl)
		g.Expect(err).ToNot(gomega.BeNil())
		return err == nil
	})
	g.Expect(ok).To(gomega.BeFalse())
}
