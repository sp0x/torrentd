package cache

import (
	"errors"
	"github.com/sp0x/torrentd/indexer/source"
	"net/url"
	"sync"
	"time"
)

func NewOptimisticConnectivityCache(fetcher source.ContentFetcher) (*OptimisticConnectivityCache, error) {
	c := &OptimisticConnectivityCache{}
	c.fetcher = fetcher
	cache, err := NewThreadSafeWithEvict(10000, nil)
	if err != nil {
		return nil, err
	}
	c.invalidatedCache = cache
	return c, nil
}

/**
This invalidatedCache should return true from the start, and only start working if the items have been non-present.
*/
type OptimisticConnectivityCache struct {
	fetcher source.ContentFetcher
	lock    sync.RWMutex
	// invalidatedCache   map[string]Details
	invalidatedCache LRUCache
}

// IsOk returns whether the invalidatedCache contains a successful response for the url
func (c *OptimisticConnectivityCache) IsOk(url *url.URL) bool {
	isInvalidated := c.invalidatedCache.Contains(url)
	return !isInvalidated
}

// Invalidate a invalidatedCache entry by removing it from the invalidatedCache.
func (c *OptimisticConnectivityCache) Invalidate(url string) {
	c.invalidatedCache.Add(url, Details{
		added: time.Now(),
	})
}

// IsOkAndSet checks if the `u` value is contained, if it's not it checks it.
// This operation is thread safe, you can use it to modify the invalidatedCache state in the function.
func (c *OptimisticConnectivityCache) IsOkAndSet(u *url.URL, f func() bool) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	var result bool
	invalidated := c.invalidatedCache.Contains(u)
	if !invalidated {
		return true
	}
	result = f()
	return result
}

// Test the connectivity for an url.
func (c *OptimisticConnectivityCache) Test(testURL *url.URL) error {
	if c.fetcher == nil {
		return errors.New("connectivity invalidatedCache has no browser. call SetBrowser first")
	}
	result, err := c.fetcher.Open(&source.RequestOptions{URL: testURL})
	if err == nil {
		err = validateBrowserCall(result)
	}
	// If the url can be opened, we remove the invalid state.
	if err == nil {
		c.invalidatedCache.Remove(testURL.String())
	}
	return err
}

func (c *OptimisticConnectivityCache) SetBrowser(fetcher source.ContentFetcher) {
	c.fetcher = fetcher
}
