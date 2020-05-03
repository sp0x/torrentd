package cache

import (
	"errors"
	"github.com/sp0x/surf/browser"
	"time"
)

type CacheInfo struct {
	added time.Time
}

type ConnectivityCache struct {
	browser *browser.Browser
	//cache   map[string]CacheInfo
	cache LRUCache
}

func NewConnectivityCache() (*ConnectivityCache, error) {
	c := ConnectivityCache{}
	//cache, err := NewThreadSafeCache(10000)
	//Connection statuses are kept for 60 minutes
	//we keep at most 10k urls
	cache, err := NewTTL(10000, time.Minute*60)
	if err != nil {
		return nil, err
	}
	c.cache = cache
	return &c, nil
}

func (c *ConnectivityCache) IsOk(url string) bool {
	ok := c.cache.Contains(url)
	return ok
}

func (c *ConnectivityCache) Test(u string) error {
	if c.browser == nil {
		return errors.New("connectivity cache has no browser. call SetBrowser first")
	}
	err := c.browser.Open(u)
	if err == nil {
		c.cache.Add(u, CacheInfo{
			added: time.Now(),
		})
	}
	return err
}

func (c *ConnectivityCache) SetBrowser(bow *browser.Browser) {
	c.browser = bow
}

func (c *ConnectivityCache) ClearBrowser() {
	c.browser = nil
}
