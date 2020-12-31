package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/seiferma/docker-ard2rss/internal"
	"github.com/seiferma/docker-ard2rss/internal/rssfeed"
)

// Constants (maybe make this configurable)
const listenAddress = ":8080"
const defaultMediaWidth = 1920
const cacheDuration = 5 * time.Minute
const maxEpisodes = 50

// Global state
var feedCache internal.Cache

func main() {
	feedCache = internal.CreateCache(cacheDuration)
	http.HandleFunc("/show/id/", showByIDServer)
	log.Printf("Starting HTTP server on %v", listenAddress)
	http.ListenAndServe(listenAddress, nil)
}

func showByIDServer(w http.ResponseWriter, r *http.Request) {
	// extract show id from request
	urlSegments := strings.Split(r.URL.Path, "/")
	if len(urlSegments) < 1 {
		return
	}
	showID := urlSegments[len(urlSegments)-1]
	if !isValidShowID(showID) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "The given show ID is not valid.")
		log.Print("Received a request for invalid show ID.")
		return
	}
	requestedMediaWidth := defaultMediaWidth
	widthParameter := r.URL.Query().Get("width")
	if widthParameter != "" {
		tmp, err := strconv.ParseInt(widthParameter, 10, 0)
		if err == nil {
			requestedMediaWidth = int(tmp)
		}
	}
	log.Printf("Received a request for show ID %v with width %v.", showID, requestedMediaWidth)

	// create ARD API
	ardAPI := internal.CreateArdAPI(maxEpisodes, doGetRequest)

	// create rss feed
	rssFeedString, error := createRssFeedCached(showID, requestedMediaWidth, &ardAPI, &feedCache)

	// report an error
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, error)
		log.Fatalf("There was an error while processing request for %v: %v", showID, error)
	}

	// return produced feed
	w.Header().Add("Content-Type", "application/rss+xml")
	fmt.Fprint(w, rssFeedString)
	log.Printf("Successfully returing RSS feed for %v.", showID)
}

func createRssFeedCached(showID string, requestedMediaWidth int, ardAPI *internal.ArdAPI, cache *internal.Cache) (result string, err error) {
	// directly return valid cache entry
	cacheKey := getCacheKey(showID, requestedMediaWidth)
	var foundCacheEntry bool
	result, foundCacheEntry = cache.GetContent(cacheKey)
	if foundCacheEntry {
		log.Printf("Answering request for %v from cache.", showID)
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

func createRssFeed(showID string, requestedMediaWidth int, ardAPI *internal.ArdAPI) (result string, err error) {
	var showInitial internal.Show
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
		var video internal.ShowVideo
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

func isValidShowID(showID string) bool {
	idRegex, _ := regexp.Compile("^[a-zA-Z0-9]+$")
	showIDBytes := []byte(showID)
	return idRegex.Match(showIDBytes)
}

func getFeedImage(feedImageCandidates map[string](internal.ShowImage)) internal.ShowImage {
	var feedImageCandidate internal.ShowImage
	for k, v := range feedImageCandidates {
		feedImageCandidate = v
		if k == "16x9" {
			break
		}
	}
	return feedImageCandidate
}

func getFeedImageURLAndAlt(img internal.ShowImage, requestedMediaWidth int) (url, alt string) {
	url = strings.Replace(img.Src, "{width}", toString(requestedMediaWidth), -1)
	alt = img.Alt
	return
}

func doGetRequest(URL string) (result []byte, err error) {
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
