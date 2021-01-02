package zdffeed

import (
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