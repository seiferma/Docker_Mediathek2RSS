package internal

import (
	"fmt"
	"log"
)

// CreateRssFeedCached produces a RSS feed for a show.
// It takes the identifier of the show as requested by the JSON API, request parameters, a pointer to the
// cache and a function to dispatch the feed creation to. It yields the RSS feed as string and an error.
//
// The requested width might not be met perfectly depending on the available media. However, the logic tries to get to the requested
// width as close as possible.
func CreateRssFeedCached(showIdentifier string, parameters RequestParameters, cache *Cache, fnCreate func(string, RequestParameters) (string, error)) (result string, err error) {
	// directly return valid cache entry
	cacheKey := getCacheKey(showIdentifier, parameters)
	var foundCacheEntry bool
	result, foundCacheEntry = cache.GetContent(cacheKey)
	if foundCacheEntry {
		log.Printf("Answering request for %v / %v from cache.", showIdentifier, parameters)
		return
	}

	// calculate RSS feed
	result, err = fnCreate(showIdentifier, parameters)

	// cache result
	if err == nil {
		cache.StoreContent(cacheKey, result)
	}
	return
}

func getCacheKey(showID string, parameters RequestParameters) string {
	return fmt.Sprintf("%v#%v", showID, parameters)
}
