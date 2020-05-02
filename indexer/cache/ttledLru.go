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
