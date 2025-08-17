package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	CacheEntries map[string]cacheEntry
	interval     time.Duration
	mu           sync.Mutex
}

func NewCache(interval time.Duration) *Cache {
	cache := Cache{
		CacheEntries: make(map[string]cacheEntry),
		interval:     interval,
	}
	go cache.reapLoop()
	return &cache
}

func (c *Cache) Add(key string, val []byte) {
	cacheEntry := cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
	c.mu.Lock()
	c.CacheEntries[key] = cacheEntry
	c.mu.Unlock()
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	elem, exists := c.CacheEntries[key]
	c.mu.Unlock()
	if exists {
		return elem.val, true
	}
	return nil, false
}

func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		for key, elem := range c.CacheEntries {
			if time.Now().Sub(elem.createdAt) > c.interval {

				delete(c.CacheEntries, key)
			}
		}
		c.mu.Unlock()
	}
}
