package internal

import (
	"fmt"
	"testing"
	"time"
)

const cacheDuration = 5 * time.Minute

func TestCreateRssFeedCachedValid(t *testing.T) {
	// mock cache
	currentTime := time.Unix(0, 0)
	fnNow := func() time.Time {
		return currentTime
	}
	cache := CreateCacheWithNowFunction(cacheDuration, fnNow)

	showID := "test"
	parameters := RequestParameters{
		Width: 42,
	}
	counter := 0
	fnCreate := func(s string, parameters RequestParameters) (string, error) {
		counter = counter + 1
		return fmt.Sprintf("%v", counter), nil
	}

	var result string
	result, _ = CreateRssFeedCached(showID, parameters, &cache, fnCreate)
	assertEquals(t, "1", result)
	result, _ = CreateRssFeedCached(showID, parameters, &cache, fnCreate)
	assertEquals(t, "1", result)
	currentTime = currentTime.Add(cacheDuration + 1)
	result, _ = CreateRssFeedCached(showID, parameters, &cache, fnCreate)
	assertEquals(t, "2", result)
}

func TestGetCacheKey(t *testing.T) {
	parameters := RequestParameters{
		Width:                  42,
		MinimumLengthInSeconds: 3,
	}
	assertGetCacheKey(t, "123", parameters, "123#{42 3}")
}

func TestGetCacheKeyWithMissingParameters(t *testing.T) {
	parameters := RequestParameters{
		Width: 42,
	}
	assertGetCacheKey(t, "123", parameters, "123#{42 0}")
}

func assertGetCacheKey(t *testing.T, showID string, parameters RequestParameters, expectedKey string) {
	cacheKey := getCacheKey(showID, parameters)
	assertEquals(t, expectedKey, cacheKey)
}

func assertEquals(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Fatalf("Expected %v but got %v.", expected, actual)
	}
}
