package internal

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strings"

	"github.com/seiferma/docker-ard2rss/internal/ardapi"
	"github.com/seiferma/docker-ard2rss/internal/rssfeed"
)

// CreateArdRssFeedCached produces a RSS feed for a show of the ARD Mediathek.
// It takes the ID of the show as requested by the JSON API, the requested media width, a pointer to the ArdAPI and a pointer to the
// cache. It yields the RSS feed as string and an error.
//
// The requested with might not be met perfectly depending on the available media. However, the logic tries to get to the requested
// width as close as possible.
func CreateArdRssFeedCached(showID string, requestedMediaWidth int, ardAPI *ardapi.ArdAPI, cache *Cache) (result string, err error) {
	// directly return valid cache entry
	cacheKey := getCacheKey(showID, requestedMediaWidth)
	var foundCacheEntry bool
	result, foundCacheEntry = cache.GetContent(cacheKey)
	if foundCacheEntry {
		log.Printf("Answering request for %v / %v from cache.", showID, requestedMediaWidth)
		return
	}

	// calculate RSS feed
	result, err = createRssFeed(showID, requestedMediaWidth, ardAPI)

	// cache result
	if err == nil {
		cache.StoreContent(cacheKey, result)
	}
	return
}

func createRssFeed(showID string, requestedMediaWidth int, ardAPI *ardapi.ArdAPI) (result string, err error) {
	var showInitial ardapi.Show
	showInitial, err = ardAPI.GetShow(showID)
	if err != nil {
		return
	}

	feedURL := "https://www.ardmediathek.de/ard/sendung/" + showID
	feedTitle := showInitial.Teasers[0].Show.Title
	feedDescription := showInitial.Teasers[0].Show.LongSynopsis
	feedImage := getFeedImage(showInitial.Teasers[0].Show.Images)
	feedImageURL, feedImageAlt := getFeedImageURLAndAlt(feedImage, requestedMediaWidth)

	feedItems := make([]rssfeed.FeedItem, len(showInitial.Teasers))
	for i, teaser := range showInitial.Teasers {
		mediathekLink := "https://www.ardmediathek.de/ard/video/" + teaser.ID
		videoAPIURL := teaser.Links.Target.Href
		var video ardapi.ShowVideo
		video, err = ardAPI.GetVideoByURL(videoAPIURL)
		if err != nil {
			return
		}
		mediaStreams := video.Widgets[0].MediaCollection.Embedded.MediaArray[0].MediaStreamArray
		synopsis := video.Widgets[0].Synopsis
		videoImage := video.Widgets[0].Image
		videoImageURL, _ := getFeedImageURLAndAlt(videoImage, requestedMediaWidth)

		lastWidth := 0
		var lastURL string
		for _, mediaStream := range mediaStreams {
			newDistance := math.Abs(float64(requestedMediaWidth - mediaStream.Width))
			oldDistance := math.Abs(float64(requestedMediaWidth - lastWidth))
			if strings.Contains(mediaStream.Stream, "mp4") && newDistance < oldDistance {
				lastWidth = mediaStream.Width
				lastURL = mediaStream.Stream
			}
		}

		feedItems[i] = rssfeed.FeedItem{
			Title:                teaser.LongTitle,
			Description:          &rssfeed.FeedDescription{Text: synopsis},
			PubDate:              &teaser.BroadcastedOn,
			GUID:                 teaser.ID,
			Link:                 mediathekLink,
			ITunesTitle:          teaser.LongTitle,
			ITunesSummary:        &rssfeed.ItunesSummary{Text: synopsis},
			ITunesDurationString: rssfeed.CreateItunesDurationStringFromSeconds(teaser.Duration),
			ITunesImage: &rssfeed.ITunesImage{
				URL: videoImageURL,
			},
			Enclosure: &rssfeed.FeedItemEnclosure{
				URL:  lastURL,
				Type: "video/mp4",
			},
		}
	}

	feed := rssfeed.CreateFeed()
	feed.Channel = rssfeed.Channel{
		Title:       feedTitle,
		Link:        feedURL,
		Description: &rssfeed.FeedDescription{Text: feedDescription},
		Image: &rssfeed.Image{
			URL:   feedImageURL,
			Title: feedImageAlt,
		},
		FeedItems: feedItems,
	}
	result, err = feed.SerializeToString()
	return
}

func getFeedImage(feedImageCandidates map[string](ardapi.ShowImage)) ardapi.ShowImage {
	var feedImageCandidate ardapi.ShowImage
	for k, v := range feedImageCandidates {
		feedImageCandidate = v
		if k == "16x9" {
			break
		}
	}
	return feedImageCandidate
}

func getFeedImageURLAndAlt(img ardapi.ShowImage, requestedMediaWidth int) (url, alt string) {
	url = strings.Replace(img.Src, "{width}", toString(requestedMediaWidth), -1)
	alt = img.Alt
	return
}

// DoGetRequest performs a simple HTTP GET request to the given URL
func DoGetRequest(URL string) (result []byte, err error) {
	var resp *http.Response
	resp, err = http.Get(URL)
	if err != nil {
		log.Fatalf("Received HTTP response %v for URL %v.", resp.StatusCode, URL)
		return
	}
	defer resp.Body.Close()

	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Could not read body from GET request to URL %v.", URL)
		return
	}
	return
}

func toString(input interface{}) string {
	return fmt.Sprintf("%v", input)
}

func getCacheKey(showID string, requestedWidth int) string {
	return fmt.Sprintf("%v-%v", showID, requestedWidth)
}
