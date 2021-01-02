package internal

import (
	"regexp"
	"strconv"
	"time"

	"github.com/seiferma/docker-ard2rss/internal/rssfeed"
	"github.com/seiferma/docker-ard2rss/internal/zdfapi"
)

const wantedMimeType = "video/mp4"

// CreateZdfRssFeed creates an RSS feed for a given showPath and a requested media width. The ZDFApi has to be passed as well.
func CreateZdfRssFeed(showPath string, requestedMediaWidth int, api *zdfapi.ZDFApi) (result string, err error) {
	var show zdfapi.Show
	show, err = api.GetShow(showPath)
	if err != nil {
		return
	}
	var searchResults zdfapi.ShowSearchResult
	searchResults, err = api.GetShowVideos(show)
	if err != nil {
		return
	}

	feed := rssfeed.CreateFeed()
	feed.Channel.Title = show.Title
	feed.Channel.ITunesSubtitle = feed.Channel.Title
	feed.Channel.Description = &rssfeed.FeedDescription{
		Text: show.GetDescription(),
	}
	feed.Channel.ITunesSummary = &rssfeed.ITunesSummary{
		Text: feed.Channel.Description.Text,
	}
	feed.Channel.Link = show.URL
	feed.Channel.Image = &rssfeed.Image{
		Title: show.Image.Alt,
		Link:  show.URL,
		URL:   findBestMatchingImageURL(&show.Image),
	}
	feed.Channel.ITunesImage = &rssfeed.ITunesImage{
		URL: feed.Channel.Image.URL,
	}
	now := time.Now()
	feed.Channel.LastBuildDate = &now
	feed.Channel.FeedItems = make([]rssfeed.FeedItem, searchResults.ResultCount)

	for i, result := range searchResults.Results {
		var streams zdfapi.VideoStreams
		streams, err = api.GetStreams(result.Video)
		videoURL := findBestMatchingVideoStreamURL(&streams)

		feed.Channel.FeedItems[i] = rssfeed.FeedItem{}
		item := &feed.Channel.FeedItems[i]

		item.Title = result.Video.Title
		item.ITunesTitle = item.Title
		item.Description = &rssfeed.FeedDescription{
			Text: result.Video.Description,
		}
		item.ITunesSummary = &rssfeed.ItunesSummary{
			Text: item.Description.Text,
		}
		item.GUID = &rssfeed.FeedGUID{
			Text: result.Video.ID,
		}
		item.ITunesDurationString = rssfeed.CreateItunesDurationStringFromSeconds(result.Video.Streams.Streams.Duration)
		item.PubDate = &result.Video.Date
		item.ITunesImage = &rssfeed.ITunesImage{
			URL: findBestMatchingImageURL(&result.Video.Image),
		}
		item.Link = result.Video.URL
		item.Enclosure = &rssfeed.FeedItemEnclosure{
			URL:  videoURL,
			Type: wantedMimeType,
		}
	}

	result, err = feed.SerializeToString()
	return
}

func findBestMatchingImageURL(images *zdfapi.ZDFTeaserImage) string {
	biggestArea := 0
	bestURL := ""
	for resolution, URL := range images.Images {
		width, height := getDimensions(resolution)
		area := width * height
		if area > biggestArea {
			biggestArea = area
			bestURL = URL
		}
	}
	return bestURL
}

func getDimensions(resolution string) (width, height int) {
	width = 0
	height = 0
	regex := regexp.MustCompile("([0-9]+)x([0-9]+)")
	matches := regex.FindSubmatch([]byte(resolution))
	if len(matches) != 3 {
		return
	}

	foundWidth, err := strconv.ParseInt(string(matches[1]), 10, 0)
	if err != nil {
		return
	}
	foundHeight, err := strconv.ParseInt(string(matches[2]), 10, 0)
	if err != nil {
		return
	}
	width = int(foundWidth)
	height = int(foundHeight)
	return
}

func findBestMatchingVideoStreamURL(streams *zdfapi.VideoStreams) string {
	const adaptive = false
	const mimeType = wantedMimeType
	const lang = "deu"
	const class = "main"

	qualityToURL := map[string](string){}
	for _, stream := range streams.Streams {
		for _, format := range stream.Formats {
			if adaptive != format.IsAdaptive || mimeType != format.MimeType {
				continue
			}
			for _, quality := range format.Qualities {
				qualityString := quality.Quality
				for _, track := range quality.Audio.Tracks {
					if lang != track.Language || class != track.Class {
						continue
					}
					qualityToURL[qualityString] = track.URL
				}
			}
		}
	}

	url, ok := qualityToURL["veryhigh"]
	if ok {
		return url
	}

	url, ok = qualityToURL["high"]
	if ok {
		return url
	}

	url, ok = qualityToURL["low"]
	if ok {
		return url
	}

	for _, value := range qualityToURL {
		return value
	}

	return ""
}
