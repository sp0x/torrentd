package cache

import (
	"sync"
	"time"
)

type CacheWithTTL struct {
	*LRU
	lock sync.RWMutex
	TTL  time.Duration
}
