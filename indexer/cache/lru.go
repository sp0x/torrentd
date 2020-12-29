package cache

import (
	"container/list"
	"errors"
)

type EvictionCallback func(key interface{}, value interface{})

// LRU is a non-thread safe fixed size LRU cache
type LRU struct {
	size         int
	evictionList *list.List
	items        map[interface{}]*list.Element
	onEviction   EvictionCallback
}

// entry is used to hold value in an evictionList
type entry struct {
	key   interface{}
	value interface{}
}

// NewLRU creates a new LRU of the given size
func NewLRU(size int, onEviction EvictionCallback) (*LRU, error) {
	if size < 0 {
		return nil, errors.New("size must be > 0")
	}
	return &LRU{
		size:         size,
		evictionList: list.New(),
		items:        make(map[interface{}]*list.Element),
		onEviction:   onEviction,
	}, nil
}

// Clear the cahce
func (c *LRU) Clear() {
	for k, v := range c.items {
		if c.onEviction != nil {
			c.onEviction(k, v.Value.(*entry).value)
		}
		delete(c.items, k)
	}
	// Initialize the list again.
	c.evictionList.Init()
}

// Add a key with a given value
// returns true if an element was evicted so this one could be added
func (c *LRU) Add(key, value interface{}) bool {
	// HealthCheck if the key exists already, if so we'll move it to the front since it's modified
	if ent, ok := c.items[key]; ok {
		c.evictionList.MoveToFront(ent)
		ent.Value.(*entry).value = value
		return false
	}
	ent := &entry{key, value}
	// Add the item to the eviction list
	entry := c.evictionList.PushFront(ent)
	// Store the actual entry
	c.items[key] = entry
	// If we have more items than we can store
	shouldEvict := c.evictionList.Len() > c.size
	if shouldEvict {
		c.evictOldest()
	}
	return shouldEvict
}

// Get looks up a key's value from the cache.
// this updates the recent-ness of the cache
func (c *LRU) Get(key interface{}) (value interface{}, ok bool) {
	if ent, ok := c.items[key]; ok {
		c.evictionList.MoveToFront(ent)
		if ent.Value.(*entry) == nil {
			return nil, false
		}
		return ent.Value.(*entry).value, true
	}
	return
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *LRU) Contains(key interface{}) (ok bool) {
	_, ok = c.items[key]
	return ok
}

// Peek returns the key value (or nil if not found) without updating
// the "recently used"-ness of the key.
func (c *LRU) Peek(key interface{}) (value interface{}, ok bool) {
	var ent *list.Element
	if ent, ok = c.items[key]; ok {
		return ent.Value.(*entry).value, true
	}
	return nil, ok
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *LRU) Remove(key interface{}) (present bool) {
	if ent, ok := c.items[key]; ok {
		c.remove(ent)
		return true
	}
	return false
}

// RemoveOldest removes the oldest item in the cache
func (c *LRU) RemoveOldest() (key, value interface{}, ok bool) {
	ent := c.evictionList.Back()
	if ent != nil {
		c.remove(ent)
		kv := ent.Value.(*entry)
		return kv.key, kv.value, true
	}
	return nil, nil, false
}

// GetOldest returns the oldest item in the cache.
func (c *LRU) GetOldest() (key interface{}, value interface{}, ok bool) {
	ent := c.evictionList.Back()
	if ent != nil {
		entry := ent.Value.(*entry)
		return entry.key, entry.value, true
	}
	return nil, nil, false
}

// Keys returns the cache keys.
func (c *LRU) Keys() []interface{} {
	i := 0
	keys := make([]interface{}, len(c.items))
	for ent := c.evictionList.Back(); ent != nil; ent = ent.Prev() {
		keys[i] = ent.Value.(*entry).key
		i++
	}
	return keys
}

// Len returns the number of items in the cache
func (c *LRU) Len() int {
	return c.evictionList.Len()
}

// Resize changes the cache size
// returns the number of items removed when the cache shrinks
func (c *LRU) Resize(newSize int) int {
	diff := c.Len() - newSize
	if diff < 0 {
		diff = 0
	}
	// Remove items that aren't needed, if the cache is shrunk
	for i := 0; i < diff; i++ {
		c.evictOldest()
	}
	c.size = newSize
	return diff
}

func (c *LRU) evictOldest() {
	// Get the oldest item
	entry := c.evictionList.Back()
	if entry != nil {
		c.remove(entry)
	}
}

// remove an element from the cache and from our items
func (c *LRU) remove(e *list.Element) {
	// Uncache it
	c.evictionList.Remove(e)
	ent := e.Value.(*entry)
	// Remove actual item
	delete(c.items, ent.key)
	if c.onEviction != nil {
		c.onEviction(ent.key, ent.value)
	}
}
