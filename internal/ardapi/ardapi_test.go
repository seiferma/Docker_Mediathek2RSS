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
	ardAPI := CreateArdAPI(maxEpisodes, fnGet)
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
	ardAPI := CreateArdAPI(maxEpisodes, fnGet)
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
	ardAPI := CreateArdAPI(maxEpisodes, fnGet)
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
