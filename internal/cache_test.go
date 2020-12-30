package internal

import (
	"strings"
	"testing"
	"time"
)

func TestCreateCache(t *testing.T) {
	const expectedDuration = 5
	cache := CreateCache(expectedDuration)

	if cache.entryDuration != expectedDuration {
		t.Errorf("The actual duration of %v is not the same as the expected duration of %v.", cache.entryDuration, expectedDuration)
	}

	cache.entries.Range(func(k, v interface{}) bool {
		t.Errorf("Expected an empty cache but found a key %v and %v value.", k, v)
		return true
	})
}

func TestAddEntry(t *testing.T) {
	const expectedDuration = 10
	const key = "test"
	const value = "123"
	now := time.Unix(0, 0)
	fnNow := func() time.Time {
		return now
	}
	cache := CreateCacheWithNowFunction(expectedDuration, fnNow)
	cache.StoreContent(key, value)

	hasEntry := false
	cache.entries.Range(func(k, v interface{}) bool {
		hasEntry = true
		if strings.Compare(key, k.(string)) != 0 {
			t.Errorf("Expected the key to be %v but was %v.", key, k)
		}
		cacheValue := v.(cacheValue)
		if strings.Compare(value, cacheValue.Content) != 0 {
			t.Errorf("Expected the value to be %v but was %v.", value, v)
		}
		expirationTime := now.Add(expectedDuration)
		if !cacheValue.ValidTo.Equal(expirationTime) {
			t.Errorf("Expected the expiration time to be %v but it was %v.", expirationTime, cacheValue.ValidTo)
		}
		return true
	})
	if !hasEntry {
		t.Errorf("The cache has no entry but it should have one.")
	}
}

func TestGetEntry(t *testing.T) {
	const expectedDuration = 10
	const key = "test"
	const value = "123"
	now := time.Unix(0, 0)
	fnNow := func() time.Time {
		return now
	}
	cache := CreateCacheWithNowFunction(expectedDuration, fnNow)
	cache.StoreContent(key, value)

	_, foundUnexpired := cache.GetContent(key)
	if !foundUnexpired {
		t.Error("There should have been an entry returned.")
	}

	now = now.Add(expectedDuration + 1)

	_, foundExpired := cache.GetContent(key)
	if foundExpired {
		t.Error("There should not have been an entry returned.")
	}
}
