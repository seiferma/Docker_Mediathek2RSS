package ardapi

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestGetShow(t *testing.T) {
	const maxEpisodes = 2
	const showID = "test"
	expctedURL := fmt.Sprintf("https://api.ardmediathek.de/page-gateway/widgets/ard/asset/%v?pageNumber=0&pageSize=%v", showID, maxEpisodes)
	fnGet := func(url string) (result []byte, err error) {
		if strings.Compare(expctedURL, url) != 0 {
			t.Fatalf("We expected the URL %v but received %v.", expctedURL, url)
		}
		return ioutil.ReadFile("../testdata/Y3JpZDovL2Z1bmsubmV0LzEwMzE.json")
	}

	// get a show from the API
	ardAPI := CreateArdAPIWithGetFunc(maxEpisodes, fnGet)
	result, err := ardAPI.GetShow(showID)

	// assert results
	if err != nil {
		t.Fatalf("There should be no error reported.")
		return
	}

	actualEpisodes := len(result.Teasers)
	if len(result.Teasers) != maxEpisodes {
		t.Fatalf("Expected %v episodes but got %v.", maxEpisodes, actualEpisodes)
	}
}

func TestGetShowWithoutTeasers(t *testing.T) {
	const maxEpisodes = 2
	const showID = "test"
	fnGet := func(url string) (result []byte, err error) {
		return ioutil.ReadFile("../testdata/Y3JpZDovL2Z1bmsubmV0LzEwMzE_noTeasers.json")
	}

	// get a show from the API
	ardAPI := CreateArdAPIWithGetFunc(maxEpisodes, fnGet)
	_, err := ardAPI.GetShow(showID)

	// assert results
	if err == nil {
		t.Fatalf("There should be an error reported.")
		return
	}
}

func TestGetShowWithLessThanMaxTeasers(t *testing.T) {
	const maxEpisodes = 2
	const showID = "test"
	fnGet := func(url string) (result []byte, err error) {
		result, err = ioutil.ReadFile("../testdata/Y3JpZDovL2Z1bmsubmV0LzEwMzE_oneTeaser.json")
		return
	}

	// get a show from the API
	ardAPI := CreateArdAPIWithGetFunc(maxEpisodes, fnGet)
	result, err := ardAPI.GetShow(showID)

	// assert results
	if err != nil {
		t.Fatalf("There should be no error reported.")
		return
	}

	actualEpisodes := len(result.Teasers)
	if len(result.Teasers) != 1 {
		t.Fatalf("Expected %v episodes but got %v.", 1, actualEpisodes)
	}
}

func TestGetVideoByURLWithMultipleStreamURLs(t *testing.T) {
	fnGet := func(url string) (result []byte, err error) {
		result, err = ioutil.ReadFile("../testdata/Y3JpZDovL2Rhc2Vyc3RlLmRlL3RhZ2VzdGhlbWVuL2Q1N2VjY2VmLWY2ZTQtNDVhZS1iNGNlLTcyMThiZjBhMzMxZg.json")
		return
	}

	// get a show from the API
	const maxEpisodes = 2
	ardAPI := CreateArdAPIWithGetFunc(maxEpisodes, fnGet)
	result, err := ardAPI.GetVideoByURL("https://api.ardmediathek.de/page-gateway/pages/ard/item/Y3JpZDovL2Rhc2Vyc3RlLmRlL3RhZ2VzdGhlbWVuL2Q1N2VjY2VmLWY2ZTQtNDVhZS1iNGNlLTcyMThiZjBhMzMxZg?devicetype=pc&embedded=true")

	// assert that result exists
	if err != nil {
		t.Fatalf("There should be no error reported.")
		return
	}

	// assert stream URLs of show
	mediaStreams := result.Widgets[0].MediaCollection.Embedded.MediaArray[0].MediaStreamArray
	if !assertEquals(t, 5, len(mediaStreams)) ||
		!assertEquals(t, 1, len(mediaStreams[0].Stream.StreamUrls)) ||
		!assertEquals(t, "https://adaptive.tagesschau.de/i/video/2021/0925/TV-20210925-2356-5100.,webs,websm,webm,webml,webl,webxl,.h264.mp4.csmil/master.m3u8", mediaStreams[0].Stream.StreamUrls[0]) ||
		!assertEquals(t, 3, len(mediaStreams[1].Stream.StreamUrls)) ||
		!assertContains(t, mediaStreams[1].Stream.StreamUrls, "https://media.tagesschau.de/video/2021/0925/TV-20210925-2356-5100.websm.h264.mp4") ||
		!assertContains(t, mediaStreams[1].Stream.StreamUrls, "https://media.tagesschau.de/video/2021/0925/TV-20210925-2356-5100.webs.h264.mp4") ||
		!assertContains(t, mediaStreams[1].Stream.StreamUrls, "https://download.media.tagesschau.de/video/2021/0925/TV-20210925-2356-5100.websm.h264.mp4") {
		return
	}
}

func assertEquals(t *testing.T, expected, actual interface{}) bool {
	if expected != actual {
		t.Fatalf("Expected \"%v\" but got \"%v\".", expected, actual)
		return false
	}
	return true
}

func assertContains(t *testing.T, collection []string, element string) bool {
	for _, collectionElement := range collection {
		if element == collectionElement {
			return true
		}
	}
	t.Fatalf("Element \"%v\" not contained in collection.", element)
	return false
}
