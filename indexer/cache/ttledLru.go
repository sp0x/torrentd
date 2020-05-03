package cache

import (
	"errors"
	"sync"
	"time"
)

type CacheWithTTL struct {
	*LRU
	lock        sync.RWMutex
	TTL         time.Duration
	stopChannel chan struct{}
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
	ttlLruCache := &CacheWithTTL{LRU: lru, TTL: ttl, stopChannel: make(chan struct{})}

	go ttlLruCache.removeStales()
	return ttlLruCache, nil
}

// Add adds the item to the cache. It also includes the `lastAccessTime` to the value.
// Life of an item can be increased by calling `Add` multiple times on the same key.
func (t *CacheWithTTL) Add(key, value interface{}) bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.LRU.Add(key,
		ttledCacheValue{
			value:          value,
			lastAccessTime: time.Now(),
		})
}

// Get looks up a key's value from the cache.
// Also, it unmarshals `lastAccessTime` from `Get` response
func (t *CacheWithTTL) Get(key interface{}) (value interface{}, ok bool) {
	t.lock.Lock()
	defer t.lock.Unlock()

	val, ok := t.LRU.Get(key)
	// while the cleanup routine is catching will the other items, remove the item
	// if someone tries to access it through this GET call.
	if ok {
		if time.Now().After(val.(ttledCacheValue).lastAccessTime.Add(t.TTL)) {
			t.LRU.Remove(key)
		} else {
			return val.(ttledCacheValue).value, ok
		}
	}

	return nil, false
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
// Also, it unmarshals the `lastAccessTime` from the result
func (t *CacheWithTTL) Peek(key interface{}) (value interface{}, ok bool) {
	t.lock.Lock()
	defer t.lock.Unlock()

	val, ok := t.LRU.Peek(key)
	if ok {
		return val.(ttledCacheValue).value, ok
	}
	return val, ok
}

// Contains checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (t *CacheWithTTL) Contains(key interface{}) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.LRU.Contains(key)
}

// Purge is used to completely clear the cache.
func (t *CacheWithTTL) Clear() {
	t.lock.Lock()
	t.LRU.Clear()
	t.lock.Unlock()
}

// Remove removes the provided key from the cache.
func (t *CacheWithTTL) Remove(key interface{}) bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.LRU.Remove(key)
}

// RemoveOldest removes the oldest item from the cache.
func (t *CacheWithTTL) RemoveOldest() (key interface{}, value interface{}, ok bool) {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.LRU.RemoveOldest()
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (t *CacheWithTTL) Keys() []interface{} {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.LRU.Keys()
}

// Len returns the number of items in the cache.
func (t *CacheWithTTL) Len() int {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.LRU.Len()
}

func (t *CacheWithTTL) Dispose() {
	t.stopChannel <- struct{}{}
}

func (t *CacheWithTTL) removeStales() {
	//TODO: add termination way
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
		case <-t.stopChannel:
			return
		}
	}
}
