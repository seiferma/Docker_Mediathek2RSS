package zdffeed

import (
	"errors"
	"testing"

	"github.com/seiferma/docker_mediathek2rss/internal/zdfapi"
)

func TestFindBestMatchingImageURL(t *testing.T) {
	assertFindBestMatchingImageURL(t, "2", map[string]string{
		"10x10": "1",
		"2x51":  "2",
		"3x33":  "3",
	})
	assertFindBestMatchingImageURL(t, "1", map[string]string{
		"10x10": "1",
		"auto":  "2",
		"3xmax": "3",
		"5x5":   "4",
	})
}

func TestFindHighestResolutionStream(t *testing.T) {
	const prefix = "https://bla/sendung_zmr_"
	const testURL = prefix + "1628k_p13v15.mp4"
	const expectedURL = prefix + "3328k_p36v14.mp4"
	fnGet := func(api *zdfapi.ZDFApi, URL string, onlyPeek bool) (result []byte, err error) {
		switch URL {
		case prefix + "3328k_p36v13.mp4":
			return []byte{}, nil
		case prefix + "3328k_p36v14.mp4":
			return []byte{}, nil
		}
		return []byte{}, errors.New("test error")
	}
	api, _ := zdfapi.CreateZDFApiWithFnGet(1, fnGet)
	actual := findHighestResolutionStream(&api, testURL)
	assertEquals(t, expectedURL, actual)
}

func assertFindBestMatchingImageURL(t *testing.T, expected string, images map[string]string) {
	image := &zdfapi.ZDFTeaserImage{
		Images: images,
	}
	actual := findBestMatchingImageURL(image)
	assertEquals(t, expected, actual)
}

func assertEquals(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Fatalf("Expected %v but got %v.", expected, actual)
	}
}
