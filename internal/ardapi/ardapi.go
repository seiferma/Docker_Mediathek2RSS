package ardapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// ArdAPI gives access to various operations of the ARD Mediathek API.
// Its main purpose is to hold configuration parameters and provide them to
// the API functions.
type ArdAPI struct {
	maxEpisodes  int
	fnGetRequest func(string) ([]byte, error)
}

// ShowImage represents an image DTO from the API.
type ShowImage struct {
	Title       string
	Src         string
	AspectRatio string
	Alt         string
}

// Show represents a DTO for a show from the API.
type Show struct {
	Teasers []struct {
		LongTitle string
		Links     struct {
			Target struct {
				Href string
			}
		}
		Show struct {
			Title        string
			LongSynopsis string
			Images       map[string](ShowImage)
		}
		Images        map[string](ShowImage)
		BroadcastedOn time.Time
		Duration      int
		ID            string `json:"id"`
	}
}

// ShowVideo represents a DTO for a video of a show from the API.
type ShowVideo struct {
	Widgets []struct {
		MediaCollection struct {
			Embedded struct {
				MediaArray []struct {
					MediaStreamArray []struct {
						CDN    string `json:"_cdn"`
						Width  int    `json:"_width"`
						Height int    `json:"_height"`
						Stream string `json:"_stream"`
					} `json:"_mediaStreamArray"`
				} `json:"_mediaArray"`
			}
		}
		Image    ShowImage
		Synopsis string
	}
}

// CreateArdAPI creates a new API instance taking configuration values to be considered when working with the API.
// The maxEpisodes parameter defines how many episodes of a show shall be received at most.
func CreateArdAPI(maxEpisodes int) ArdAPI {
	return CreateArdAPIWithGetFunc(maxEpisodes, doGetRequest)
}

// CreateArdAPIWithGetFunc creates a new API instance taking configuration values to be considered when working with the API.
// The maxEpisodes parameter defines how many episodes of a show shall be received at most.
// The fnGetRequest parameter provides a function that carries out a get request and provides the body as byte array.
func CreateArdAPIWithGetFunc(maxEpisodes int, fnGetRequest func(string) ([]byte, error)) ArdAPI {
	return ArdAPI{
		maxEpisodes:  maxEpisodes,
		fnGetRequest: fnGetRequest,
	}
}

// GetShow retrieves a show from the API by the given showID.
func (api *ArdAPI) GetShow(showID string) (result Show, err error) {
	showURL := fmt.Sprintf("https://api.ardmediathek.de/page-gateway/widgets/ard/asset/%v?pageNumber=0&pageSize=%v", showID, api.maxEpisodes)
	var body []byte
	body, err = api.fnGetRequest(showURL)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatalf("Could not parse JSON body for request to URL %v. %v", showURL, err)
		return
	}

	if !result.hasAtLeastOneValidTeaser() {
		err = errors.New("The show has no valid teasers")
	}

	return
}

// GetVideoByURL retrieves a video from the given API URL.
func (api *ArdAPI) GetVideoByURL(videoURL string) (result ShowVideo, err error) {
	var body []byte
	body, err = api.fnGetRequest(videoURL)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatalf("Could not parse JSON body for request to URL %v. %v", videoURL, err)
		return
	}
	return
}

func (show Show) hasAtLeastOneValidTeaser() bool {
	if len(show.Teasers) < 1 {
		return false
	}
	teaser := show.Teasers[0]

	if len(teaser.Show.Title) < 1 {
		return false
	}

	if len(teaser.Show.Images) < 1 {
		return false
	}

	return true
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
