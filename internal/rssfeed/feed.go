package rssfeed

import (
	"encoding/xml"
	"fmt"
	"time"
)

// Feed represents the root element of an RSS feed.
// It should be created by the CreateFeed function in order to initialize the feed with reasonable default values.
type Feed struct {
	XMLName     xml.Name `xml:"rss"`
	XMLNSItunes string   `xml:"xmlns:itunes,attr"`
	Version     string   `xml:"version,attr"`
	Channel     Channel  `xml:"channel"`
}

// Channel is the second mandatory root element of an RSS feed.
type Channel struct {
	XMLName        xml.Name         `xml:"channel"`
	Title          string           `xml:"title"`
	Description    *FeedDescription `xml:"description"`
	Link           string           `xml:"link,omitempty"`
	LastBuildDate  *time.Time       `xml:"lastBuildDate"`
	Image          *Image           `xml:"image"`
	ITunesSubtitle string           `xml:"itunes:subtitle,omitempty"`
	ITunesAuthor   string           `xml:"itunes:author,omitempty"`
	ITunesSummary  *ITunesSummary   `xml:"itunes:summary"`
	ITunesCategory string           `xml:"itunes:category,omitempty"`
	ITunesImage    *ITunesImage     `xml:"itunes:image"`
	ITunesExplicit bool             `xml:"itunes:explicit"`
	FeedItems      []FeedItem       `xml:"item"`
}

// ITunesSummary is the XML element to represent the itunes:summary element.
type ITunesSummary struct {
	XMLName xml.Name `xml:"itunes:summary"`
	Text    string   `xml:",cdata"`
}

// Image represents the RSS image element.
type Image struct {
	XMLName xml.Name `xml:"image"`
	URL     string   `xml:"url"`
	Title   string   `xml:"title,omitempty"`
	Link    string   `xml:"link,omitempty"`
	Height  int      `xml:"height,omitempty"`
	Width   int      `xml:"width,omitempty"`
}

// ITunesImage represents the RSS image element for itunes.
type ITunesImage struct {
	XMLName xml.Name `xml:"itunes:image"`
	URL     string   `xml:"href,attr"`
}

// FeedItem represents an entry of a Podcast in the RSS feed.
type FeedItem struct {
	XMLName              xml.Name           `xml:"item"`
	Title                string             `xml:"title"`
	Link                 string             `xml:"link,omitempty"`
	Description          *FeedDescription   `xml:"description"`
	PubDate              *time.Time         `xml:"pubDate"`
	GUID                 string             `xml:"guid,omitempty"`
	Enclosure            *FeedItemEnclosure `xml:"enclosure"`
	ITunesDurationString string             `xml:"itunes:duration,omitempty"`
	ITunesTitle          string             `xml:"itunes:title,omitempty"`
	ITunesSubtitle       string             `xml:"itunes:subtitle,omitempty"`
	ITunesSummary        *ItunesSummary     `xml:"itunes:summary"`
	ITunesImage          *ITunesImage       `xml:"itunes:image"`
}

// FeedDescription represents the RSS description element.
type FeedDescription struct {
	XMLName xml.Name `xml:"description"`
	Text    string   `xml:",cdata"`
}

// ItunesSummary represents the RSS summary element of itunes.
type ItunesSummary struct {
	XMLName xml.Name `xml:"itunes:summary"`
	Text    string   `xml:",cdata"`
}

// FeedItemEnclosure represents enclosures (media elements) within the feed items.
type FeedItemEnclosure struct {
	XMLName xml.Name `xml:"enclosure"`
	URL     string   `xml:"url"`
	Type    string   `xml:"type"`
	Length  string   `xml:"length,omitempty"`
}

// CreateFeed creates and initializes the RSS feed.
func CreateFeed() Feed {
	return Feed{
		XMLNSItunes: "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Version:     "2.0",
	}
}

// Serialize serializes the RSS feed to a byte array with proper identation.
func (feed *Feed) Serialize() (result []byte, err error) {
	result, err = xml.MarshalIndent(feed, "", "    ")
	return
}

// SerializeToString serializes the RSS feed to a string with proper identation.
func (feed *Feed) SerializeToString() (result string, err error) {
	var bytes []byte
	bytes, err = feed.Serialize()
	if err == nil {
		result = string(bytes)
	}
	return
}

// CreateItunesDurationStringFromSeconds creates a string of form HH:MM:SS based on a duration given as seconds.
func CreateItunesDurationStringFromSeconds(seconds int) string {
	secondsPart := seconds % 60
	minutesPart := (seconds / 60) % 60
	hoursPart := (seconds / 60 / 60) % 60
	if minutesPart+hoursPart == 0 {
		return fmt.Sprintf("%v", secondsPart)
	}
	if hoursPart == 0 {
		return fmt.Sprintf("%v:%02d", minutesPart, secondsPart)
	}
	return fmt.Sprintf("%v:%02d:%02d", hoursPart, minutesPart, secondsPart)
}
