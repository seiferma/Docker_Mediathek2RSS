package internal

import (
	"sync"
	"time"
)

// Cache is a data structure for holding content that expires after a given amount of time.
// Access to the contained data as well as creation of the cache shall be done via the
// functions provided in this package.
type Cache struct {
	entries       sync.Map
	entryDuration time.Duration
	fnNow         func() time.Time
}

// CreateCache creates a new cache instance with a cache entry expiration of the given
// cacheDuration parameter.
func CreateCache(cacheDuration time.Duration) Cache {
	return CreateCacheWithNowFunction(cacheDuration, time.Now)
}

// CreateCacheWithNowFunction creates a new cache instance with a cache entry expiration
// of the given cacheDuration parameter as well as a user defined now function.
func CreateCacheWithNowFunction(cacheDuration time.Duration, fnNow func() time.Time) Cache {
	return Cache{
		entryDuration: cacheDuration,
		fnNow:         fnNow,
		entries:       sync.Map{},
	}
}

type cacheValue struct {
	ValidTo time.Time
	Content string
}

// GetContent retrieves the content of the cache for a given key. The function yields two results.
// If a not expired entry has been found, the found result is set to true and the content contains the entry.
// If there is no entry or it is expired, found will be false.
func (cache *Cache) GetContent(key string) (content string, found bool) {
	var loadResult interface{}
	loadResult, found = cache.entries.Load(key)
	if !found {
		return
	}

	cacheEntry := loadResult.(cacheValue)
	content = cacheEntry.Content
	if cache.fnNow().After(cacheEntry.ValidTo) {
		cache.entries.Delete(key)
		found = false
	}
	return
}

// StoreContent adds a new entry to the cache, which overrides existing entries.
func (cache *Cache) StoreContent(key string, content string) {
	cache.entries.Store(key, cacheValue{
		ValidTo: cache.fnNow().Add(cache.entryDuration),
		Content: content,
	})
}
