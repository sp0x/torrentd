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
	cache   map[string]CacheInfo
}

func NewConnectivityCache() *ConnectivityCache {
	c := ConnectivityCache{}
	c.cache = make(map[string]CacheInfo)
	return &c
}

func (c *ConnectivityCache) IsOk(url string) bool {
	_, ok := c.cache[url]
	return ok
}

func (c *ConnectivityCache) Test(u string) error {
	if c.browser == nil {
		return errors.New("connectivity cache has no browser. call SetBrowser first")
	}
	err := c.browser.Open(u)
	if err == nil {
		c.cache[u] = CacheInfo{
			added: time.Now(),
		}
	}
	return err
}

func (c *ConnectivityCache) SetBrowser(bow *browser.Browser) {
	c.browser = bow
}

func (c *ConnectivityCache) ClearBrowser() {
	c.browser = nil
}
