package zdffeed

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"testing"

	"github.com/seiferma/docker_mediathek2rss/internal"
	"github.com/seiferma/docker_mediathek2rss/internal/zdfapi"
)

var defaultParameters internal.RequestParameters = internal.RequestParameters{
	Width: 1080,
}

func TestCreateRssFeedValid(t *testing.T) {
	urlToFilename := map[string](string){}
	urlToFilename["https://www.zdf.de/nachrichten/heute-journal"] = "zdf-heute-journal.html"
	urlToFilename["https://api.zdf.de/content/documents/zdf/comedy/zdf-magazin-royale"] = "zdf-magazin-royale.json"
	urlToFilename["https://api.zdf.de/search/documents/zdf/comedy/zdf-magazin-royale?q=*&limit=2&types=page-video&hasVideo=true"] = "zdf-magazin-royale-search.json"
	urlToFilename["https://api.zdf.de/tmd/2/ngplayer_2_4/vod/ptmd/mediathek/201218_2330_sendung_zmr"] = "zdf-magazin-royale-stream.json"
	urlToFilename["https://api.zdf.de/tmd/2/ngplayer_2_4/vod/ptmd/mediathek/201211_2300_sendung_zmr"] = "zdf-magazin-royale-stream2.json"

	result, err := createRssFeedMocked("comedy/zdf-magazin-royale", 2, defaultParameters, urlToFilename, CreateZdfRssFeed)
	if err != nil {
		t.Fatalf("There should not be an error.\n%v", err)
	}
	expectedBytes, _ := ioutil.ReadFile("../testdata/zdf-magazin-royale.xml")
	expected := string(expectedBytes)

	buildDateReplacement := regexp.MustCompile(`<lastBuildDate>[^<]+</lastBuildDate>`)
	result = buildDateReplacement.ReplaceAllString(result, "<lastBuildDate>NOW</lastBuildDate>")

	if strings.Compare(result, expected) != 0 {
		// ioutil.WriteFile("/tmp/actual.xml", []byte(result), 0644)
		// ioutil.WriteFile("/tmp/expected.xml", []byte(expected), 0644)
		t.Fatalf("The created XML is not as expected. Created:\n%v\n\nExpected:\n%v", result, expected)
	}
}

func TestCreateRssFeedValidWithLengthFilter(t *testing.T) {
	urlToFilename := map[string](string){}
	urlToFilename["https://www.zdf.de/nachrichten/heute-journal"] = "zdf-heute-journal.html"
	urlToFilename["https://api.zdf.de/content/documents/zdf/comedy/zdf-magazin-royale"] = "zdf-magazin-royale.json"
	urlToFilename["https://api.zdf.de/search/documents/zdf/comedy/zdf-magazin-royale?q=*&limit=2&types=page-video&hasVideo=true"] = "zdf-magazin-royale-search.json"
	urlToFilename["https://api.zdf.de/tmd/2/ngplayer_2_4/vod/ptmd/mediathek/201218_2330_sendung_zmr"] = "zdf-magazin-royale-stream.json"
	urlToFilename["https://api.zdf.de/tmd/2/ngplayer_2_4/vod/ptmd/mediathek/201211_2300_sendung_zmr"] = "zdf-magazin-royale-stream2.json"

	parameters := defaultParameters
	parameters.MinimumLengthInSeconds = 31 * 60

	result, err := createRssFeedMocked("comedy/zdf-magazin-royale", 2, parameters, urlToFilename, CreateZdfRssFeed)
	if err != nil {
		t.Fatalf("There should not be an error.\n%v", err)
	}
	expectedBytes, _ := ioutil.ReadFile("../testdata/zdf-magazin-royale_min31min.xml")
	expected := string(expectedBytes)

	buildDateReplacement := regexp.MustCompile(`<lastBuildDate>[^<]+</lastBuildDate>`)
	result = buildDateReplacement.ReplaceAllString(result, "<lastBuildDate>NOW</lastBuildDate>")

	if strings.Compare(result, expected) != 0 {
		// ioutil.WriteFile("/tmp/actual.xml", []byte(result), 0644)
		// ioutil.WriteFile("/tmp/expected.xml", []byte(expected), 0644)
		t.Fatalf("The created XML is not as expected. Created:\n%v\n\nExpected:\n%v", result, expected)
	}
}

func createRssFeedMocked(showID string, maxEpisodes int, parameters internal.RequestParameters, urlToFilename map[string](string), fnCreate func(showID string, parameters internal.RequestParameters, zdfAPI *zdfapi.ZDFApi) (result string, err error)) (result string, err error) {
	fnGetHTTP := func(api *zdfapi.ZDFApi, URL string, onlyPeek bool) (result []byte, err error) {
		filename, ok := urlToFilename[URL]
		if !ok {
			err = fmt.Errorf("Unknown URL: %v", URL)
			return
		}
		result, err = ioutil.ReadFile("../testdata/" + filename)
		return
	}
	zdfAPI, err := zdfapi.CreateZDFApiWithFnGet(maxEpisodes, fnGetHTTP)
	if err != nil {
		return
	}
	result, err = fnCreate(showID, parameters, &zdfAPI)
	return
}

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
