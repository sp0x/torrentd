package cache

import (
	"errors"
	"sync"
	"time"
)

type CacheWithTTL struct {
	*LRU
	lock sync.RWMutex
	TTL  time.Duration
}

type ttledCacheValue struct {
	value          interface{}
	lastAccessTime time.Time
}

func NewTTL(size int, ttl time.Duration) (LRUCache, error) {
	return NewTTLWithEvict(size, ttl, nil)
}

func NewTTLWithEvict(size int, ttl time.Duration, onEvict EvictionCallback) (LRUCache, error) {
	if size <= 0 {
		return nil, errors.New("size must be a positive int")
	}
	lru, err := NewLRU(size, func(k interface{}, v interface{}) {
		if onEvict != nil {
			onEvict(k, v.(ttledCacheValue).value)
		}
	})
	if err != nil {
		return nil, err
	}
	ttlLruCache := &CacheWithTTL{LRU: lru, TTL: ttl}
	go ttlLruCache.removeStales()
	return ttlLruCache, nil
}

// Add adds the item to the cache. It also includes the `lastAccessTime` to the value.
// Life of an item can be increased by calling `Add` multiple times on the same key.
func (c *CacheWithTTL) Add(key, value interface{}) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.LRU.Add(key,
		ttledCacheValue{
			value:          value,
			lastAccessTime: time.Now(),
		})
}

// Get looks up a key's value from the cache.
// Also, it unmarshals `lastAccessTime` from `Get` response
func (c *CacheWithTTL) Get(key interface{}) (value interface{}, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	val, ok := c.LRU.Get(key)
	// while the cleanup routine is catching will the other items, remove the item
	// if someone tries to access it through this GET call.
	if ok {
		if time.Now().After(val.(ttledCacheValue).lastAccessTime.Add(c.TTL)) {
			c.LRU.Remove(key)
		} else {
			return val.(ttledCacheValue).value, ok
		}
	}

	return nil, false
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
// Also, it unmarshals the `lastAccessTime` from the result
func (c *CacheWithTTL) Peek(key interface{}) (value interface{}, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	val, ok := c.LRU.Peek(key)
	if ok {
		return val.(ttledCacheValue).value, ok
	}
	return val, ok
}

// Contains checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (c *CacheWithTTL) Contains(key interface{}) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.LRU.Contains(key)
}

// Purge is used to completely clear the cache.
func (c *CacheWithTTL) Clear() {
	c.lock.Lock()
	c.LRU.Clear()
	c.lock.Unlock()
}

// Remove removes the provided key from the cache.
func (c *CacheWithTTL) Remove(key interface{}) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.LRU.Remove(key)
}

// RemoveOldest removes the oldest item from the cache.
func (c *CacheWithTTL) RemoveOldest() (key interface{}, value interface{}, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.LRU.RemoveOldest()
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *CacheWithTTL) Keys() []interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.LRU.Keys()
}

// Len returns the number of items in the cache.
func (c *CacheWithTTL) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.LRU.Len()
}

func (t *CacheWithTTL) removeStales() {
	//TODO: add termination way
	// - use heap instead of ticker
	ticker := time.NewTicker(200 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			for _, key := range t.Keys() {
				t.lock.Lock()
				val, ok := t.LRU.Get(key)
				t.lock.Unlock()
				//If there's a value behind the key and it's a stale one, exceeding the ttl for the cache
				if ok && time.Now().After(val.(ttledCacheValue).lastAccessTime.Add(t.TTL)) {
					t.Remove(key)
				}
			}
		}
	}
}
