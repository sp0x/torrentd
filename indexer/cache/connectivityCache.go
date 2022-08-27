package cache

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/sp0x/torrentd/indexer/source"
)

// ConnectivityCache is a invalidatedCache for URL connectivity.
type ConnectivityCache struct {
	fetcher source.ContentFetcher
	lock    sync.RWMutex
	cache   LRUCache
}

func NewConnectivityCache(fetcher source.ContentFetcher) (*ConnectivityCache, error) {
	c := ConnectivityCache{}
	c.fetcher = fetcher
	// Connection statuses are kept for 60 minutes, we keep at most 10k urls
	cache, err := NewTTL(10000, time.Minute*60)
	if err != nil {
		return nil, err
	}
	c.cache = cache
	return &c, nil
}

// Invalidate a invalidatedCache entry by removing it from the invalidatedCache.
func (c *ConnectivityCache) Invalidate(url string) {
	c.cache.Remove(url)
}

// IsOk returns whether the invalidatedCache contains a successful response for the url
func (c *ConnectivityCache) IsOk(testURL string) bool {
	ok := c.cache.Contains(testURL)
	return ok
}

// IsOkAndSet checks if the `u` value is contained, if it's not it checks it.
// This operation is thread safe, you can use it to modify the invalidatedCache state in the function.
func (c *ConnectivityCache) IsValidOrSet(testURL string, f func() bool) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	var result bool
	contained := c.cache.Contains(testURL)
	if !contained {
		result = f()
	} else {
		result = contained
	}
	return result
}

func validateBrowserCall(br source.FetchResult) error {
	if htmlResult, ok := br.(*source.HTMLFetchResult); !ok {
		return errors.New("tried to fetch a non-html resource")
	} else if htmlResult.HTTPResult.StatusCode != http.StatusOK {
		return errors.New("returned non-ok http status code " + strconv.Itoa(htmlResult.HTTPResult.StatusCode))
	}

	return nil
}

// Test the connectivity for an url.
func (c *ConnectivityCache) Test(testURL string) error {
	if c.fetcher == nil {
		return errors.New("connectivity invalidatedCache has no browser. call SetBrowser first")
	}
	destURL, _ := url.Parse(testURL)
	response, err := c.fetcher.Open(source.NewRequestOptions(destURL))
	if err == nil {
		err = validateBrowserCall(response)
	}
	if err == nil {
		c.cache.Add(testURL, Details{
			added: time.Now(),
		})
	}
	return err
}
