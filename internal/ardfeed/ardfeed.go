package ardfeed

import (
	"fmt"
	"math"
	"strings"

	"github.com/seiferma/docker_mediathek2rss/internal/ardapi"
	"github.com/seiferma/docker_mediathek2rss/internal/rssfeed"
)

// CreateArdRssFeed creates an RSS feed for an ARD show.
//
// It takes the ID of the show, a requested media width and the ARD API to use. It yields the feed as a string.
// The effective media width might not perfectly match the requested media width but tries to get as close as possible.
func CreateArdRssFeed(showID string, requestedMediaWidth int, ardAPI *ardapi.ArdAPI) (result string, err error) {
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
			Title:       teaser.LongTitle,
			Description: &rssfeed.FeedDescription{Text: synopsis},
			PubDate:     &teaser.BroadcastedOn,
			GUID: &rssfeed.FeedGUID{
				Text: teaser.ID,
			},
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

func toString(input interface{}) string {
	return fmt.Sprintf("%v", input)
}
