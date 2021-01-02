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
	mediaWidth := 42
	counter := 0
	fnCreate := func(s string, i int) (string, error) {
		counter = counter + 1
		return fmt.Sprintf("%v", counter), nil
	}

	var result string
	result, _ = CreateRssFeedCached(showID, mediaWidth, &cache, fnCreate)
	assertEquals(t, "1", result)
	result, _ = CreateRssFeedCached(showID, mediaWidth, &cache, fnCreate)
	assertEquals(t, "1", result)
	currentTime = currentTime.Add(cacheDuration + 1)
	result, _ = CreateRssFeedCached(showID, mediaWidth, &cache, fnCreate)
	assertEquals(t, "2", result)
}

func assertEquals(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Fatalf("Expected %v but got %v.", expected, actual)
	}
}
