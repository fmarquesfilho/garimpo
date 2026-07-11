package main

import (
	"container/list"
	"sync"
	"time"

	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
)

// CacheEntry represents a single cached item.
type CacheEntry struct {
	CollectionKey string
	Products      []*collectorpb.Product
	FetchedAt     time.Time
	ExpiresAt     time.Time
	Hash          string
	SizeBytes     int64
	AccessCount   int64
	LastAccessed  time.Time

	// element is the pointer to the LRU list element for O(1) removal.
	element *list.Element
}

// LRUCache is a thread-safe LRU cache with size-based eviction and TTL.
type LRUCache struct {
	mu           sync.RWMutex
	store        map[string]*CacheEntry
	evictList    *list.List
	maxBytes     int64
	currentBytes int64
	ttl          time.Duration

	// Metrics
	hitsTotal      int64
	missesTotal    int64
	evictionsTotal int64
}

// NewLRUCache creates a new LRU cache with the given max size and TTL.
func NewLRUCache(maxBytes int64, ttl time.Duration) *LRUCache {
	return &LRUCache{
		store:     make(map[string]*CacheEntry),
		evictList: list.New(),
		maxBytes:  maxBytes,
		ttl:       ttl,
	}
}

// Get retrieves an entry from the cache. Returns nil if not found or expired.
func (c *LRUCache) Get(key string) *CacheEntry {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.store[key]
	if !ok {
		c.missesTotal++
		return nil
	}

	// Check TTL expiry
	if time.Now().After(entry.ExpiresAt) {
		c.removeEntry(entry)
		c.missesTotal++
		return nil
	}

	// Move to front (most recently used)
	c.evictList.MoveToFront(entry.element)
	entry.LastAccessed = time.Now()
	entry.AccessCount++
	c.hitsTotal++

	return entry
}

// Put inserts or updates an entry in the cache, evicting LRU entries if needed.
func (c *LRUCache) Put(key string, entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key already exists, remove old entry first
	if old, ok := c.store[key]; ok {
		c.removeEntry(old)
	}

	// Evict until we have space
	for c.currentBytes+entry.SizeBytes > c.maxBytes && c.evictList.Len() > 0 {
		c.evictOldest()
	}

	// Insert new entry
	elem := c.evictList.PushFront(entry)
	entry.element = elem
	entry.LastAccessed = time.Now()
	c.store[key] = entry
	c.currentBytes += entry.SizeBytes
}

// Delete removes a key from the cache. Returns true if it existed.
func (c *LRUCache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.store[key]
	if !ok {
		return false
	}

	c.removeEntry(entry)
	return true
}

// removeEntry removes an entry from both the map and the list.
// Caller must hold c.mu.
func (c *LRUCache) removeEntry(entry *CacheEntry) {
	c.evictList.Remove(entry.element)
	delete(c.store, entry.CollectionKey)
	c.currentBytes -= entry.SizeBytes
}

// evictOldest removes the least recently used entry.
// Caller must hold c.mu.
func (c *LRUCache) evictOldest() {
	elem := c.evictList.Back()
	if elem == nil {
		return
	}
	entry := elem.Value.(*CacheEntry)
	c.removeEntry(entry)
	c.evictionsTotal++
}

// Stats returns current cache statistics.
func (c *LRUCache) Stats() (sizeBytes int64, hits int64, misses int64, entries int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentBytes, c.hitsTotal, c.missesTotal, int64(len(c.store))
}

// Len returns the number of entries in the cache.
func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.store)
}

// SizeBytes returns the current total size in bytes.
func (c *LRUCache) SizeBytes() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentBytes
}
