package main

import (
	"errors"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/seiferma/docker-ard2rss/internal"
)

func TestCreateRssFeedCachedValid(t *testing.T) {
	// mock cache
	currentTime := time.Unix(0, 0)
	fnNow := func() time.Time {
		return currentTime
	}
	cache := internal.CreateCacheWithNowFunction(cacheDuration, fnNow)

	// inject cache mock into function to be tested
	fnCreate := func(showID string, ardAPI *internal.ArdAPI) (result string, err error) {
		return createRssFeedCached(showID, ardAPI, &cache)
	}

	// provide mock data to the HTTP GET function
	urlToFilename := map[string](string){}
	urlToFilename["https://api.ardmediathek.de/page-gateway/widgets/ard/asset/Y3JpZDovL2Z1bmsubmV0LzEwMzE?pageNumber=0&pageSize=2"] = "Y3JpZDovL2Z1bmsubmV0LzEwMzE.json"
	urlToFilename["https://api.ardmediathek.de/page-gateway/pages/ard/item/Y3JpZDovL2Z1bmsubmV0LzEwMzEvdmlkZW8vMTcwNzQ0Mg?devicetype=pc&embedded=true"] = "Y3JpZDovL2Z1bmsubmV0LzEwMzEvdmlkZW8vMTcwNzQ0Mg.json"
	urlToFilename["https://api.ardmediathek.de/page-gateway/pages/ard/item/Y3JpZDovL2Z1bmsubmV0LzEwMzEvdmlkZW8vMTcwNjkzOA?devicetype=pc&embedded=true"] = "Y3JpZDovL2Z1bmsubmV0LzEwMzEvdmlkZW8vMTcwNjkzOA.json"

	// test creating the feed
	result, err := createRssFeedMocked("Y3JpZDovL2Z1bmsubmV0LzEwMzE", 2, urlToFilename, fnCreate)
	if err != nil {
		t.Fatalf("There should not be an error.\n%v", err)
	}
	expectedBytes, err := ioutil.ReadFile("testdata/Y3JpZDovL2Z1bmsubmV0LzEwMzE.xml")
	expected := string(expectedBytes)
	if result != expected {
		// ioutil.WriteFile("testdata/actual.xml", []byte(result), 0644)
		// ioutil.WriteFile("testdata/expected.xml", []byte(expected), 0644)
		t.Fatalf("The created XML is not as expected. Created:\n%v\n\nExpected:\n%v", result, expected)
	}

	// test reading from cache
	urlToFilenameEmpty := map[string](string){}
	result2, err2 := createRssFeedMocked("Y3JpZDovL2Z1bmsubmV0LzEwMzE", 2, urlToFilenameEmpty, fnCreate)
	if err2 != nil {
		t.Fatalf("There should not be an error because the result is already cached.\n%v", err)
	}
	if result != result2 {
		t.Fatal("The first and the second result are different, which should not be the case.")
	}

	// test reading from expired cache
	currentTime = currentTime.Add(cacheDuration + 1)
	_, err3 := createRssFeedMocked("Y3JpZDovL2Z1bmsubmV0LzEwMzE", 2, urlToFilenameEmpty, fnCreate)
	if err3 == nil {
		t.Fatal("There should be an error.")
	}
}

func TestCreateRssFeedInvalidShowId(t *testing.T) {
	urlToFilename := map[string](string){}
	_, err := createRssFeedMocked("Y3JpZDovL2Z1bmsubmV0LzEwMzE", 2, urlToFilename, createRssFeed)
	if err == nil {
		t.Fatal("There should be an error.")
	}
}

func TestCreateRssFeedInvalidEpisodeURL(t *testing.T) {
	urlToFilename := map[string](string){}
	urlToFilename["https://api.ardmediathek.de/page-gateway/widgets/ard/asset/Y3JpZDovL2Z1bmsubmV0LzEwMzE?pageNumber=0&pageSize=2"] = "Y3JpZDovL2Z1bmsubmV0LzEwMzE.json"

	_, err := createRssFeedMocked("Y3JpZDovL2Z1bmsubmV0LzEwMzE", 2, urlToFilename, createRssFeed)
	if err == nil {
		t.Fatal("There should be an error.")
	}
}

func TestCreateRssFeedValid(t *testing.T) {
	urlToFilename := map[string](string){}
	urlToFilename["https://api.ardmediathek.de/page-gateway/widgets/ard/asset/Y3JpZDovL2Z1bmsubmV0LzEwMzE?pageNumber=0&pageSize=2"] = "Y3JpZDovL2Z1bmsubmV0LzEwMzE.json"
	urlToFilename["https://api.ardmediathek.de/page-gateway/pages/ard/item/Y3JpZDovL2Z1bmsubmV0LzEwMzEvdmlkZW8vMTcwNzQ0Mg?devicetype=pc&embedded=true"] = "Y3JpZDovL2Z1bmsubmV0LzEwMzEvdmlkZW8vMTcwNzQ0Mg.json"
	urlToFilename["https://api.ardmediathek.de/page-gateway/pages/ard/item/Y3JpZDovL2Z1bmsubmV0LzEwMzEvdmlkZW8vMTcwNjkzOA?devicetype=pc&embedded=true"] = "Y3JpZDovL2Z1bmsubmV0LzEwMzEvdmlkZW8vMTcwNjkzOA.json"

	result, err := createRssFeedMocked("Y3JpZDovL2Z1bmsubmV0LzEwMzE", 2, urlToFilename, createRssFeed)
	if err != nil {
		t.Fatalf("There should not be an error.\n%v", err)
	}
	expectedBytes, err := ioutil.ReadFile("testdata/Y3JpZDovL2Z1bmsubmV0LzEwMzE.xml")
	expected := string(expectedBytes)

	if strings.Compare(result, expected) != 0 {
		// ioutil.WriteFile("testdata/actual.xml", []byte(result), 0644)
		// ioutil.WriteFile("testdata/expected.xml", []byte(expected), 0644)
		t.Fatalf("The created XML is not as expected. Created:\n%v\n\nExpected:\n%v", result, expected)
	}
}

func createRssFeedMocked(showID string, maxEpisodes int, urlToFilename map[string](string), fnCreate func(showID string, ardAPI *internal.ArdAPI) (result string, err error)) (result string, err error) {
	fnGetHTTP := func(URL string) (result []byte, err error) {
		filename, ok := urlToFilename[URL]
		if !ok {
			err = errors.New("unknown URL")
			return
		}
		result, err = ioutil.ReadFile("testdata/" + filename)
		return
	}
	ardAPI := internal.CreateArdAPI(maxEpisodes, fnGetHTTP)
	result, err = fnCreate(showID, &ardAPI)
	return
}

func TestFeedImageInformationExtraction(t *testing.T) {
	image := internal.ShowImage{
		Src:         "http://foo.bar/{width}.png",
		Title:       "Foo Title",
		Alt:         "Foo Alternative",
		AspectRatio: "16x9",
	}
	url, alt := getFeedImageURLAndAlt(image)
	assertStringEquals(t, "http://foo.bar/"+toString(defaultMediaWidth)+".png", url)
	assertStringEquals(t, "Foo Alternative", alt)
}

func TestFeedImageFinderWith169(t *testing.T) {
	feedImageCandidates := map[string](internal.ShowImage){}
	feedImageCandidates["16x9"] = internal.ShowImage{
		Src:         "http://foo.bar/image.png",
		Title:       "Foo Title",
		Alt:         "Foo Alternative",
		AspectRatio: "16x9",
	}
	feedImageCandidates["4x3"] = internal.ShowImage{
		Src:         "http://bar.foo/image.png",
		Title:       "Bar Title",
		Alt:         "Bar Alternative",
		AspectRatio: "4x3",
	}
	actualImage := getFeedImage(feedImageCandidates)
	expectedImage := feedImageCandidates["16x9"]
	if actualImage != expectedImage {
		t.Fatalf("Selected image is wrong.")
	}
}

func TestFeedImageFinderWithout169(t *testing.T) {
	feedImageCandidates := map[string](internal.ShowImage){}
	feedImageCandidates["16x10"] = internal.ShowImage{
		Src:         "http://foo.bar/image.png",
		Title:       "Foo Title",
		Alt:         "Foo Alternative",
		AspectRatio: "16x10",
	}
	feedImageCandidates["4x3"] = internal.ShowImage{
		Src:         "http://bar.foo/image.png",
		Title:       "Bar Title",
		Alt:         "Bar Alternative",
		AspectRatio: "4x3",
	}
	actualImage := getFeedImage(feedImageCandidates)
	if actualImage != feedImageCandidates["16x10"] && actualImage != feedImageCandidates["4x3"] {
		t.Fatal("The returned image is not in the initial list.")
	}
}

func TestConvertToString(t *testing.T) {
	assertConvertEquals(t, 0, "0")
	assertConvertEquals(t, 1, "1")
	assertConvertEquals(t, 5.9, "5.9")
	assertConvertEquals(t, "a", "a")
}

func assertConvertEquals(t *testing.T, v interface{}, expected string) {
	actual := toString(v)
	assertStringEquals(t, expected, actual)
}

func assertStringEquals(t *testing.T, expected, actual string) {
	if strings.Compare(actual, expected) != 0 {
		t.Fatalf("Expected %v but got %v.", expected, actual)
	}
}
