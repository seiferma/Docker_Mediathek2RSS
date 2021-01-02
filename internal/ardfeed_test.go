package internal

import (
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/seiferma/docker_mediathek2rss/internal/ardapi"
)

const defaultMediaWidth = 1920

func TestCreateRssFeedInvalidShowId(t *testing.T) {
	urlToFilename := map[string](string){}
	_, err := createRssFeedMocked("Y3JpZDovL2Z1bmsubmV0LzEwMzE", 2, defaultMediaWidth, urlToFilename, CreateArdRssFeed)
	if err == nil {
		t.Fatal("There should be an error.")
	}
}

func TestCreateRssFeedInvalidEpisodeURL(t *testing.T) {
	urlToFilename := map[string](string){}
	urlToFilename["https://api.ardmediathek.de/page-gateway/widgets/ard/asset/Y3JpZDovL2Z1bmsubmV0LzEwMzE?pageNumber=0&pageSize=2"] = "Y3JpZDovL2Z1bmsubmV0LzEwMzE.json"

	_, err := createRssFeedMocked("Y3JpZDovL2Z1bmsubmV0LzEwMzE", 2, defaultMediaWidth, urlToFilename, CreateArdRssFeed)
	if err == nil {
		t.Fatal("There should be an error.")
	}
}

func TestCreateRssFeedValid(t *testing.T) {
	urlToFilename := map[string](string){}
	urlToFilename["https://api.ardmediathek.de/page-gateway/widgets/ard/asset/Y3JpZDovL2Z1bmsubmV0LzEwMzE?pageNumber=0&pageSize=2"] = "Y3JpZDovL2Z1bmsubmV0LzEwMzE.json"
	urlToFilename["https://api.ardmediathek.de/page-gateway/pages/ard/item/Y3JpZDovL2Z1bmsubmV0LzEwMzEvdmlkZW8vMTcwNzQ0Mg?devicetype=pc&embedded=true"] = "Y3JpZDovL2Z1bmsubmV0LzEwMzEvdmlkZW8vMTcwNzQ0Mg.json"
	urlToFilename["https://api.ardmediathek.de/page-gateway/pages/ard/item/Y3JpZDovL2Z1bmsubmV0LzEwMzEvdmlkZW8vMTcwNjkzOA?devicetype=pc&embedded=true"] = "Y3JpZDovL2Z1bmsubmV0LzEwMzEvdmlkZW8vMTcwNjkzOA.json"

	result, err := createRssFeedMocked("Y3JpZDovL2Z1bmsubmV0LzEwMzE", 2, defaultMediaWidth, urlToFilename, CreateArdRssFeed)
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

func createRssFeedMocked(showID string, maxEpisodes int, requestedMediaWidth int, urlToFilename map[string](string), fnCreate func(showID string, requestedMediaWidth int, ardAPI *ardapi.ArdAPI) (result string, err error)) (result string, err error) {
	fnGetHTTP := func(URL string) (result []byte, err error) {
		filename, ok := urlToFilename[URL]
		if !ok {
			err = errors.New("unknown URL")
			return
		}
		result, err = ioutil.ReadFile("testdata/" + filename)
		return
	}
	ardAPI := ardapi.CreateArdAPI(maxEpisodes, fnGetHTTP)
	result, err = fnCreate(showID, requestedMediaWidth, &ardAPI)
	return
}

func TestFeedImageInformationExtraction(t *testing.T) {
	image := ardapi.ShowImage{
		Src:         "http://foo.bar/{width}.png",
		Title:       "Foo Title",
		Alt:         "Foo Alternative",
		AspectRatio: "16x9",
	}
	url, alt := getFeedImageURLAndAlt(image, 1280)
	assertStringEquals(t, "http://foo.bar/"+toString(1280)+".png", url)
	assertStringEquals(t, "Foo Alternative", alt)
}

func TestFeedImageFinderWith169(t *testing.T) {
	feedImageCandidates := map[string](ardapi.ShowImage){}
	feedImageCandidates["16x9"] = ardapi.ShowImage{
		Src:         "http://foo.bar/image.png",
		Title:       "Foo Title",
		Alt:         "Foo Alternative",
		AspectRatio: "16x9",
	}
	feedImageCandidates["4x3"] = ardapi.ShowImage{
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
	feedImageCandidates := map[string](ardapi.ShowImage){}
	feedImageCandidates["16x10"] = ardapi.ShowImage{
		Src:         "http://foo.bar/image.png",
		Title:       "Foo Title",
		Alt:         "Foo Alternative",
		AspectRatio: "16x10",
	}
	feedImageCandidates["4x3"] = ardapi.ShowImage{
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

func TestGetCacheKey(t *testing.T) {
	assertGetCacheKey(t, "123", 123, "123#123")
}

func assertGetCacheKey(t *testing.T, showID string, requestedWidth int, expectedKey string) {
	cacheKey := getCacheKey(showID, requestedWidth)
	if expectedKey != cacheKey {
		t.Fatalf("Expected cache key %v but got %v.", expectedKey, cacheKey)
	}
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
