package main

import (
	"testing"
	"time"

	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
)

func makeEntry(key string, sizeBytes int64) *CacheEntry {
	return &CacheEntry{
		CollectionKey: key,
		Products: []*collectorpb.Product{
			{ItemId: 1, Name: "test-product"},
		},
		FetchedAt: time.Now(),
		ExpiresAt: time.Now().Add(30 * time.Minute),
		Hash:      "abc123",
		SizeBytes: sizeBytes,
	}
}

func TestLRU_PutGet(t *testing.T) {
	cache := NewLRUCache(1024*1024, 30*time.Minute)

	entry := makeEntry("user1:serum", 100)
	cache.Put("user1:serum", entry)

	got := cache.Get("user1:serum")
	if got == nil {
		t.Fatal("expected entry, got nil")
	}
	if got.CollectionKey != "user1:serum" {
		t.Errorf("expected key user1:serum, got %s", got.CollectionKey)
	}
	if len(got.Products) != 1 {
		t.Errorf("expected 1 product, got %d", len(got.Products))
	}
}

func TestLRU_GetMiss(t *testing.T) {
	cache := NewLRUCache(1024*1024, 30*time.Minute)

	got := cache.Get("nonexistent")
	if got != nil {
		t.Errorf("expected nil for miss, got %v", got)
	}
}

func TestLRU_Eviction(t *testing.T) {
	// Max 250 bytes → only 2 entries of 100 bytes fit
	cache := NewLRUCache(250, 30*time.Minute)

	cache.Put("key1", makeEntry("key1", 100))
	cache.Put("key2", makeEntry("key2", 100))

	// Access key1 to make it recently used
	cache.Get("key1")

	// Adding key3 should evict key2 (LRU)
	cache.Put("key3", makeEntry("key3", 100))

	if cache.Get("key2") != nil {
		t.Error("expected key2 to be evicted")
	}
	if cache.Get("key1") == nil {
		t.Error("expected key1 to still exist")
	}
	if cache.Get("key3") == nil {
		t.Error("expected key3 to exist")
	}
}

func TestLRU_TTLExpiry(t *testing.T) {
	cache := NewLRUCache(1024*1024, 50*time.Millisecond)

	entry := &CacheEntry{
		CollectionKey: "key1",
		Products:      []*collectorpb.Product{{ItemId: 1}},
		FetchedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(50 * time.Millisecond),
		Hash:          "hash",
		SizeBytes:     100,
	}
	cache.Put("key1", entry)

	// Should be there immediately
	if cache.Get("key1") == nil {
		t.Fatal("expected entry to exist immediately after put")
	}

	// Wait for TTL expiry
	time.Sleep(60 * time.Millisecond)

	if cache.Get("key1") != nil {
		t.Error("expected entry to be expired")
	}
}

func TestLRU_SizeLimit(t *testing.T) {
	// 500 bytes max
	cache := NewLRUCache(500, 30*time.Minute)

	// Insert 5 entries of 100 bytes each → exactly fills
	for i := 0; i < 5; i++ {
		key := "key" + string(rune('a'+i))
		cache.Put(key, makeEntry(key, 100))
	}

	if cache.Len() != 5 {
		t.Errorf("expected 5 entries, got %d", cache.Len())
	}

	// Insert one more — should evict the oldest
	cache.Put("key_new", makeEntry("key_new", 100))

	if cache.Len() != 5 {
		t.Errorf("expected 5 entries after eviction, got %d", cache.Len())
	}
	if cache.SizeBytes() > 500 {
		t.Errorf("cache size %d exceeds max 500", cache.SizeBytes())
	}
}

func TestLRU_Delete(t *testing.T) {
	cache := NewLRUCache(1024*1024, 30*time.Minute)

	cache.Put("key1", makeEntry("key1", 100))

	if !cache.Delete("key1") {
		t.Error("expected delete to return true for existing key")
	}
	if cache.Delete("key1") {
		t.Error("expected delete to return false for already-deleted key")
	}
	if cache.Get("key1") != nil {
		t.Error("expected key1 to be gone after delete")
	}
}

func TestLRU_Update(t *testing.T) {
	cache := NewLRUCache(1024*1024, 30*time.Minute)

	cache.Put("key1", makeEntry("key1", 100))

	newEntry := &CacheEntry{
		CollectionKey: "key1",
		Products:      []*collectorpb.Product{{ItemId: 99, Name: "updated"}},
		FetchedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(30 * time.Minute),
		Hash:          "newhash",
		SizeBytes:     150,
	}
	cache.Put("key1", newEntry)

	got := cache.Get("key1")
	if got == nil {
		t.Fatal("expected entry after update")
	}
	if got.Hash != "newhash" {
		t.Errorf("expected hash newhash, got %s", got.Hash)
	}
	if got.Products[0].GetItemId() != 99 {
		t.Errorf("expected item_id 99, got %d", got.Products[0].GetItemId())
	}
}
