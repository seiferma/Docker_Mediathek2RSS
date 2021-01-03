package zdfapi

import (
	"errors"
	"io/ioutil"
	"testing"
	"time"
)

const maxEpisodes = 2

func TestCreateZDFApi(t *testing.T) {
	api := createAPI(t, map[string](string){
		bearerTokenSourceURL: "../testdata/zdf-heute-journal.html"})
	if api.bearerToken != "playertoken" {
		t.Fatalf("Expected the api token to be %v but got %v.", "playertoken", api.bearerToken)
	}
}

func TestGetShow(t *testing.T) {
	showParam := "comedy/zdf-magazin-royale"
	api := createAPISimple(t, map[string](string){
		"https://api.zdf.de/content/documents/zdf/" + showParam: "../testdata/zdf-magazin-royale.json",
	})
	actualShow, err := api.GetShow(showParam)
	if err != nil {
		t.Fatal("We did not expect an error")
	}

	assertEquals(t, "zdf-magazin-royale-102", actualShow.ID)
	assertEquals(t, "ZDF Magazin Royale", actualShow.Title)
	assertEquals(t, "https://www.zdf.de/comedy/zdf-magazin-royale", actualShow.URL)
	assertEquals(t, "/search/documents/zdf/comedy/zdf-magazin-royale?q=*&limit=0&types=page-video&hasVideo=true", actualShow.Search.SearchURLTemplate)
	assertEquals(t, "ZDF Magazin Royale", actualShow.Image.Alt)
	assertEquals(t, 12, actualShow.Search.ResultCount)
	assertEquals(t, 33, len(actualShow.Image.Images))
	assertEquals(t, 2, len(actualShow.Modules))
}

func TestGetDescription(t *testing.T) {
	show := &Show{
		Modules: []struct {
			Description string "json:\"shorttext-text\""
		}{
			{},
			{Description: ""},
			{Description: "test"},
		},
	}
	assertEquals(t, "test", show.GetDescription())
}

func TestGetShowVideos(t *testing.T) {
	api := createAPISimple(t, map[string](string){
		"https://api.zdf.de/search/documents/zdf/comedy/zdf-magazin-royale?q=*&limit=2&types=page-video&hasVideo=true": "../testdata/zdf-magazin-royale-search.json",
	})
	show := Show{
		Search: struct {
			ResultCount       int    "json:\"totalResultsCount\""
			SearchURLTemplate string "json:\"self\""
		}{
			ResultCount:       50,
			SearchURLTemplate: "/search/documents/zdf/comedy/zdf-magazin-royale?q=*&limit=0&types=page-video&hasVideo=true",
		},
	}
	actual, err := api.GetShowVideos(show)
	if err != nil {
		t.Fatal("We did not expect an error.")
	}
	assertEquals(t, 2, actual.ResultCount)
	assertEquals(t, 2, len(actual.Results))
	assertEquals(t, "Corona-Unternehmer des Jahres ", actual.Results[0].Video.Title)
	assertEquals(t, "2020 ging es Ihnen richtig scheiße? Selber Schuld! Sie könnten eine tödliche Pandemie auch mal als Chance begreifen. ", actual.Results[0].Video.Description)
	expectedTime, _ := time.Parse(time.RFC3339, "2020-12-18T23:30:00.000+01:00")
	assertEqualsTime(t, expectedTime, actual.Results[0].Video.Date)
	assertEquals(t, "https://www.zdf.de/comedy/zdf-magazin-royale/zdf-magazin-royale-106.html", actual.Results[0].Video.URL)
	assertEquals(t, 1888, actual.Results[0].Video.Streams.Streams.Duration)
	assertEquals(t, "/tmd/2/{playerId}/vod/ptmd/mediathek/201218_2330_sendung_zmr", actual.Results[0].Video.Streams.Streams.URLTemplate)
}

func TestGetStream(t *testing.T) {
	api := createAPISimple(t, map[string](string){
		"https://api.zdf.de/tmd/2/" + playerID + "/vod/ptmd/mediathek/201218_2330_sendung_zmr": "../testdata/zdf-magazin-royale-stream.json",
	})
	description := VideoDescription{
		Streams: struct {
			Streams struct {
				Duration    int    "json:\"duration\""
				URLTemplate string "json:\"http://zdf.de/rels/streams/ptmd-template\""
			} "json:\"http://zdf.de/rels/target\""
		}{
			Streams: struct {
				Duration    int    "json:\"duration\""
				URLTemplate string "json:\"http://zdf.de/rels/streams/ptmd-template\""
			}{
				URLTemplate: "/tmd/2/{playerId}/vod/ptmd/mediathek/201218_2330_sendung_zmr",
			},
		},
	}
	stream, err := api.GetStreams(description)
	if err != nil {
		t.Fatal("We did not expect an error.")
	}
	assertEquals(t, 5, len(stream.Streams))
	assertEquals(t, 1, len(stream.Streams[3].Formats))
	actualFormat := stream.Streams[3].Formats[0]
	assertEquals(t, 2, len(actualFormat.Qualities))
	actualQuality := actualFormat.Qualities[1]
	assertEquals(t, 1, len(actualQuality.Audio.Tracks))
	actualTrack := actualQuality.Audio.Tracks[0]
	assertEquals(t, false, actualFormat.IsAdaptive)
	assertEquals(t, "video/mp4", actualFormat.MimeType)
	assertEquals(t, "h264_aac_mp4_http_na_na", actualFormat.Type)
	assertEquals(t, false, actualQuality.IsHD)
	assertEquals(t, "avc1.4d401f, mp4a.40.2", actualQuality.MimeCodec)
	assertEquals(t, "high", actualQuality.Quality)
	assertEquals(t, "deu", actualTrack.Language)
	assertEquals(t, "akamai", actualTrack.CDN)
	assertEquals(t, "https://rodlzdf-a.akamaihd.net/none/zdf/20/12/201218_2330_sendung_zmr/3/201218_2330_sendung_zmr_808k_p11v15.mp4", actualTrack.URL)
}

func TestGetSearchURL(t *testing.T) {
	show := &Show{
		Search: struct {
			ResultCount       int    "json:\"totalResultsCount\""
			SearchURLTemplate string "json:\"self\""
		}{
			SearchURLTemplate: "/foo/bar.json?foo=bar&limit=0&bar=foo",
		},
	}
	assertEquals(t, "https://api.zdf.de/foo/bar.json?foo=bar&limit=42&bar=foo", show.getSearchURL(42))
}

func TestGetStreamURL(t *testing.T) {
	description := &VideoDescription{
		Streams: struct {
			Streams struct {
				Duration    int    "json:\"duration\""
				URLTemplate string "json:\"http://zdf.de/rels/streams/ptmd-template\""
			} "json:\"http://zdf.de/rels/target\""
		}{
			Streams: struct {
				Duration    int    "json:\"duration\""
				URLTemplate string "json:\"http://zdf.de/rels/streams/ptmd-template\""
			}{
				URLTemplate: "/foo/{playerId}/bar.json",
			},
		},
	}
	assertEquals(t, "https://api.zdf.de/foo/"+playerID+"/bar.json", description.getStreamsURL())
}

func createAPI(t *testing.T, urlToFilename map[string](string)) ZDFApi {
	fnGet := createFnGet(t, urlToFilename)
	api, error := CreateZDFApiWithFnGet(maxEpisodes, fnGet)
	if error != nil {
		t.Fatal("We did not expect an error during API creation.")
	}
	return api
}

func createAPISimple(t *testing.T, urlToFilename map[string](string)) ZDFApi {
	fnGet := createFnGet(t, urlToFilename)
	return ZDFApi{
		maxEpisodes: maxEpisodes,
		bearerToken: "empty",
		fnGet:       fnGet,
	}
}

func createFnGet(t *testing.T, urlToFilename map[string](string)) func(api *ZDFApi, URL string, onlyPeek bool) (result []byte, err error) {
	return func(api *ZDFApi, URL string, onlyPeek bool) (result []byte, err error) {
		result = []byte{}
		filePath, ok := urlToFilename[URL]
		if ok {
			if !onlyPeek {
				return ioutil.ReadFile(filePath)
			}
			return
		}

		if onlyPeek {
			result = []byte{}
			err = errors.New("not found")
			return
		}

		t.Fatalf("Unexpected URL %v has been requested.", URL)
		return
	}
}

func assertEqualsTime(t *testing.T, expected, actual time.Time) {
	if !expected.Equal(actual) {
		t.Fatalf("Expected %v but got %v.", expected, actual)
	}
}

func assertEquals(t *testing.T, expected, actual interface{}) {
	if actual != expected {
		t.Fatalf("Expected %v but got %v.", expected, actual)
	}
}
