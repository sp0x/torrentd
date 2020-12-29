package cache

import (
	"errors"
	"sync"
	"time"

	"github.com/sp0x/surf/browser"
)

func NewOptimisticConnectivityCache() (*OptimisticConnectivityCache, error) {
	c := &OptimisticConnectivityCache{}
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
	browser browser.Browsable
	lock    sync.RWMutex
	// invalidatedCache   map[string]CacheInfo
	invalidatedCache LRUCache
}

// IsOk returns whether the invalidatedCache contains a successful response for the url
func (c *OptimisticConnectivityCache) IsOk(url string) bool {
	isInvalidated := c.invalidatedCache.Contains(url)
	return !isInvalidated
}

// Invalidate a invalidatedCache entry by removing it from the invalidatedCache.
func (c *OptimisticConnectivityCache) Invalidate(url string) {
	c.invalidatedCache.Add(url, CacheInfo{
		added: time.Now(),
	})
}

// IsOkAndSet checks if the `u` value is contained, if it's not it checks it.
// This operation is thread safe, you can use it to modify the invalidatedCache state in the function.
func (c *OptimisticConnectivityCache) IsOkAndSet(u string, f func() bool) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	var result bool
	invalidated := c.invalidatedCache.Contains(u)
	if !invalidated {
		return true
	} else {
		result = f()
	}
	return result
}

// Test the connectivity for an url.
func (c *OptimisticConnectivityCache) Test(u string) error {
	if c.browser == nil {
		return errors.New("connectivity invalidatedCache has no browser. call SetBrowser first")
	}
	err := c.browser.Open(u)
	// If the url can be opened, we remove the invalid state.
	if err == nil {
		c.invalidatedCache.Remove(u)
	}
	return err
}

func (c *OptimisticConnectivityCache) SetBrowser(bow browser.Browsable) {
	c.browser = bow
}

func (c *OptimisticConnectivityCache) ClearBrowser() {
	c.browser = nil
}
