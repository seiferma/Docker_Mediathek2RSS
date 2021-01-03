package zdfapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const bearerTokenSourceURL = "https://www.zdf.de/nachrichten/heute-journal"
const zdfAPIBase = "https://api.zdf.de"
const showAPIPrefix = zdfAPIBase + "/content/documents/zdf/"
const playerID = "ngplayer_2_4"

// Show holds all information about a show except the corresponding videos.
type Show struct {
	ID     string         `json:"id"`
	Title  string         `json:"title"`
	Image  ZDFTeaserImage `json:"teaserImageRef"`
	URL    string         `json:"http://zdf.de/rels/sharing-url"`
	Search struct {
		ResultCount       int    `json:"totalResultsCount"`
		SearchURLTemplate string `json:"self"`
	} `json:"http://zdf.de/rels/search/page-video-counter-with-video"`
	Modules []struct {
		Description string `json:"shorttext-text"`
	} `json:"module"`
}

// ShowSearchResult holds all information about a video search for a show.
type ShowSearchResult struct {
	ResultCount int `json:"totalResultsCount"`
	Results     []struct {
		Video VideoDescription `json:"http://zdf.de/rels/target"`
	} `json:"http://zdf.de/rels/search/results"`
}

// VideoDescription holds all information about a video except the available streams.
type VideoDescription struct {
	ID          string         `json:"id"`
	Title       string         `json:"teaserHeadline"`
	Description string         `json:"teasertext"`
	Date        time.Time      `json:"editorialDate"`
	Image       ZDFTeaserImage `json:"teaserImageRef"`
	URL         string         `json:"http://zdf.de/rels/sharing-url"`
	Streams     struct {
		Streams struct {
			Duration    int    `json:"duration"`
			URLTemplate string `json:"http://zdf.de/rels/streams/ptmd-template"`
		} `json:"http://zdf.de/rels/target"`
	} `json:"mainVideoContent"`
}

// VideoStreams holds all available streams for a video.
type VideoStreams struct {
	Streams []VideoStream `json:"priorityList"`
}

// VideoStream represents one video stream with various formats.
type VideoStream struct {
	Formats []struct {
		IsAdaptive bool   `json:"isAdaptive"`
		MimeType   string `json:"mimeType"`
		Type       string `json:"type"`
		Qualities  []struct {
			IsHD      bool   `json:"hd"`
			MimeCodec string `json:"mimeCodec"`
			Quality   string `json:"quality"`
			Audio     struct {
				Tracks []struct {
					CDN      string `json:"cdn"`
					Class    string `json:"class"`
					Language string `json:"language"`
					URL      string `json:"uri"`
				} `json:"tracks"`
			} `json:"audio"`
		} `json:"qualities"`
	} `json:"formitaeten"`
}

// ZDFTeaserImage represents an image of a teaser of the ZDF API.
type ZDFTeaserImage struct {
	Alt    string              `json:"altText"`
	Images map[string](string) `json:"layouts"`
}

// ZDFApi holds information for interacting with the ZDF API.
//
// Users should always create this via CreateZDFApi to correctly initialize the API.
type ZDFApi struct {
	maxEpisodes int
	bearerToken string
	fnGet       func(*ZDFApi, string, bool) ([]byte, error)
}

// CreateZDFApi creates and initializes the API.
//
// If the initialization fails, an error will be returned.
func CreateZDFApi(maxEpisodes int) (api ZDFApi, err error) {
	return CreateZDFApiWithFnGet(maxEpisodes, doHTTPGetRequest)
}

// CreateZDFApiWithFnGet creates and initializes the API with a given HTTP GET function.
//
// If the initialization fails, an error will be returned.
func CreateZDFApiWithFnGet(maxEpisodes int, fnGet func(*ZDFApi, string, bool) ([]byte, error)) (api ZDFApi, err error) {
	api = ZDFApi{
		maxEpisodes: maxEpisodes,
		bearerToken: "",
		fnGet:       fnGet,
	}
	err = api.initBearerToken()
	return
}

// GetShow loads a show for a given showPath.
func (api *ZDFApi) GetShow(showPath string) (show Show, err error) {
	requestURL := showAPIPrefix + showPath
	result, err := api.Get(requestURL, false)
	if err != nil {
		return
	}
	err = json.Unmarshal(result, &show)
	return
}

// GetShowVideos loads the videos available for a given show up to a number of maxEpisodes.
func (api *ZDFApi) GetShowVideos(show Show) (searchResult ShowSearchResult, err error) {
	searchURL := show.getSearchURL(api.maxEpisodes)
	result, err := api.Get(searchURL, false)
	if err != nil {
		return
	}
	err = json.Unmarshal(result, &searchResult)
	return
}

// GetStreams loads the information about available video streams for a given video description.
func (api *ZDFApi) GetStreams(description VideoDescription) (stream VideoStreams, err error) {
	streamsURL := description.getStreamsURL()
	result, err := api.Get(streamsURL, false)
	if err != nil {
		return
	}
	err = json.Unmarshal(result, &stream)
	return
}

// GetDescription identifies and returns the description for a show.
// If no description is found, an empty description will be returned.
func (show *Show) GetDescription() string {
	for _, module := range show.Modules {
		if module.Description != "" {
			return module.Description
		}
	}
	return ""
}

// Get performs a HTTP GET request.
//
// The request is made to the URL given as parameter. It returns the content of the page
// or an error in case of problems. If the parameter onlyPeek is set to true, the content
// will always be empty but the error can be used to see if a request would have succeeded.
func (api *ZDFApi) Get(URL string, onlyPeek bool) ([]byte, error) {
	return api.fnGet(api, URL, onlyPeek)
}

func (api *ZDFApi) initBearerToken() error {
	regex := regexp.MustCompile("[\"']?apiToken[\"']?:\\s*[\"']([a-z0-9]+)[\"']")
	mainPageContent, err := api.fnGet(api, bearerTokenSourceURL, false)
	if err != nil {
		return err
	}
	matches := regex.FindSubmatch(mainPageContent)
	if len(matches) != 2 {
		err = errors.New("Could not find bearer token on ZDF main page")
		return err
	}
	token := string(matches[1])
	api.bearerToken = token
	return nil
}

func doHTTPGetRequest(api *ZDFApi, URL string, onlyPeek bool) (result []byte, err error) {
	result = []byte{}
	client := &http.Client{
		CheckRedirect: nil,
	}
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return
	}
	if api.bearerToken != "" {
		req.Header.Add("Api-Auth", "Bearer "+api.bearerToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error during HTTP GET request for URL %v: %v.", URL, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		if !onlyPeek {
			log.Printf("Received HTTP response %v for URL %v.", resp.StatusCode, URL)
		}
		err = fmt.Errorf("Received HTTP status %v", resp.StatusCode)
		return
	}

	if onlyPeek {
		return
	}

	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Could not read body from GET request to URL %v.", URL)
		return
	}
	return
}

func (show *Show) getSearchURL(maxEpisodes int) string {
	limitParameter := fmt.Sprintf("limit=%v", maxEpisodes)
	searchPath := strings.Replace(show.Search.SearchURLTemplate, "limit=0", limitParameter, -1)
	return fmt.Sprintf("%v%v", zdfAPIBase, searchPath)
}

func (description *VideoDescription) getStreamsURL() string {
	streamsPath := strings.Replace(description.Streams.Streams.URLTemplate, "{playerId}", playerID, -1)
	return fmt.Sprintf("%v%v", zdfAPIBase, streamsPath)
}
