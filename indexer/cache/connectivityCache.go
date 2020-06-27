package cache

import (
	"errors"
	"github.com/sp0x/surf/browser"
	"sync"
	"time"
)

//ConnectivityCache is a invalidatedCache for URL connectivity.
type ConnectivityCache struct {
	browser browser.Browsable
	lock    sync.RWMutex
	//invalidatedCache   map[string]CacheInfo
	cache LRUCache
}

func NewConnectivityCache() (*ConnectivityCache, error) {
	c := ConnectivityCache{}
	//Connection statuses are kept for 60 minutes, we keep at most 10k urls
	cache, err := NewTTL(10000, time.Minute*60)
	if err != nil {
		return nil, err
	}
	c.cache = cache
	return &c, nil
}

//Invalidate a invalidatedCache entry by removing it from the invalidatedCache.
func (c *ConnectivityCache) Invalidate(url string) {
	c.cache.Remove(url)
}

//IsOk returns whether the invalidatedCache contains a successful response for the url
func (c *ConnectivityCache) IsOk(url string) bool {
	ok := c.cache.Contains(url)
	return ok
}

//IsOkAndSet checks if the `u` value is contained, if it's not it checks it.
//This operation is thread safe, you can use it to modify the invalidatedCache state in the function.
func (c *ConnectivityCache) IsOkAndSet(u string, f func() bool) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	var result bool
	contained := c.cache.Contains(u)
	if !contained {
		result = f()
	} else {
		result = contained
	}
	return result
}

//Test the connectivity for an url.
func (c *ConnectivityCache) Test(u string) error {
	if c.browser == nil {
		return errors.New("connectivity invalidatedCache has no browser. call SetBrowser first")
	}
	err := c.browser.Open(u)
	if err == nil {
		c.cache.Add(u, CacheInfo{
			added: time.Now(),
		})
	}
	return err
}

func (c *ConnectivityCache) SetBrowser(bow browser.Browsable) {
	c.browser = bow
}

func (c *ConnectivityCache) ClearBrowser() {
	c.browser = nil
}
