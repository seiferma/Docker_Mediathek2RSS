package rssfeed

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestExampleFeed(t *testing.T) {
	feed := CreateFeed()
	feed.Channel = Channel{
		Title: "Test",
		Description: &FeedDescription{
			Text: "Some long description.",
		},
		Image: &Image{
			URL: "http://foo.bar/foo.png",
		},
		ITunesImage: &ITunesImage{
			URL: "http://foo.bar/foo.png",
		},
		FeedItems: []FeedItem{
			{
				Title: "TestItem",
				Description: &FeedDescription{
					Text: "Some long description.",
				},
				Link: "http://foo.bar/item.html",
				Enclosure: &FeedItemEnclosure{
					URL:  "http://foo.bar/item.mp4",
					Type: "video/mp4",
				},
				ITunesTitle: "TestItem",
				ITunesSummary: &ItunesSummary{
					Text: "Some long description.",
				},
				ITunesDurationString: CreateItunesDurationStringFromSeconds(422),
				GUID:                 "foo-bar-uuid",
				ITunesImage: &ITunesImage{
					URL: "http://foo.bar/item.png",
				},
			},
			{
				Title: "TestItem2",
				Description: &FeedDescription{
					Text: "Some long description2.",
				},
				Link: "http://foo.bar/item2.html",
			},
		},
	}
	result, err := feed.SerializeToString()
	if err != nil {
		t.Fatal("There should be no error")
	}

	expectedBytes, err2 := ioutil.ReadFile("testdata/example_rss.xml")
	if err2 != nil {
		t.Fatal("There should be no error")
	}
	expected := string(expectedBytes)

	if strings.Compare(expected, result) != 0 {
		t.Fatal("There were differences between the expected and the actual XML string")
		t.Logf("\n%v\n", string(result))
	}

}

func TestCreateItunesDurationString(t *testing.T) {
	assertItunesDurationString(t, 5, "5")
	assertItunesDurationString(t, 59, "59")
	assertItunesDurationString(t, 60, "1:00")
	assertItunesDurationString(t, 10*60+5, "10:05")
	assertItunesDurationString(t, 1*60*60, "1:00:00")
	assertItunesDurationString(t, 1*60*60+1*60+1, "1:01:01")
}

func assertItunesDurationString(t *testing.T, seconds int, expected string) {
	actual := CreateItunesDurationStringFromSeconds(seconds)
	if strings.Compare(expected, actual) != 0 {
		t.Fatalf("Expected the duration string for %v seconds to be %v and not %v.", seconds, expected, actual)
	}
}
