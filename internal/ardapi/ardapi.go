package ardapi

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const funkDomainId = 741
const funkDomainHash = "CA4SDGOBTRM421IRNO0"

// ArdAPI gives access to various operations of the ARD Mediathek API.
// Its main purpose is to hold configuration parameters and provide them to
// the API functions.
type ArdAPI struct {
	maxEpisodes   int
	fnGetRequest  func(string) ([]byte, error)
	fnPostRequest func(string, url.Values, map[string]string) ([]byte, error)
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
			ID           string `json:"id"`
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
	Tracking struct {
		AtiCustomVars struct {
			Channel    string
			MetadataId string
		}
	}
	Widgets []struct {
		MediaCollection struct {
			Embedded struct {
				MediaArray []struct {
					MediaStreamArray *[]MediaStreamArray `json:"_mediaStreamArray"`
				} `json:"_mediaArray"`
			}
		}
		Image    ShowImage
		Synopsis string
	}
}

type MediaStreamArray struct {
	CDN    string  `json:"_cdn"`
	Width  int     `json:"_width"`
	Height int     `json:"_height"`
	Stream Streams `json:"_stream"`
}

// Streams represents an array of stream URLs.
type Streams struct {
	StreamUrls []string
}

type NexxInitSessionResponse struct {
	Result struct {
		General struct {
			Cid string
		}
	}
}

type NexxVideoMetadata struct {
	Result struct {
		Streamdata struct {
			CdnType               string
			CdnShieldHTTPS        string
			QAccount              string
			QPrefix               string
			QLocator              string
			AzureFileDistribution string
		}
	}
}

// Custom logic to unmarshal a stream array from JSON.
// The function creates an array of stream URLs no matter if JSON contains an array of stream URLs or just one single URL.
func (s *Streams) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("no data about streams available")
	}

	if data[0] == '[' {
		json.Unmarshal(data, &s.StreamUrls)
	} else {
		var streamUrl string
		json.Unmarshal(data, &streamUrl)
		s.StreamUrls = append(s.StreamUrls, streamUrl)
	}

	return nil
}

// CreateArdAPI creates a new API instance taking configuration values to be considered when working with the API.
// The maxEpisodes parameter defines how many episodes of a show shall be received at most.
func CreateArdAPI(maxEpisodes int) ArdAPI {
	return CreateArdAPIWithGetFunc(maxEpisodes, doGetRequest, doPostRequest)
}

// CreateArdAPIWithGetFunc creates a new API instance taking configuration values to be considered when working with the API.
// The maxEpisodes parameter defines how many episodes of a show shall be received at most.
// The fnGetRequest parameter provides a function that carries out a get request and provides the body as byte array.
func CreateArdAPIWithGetFunc(maxEpisodes int, fnGetRequest func(string) ([]byte, error), fnPostRequest func(string, url.Values, map[string]string) ([]byte, error)) ArdAPI {
	return ArdAPI{
		maxEpisodes:   maxEpisodes,
		fnGetRequest:  fnGetRequest,
		fnPostRequest: fnPostRequest,
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
		log.Printf("Could not parse JSON body for request to URL %v. %v", showURL, err)
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
		log.Printf("Could not parse JSON body for request to URL %v. %v", videoURL, err)
		return
	}

	if result.Tracking.AtiCustomVars.Channel == "funk" {
		re := regexp.MustCompile("video/([0-9]+)$")
		match := re.FindStringSubmatch(result.Tracking.AtiCustomVars.MetadataId)
		if match == nil {
			return
		}
		funkVideoId := match[1]

		err = api.replaceMediaStreamArrayUsingFunkVideoId(funkVideoId, &result)
		if err != nil {
			log.Printf("Could not replace media stream by funk videos. %v", err)
			err = nil
		}
	}

	return
}

func (api *ArdAPI) replaceMediaStreamArrayUsingFunkVideoId(id string, showVideo *ShowVideo) (err error) {
	// there are already streams available, so we should not try to replace anything here
	if showVideo.getNumberOfNonAdaptiveStreams() != 0 {
		return
	}

	// get cid, which is required for getting metadata
	cid, err := api.initializeNexxSession()
	if err != nil {
		return
	}

	// get metadata about video
	videoMetadata, err := api.getNexxVideoMetadata(id, cid)
	if err != nil {
		return
	}

	// video url (prefix of all is result.streamdata)
	// https://[cdnShieldHTTPS]/[qAccount]/files/[qPrefix]/[qLocator]/[azureFileDistribution].mp4
	// example of azureFileDistribution: (3001:1280x720:2-w7FWPztXm9hnZpbjMcKq,4152:1920x1080:1-Nf3kJwdpTrgHGLY2jtPv)

	// parse file distribution data
	re := regexp.MustCompile("[0-9]+:([0-9]+)x([0-9]+):([^:,]+)")
	matches := re.FindAllStringSubmatch(videoMetadata.Result.Streamdata.AzureFileDistribution, -1)
	if matches == nil {
		return
	}

	// build new media stream array
	newMediaStreamArray := make([]MediaStreamArray, len(matches))
	for i, match := range matches {
		width, _ := strconv.Atoi(match[1])
		height, _ := strconv.Atoi(match[2])
		selector := match[3]
		videoUrl := fmt.Sprintf("https://%v%v/files/%v/%v/%v.mp4",
			videoMetadata.Result.Streamdata.CdnShieldHTTPS,
			videoMetadata.Result.Streamdata.QAccount,
			videoMetadata.Result.Streamdata.QPrefix,
			videoMetadata.Result.Streamdata.QLocator,
			selector)
		newMediaStreamArray[i] = MediaStreamArray{
			CDN:    videoMetadata.Result.Streamdata.CdnType,
			Width:  width,
			Height: height,
			Stream: Streams{
				StreamUrls: []string{
					videoUrl,
				},
			},
		}
	}

	// replace media stream array in existing show video
	for i, widget := range showVideo.Widgets {
		for j := 0; j < len(widget.MediaCollection.Embedded.MediaArray); j++ {
			showVideo.Widgets[i].MediaCollection.Embedded.MediaArray[j].MediaStreamArray = &newMediaStreamArray
		}
	}

	return
}

func (showVideo *ShowVideo) getNumberOfNonAdaptiveStreams() int {
	nonAdaptiveStreams := 0
	for _, widget := range showVideo.Widgets {
		for _, mediaArray := range widget.MediaCollection.Embedded.MediaArray {
			arrayPtr := mediaArray.MediaStreamArray
			if arrayPtr != nil {
				for _, mediaStreamArray := range *arrayPtr {
					if mediaStreamArray.Width != 0 {
						nonAdaptiveStreams++
					}
				}

			}

		}
	}
	return nonAdaptiveStreams
}

func (api *ArdAPI) initializeNexxSession() (cid string, err error) {
	currentTime := time.Now().Unix()
	randomNumber := 10000 + rand.Intn(90000)
	deviceId := fmt.Sprintf("%v:%v", currentTime, randomNumber)

	postData := url.Values{
		"nxp_devh": {deviceId},
	}
	URL := fmt.Sprintf("https://api.nexx.cloud/v3/%v/session/init", funkDomainId)
	body, err := api.fnPostRequest(URL, postData, map[string]string{})
	if err != nil {
		return
	}
	var initSessionResponse NexxInitSessionResponse
	err = json.Unmarshal(body, &initSessionResponse)
	if err != nil {
		return
	}
	if initSessionResponse.Result.General.Cid == "" {
		err = errors.New("No cid from initializing session.")
		return
	}

	cid = initSessionResponse.Result.General.Cid
	return
}

func (api *ArdAPI) getNexxVideoMetadata(videoId, cid string) (result NexxVideoMetadata, err error) {
	postData := url.Values{
		"addStatusDetails": {"1"},
		"addStreamDetails": {"1"},
		"addFeatures":      {"1"},
		"addCaptions":      {"1"},
		"addBumpers":       {"1"},
		"captionFormat":    {"data"},
	}

	requestTokenValue := getMD5Hash(fmt.Sprintf("%v%v%v", "byid", funkDomainId, funkDomainHash))
	headers := map[string]string{
		"x-request-cid":   cid,
		"x-request-token": requestTokenValue,
	}
	URL := fmt.Sprintf("https://api.nexx.cloud/v3/%v/videos/byid/%v", funkDomainId, videoId)
	body, err := api.fnPostRequest(URL, postData, headers)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &result)
	return
}

func getMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
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
		log.Printf("Received error for URL %v: %v", URL, err)
		return
	}
	defer resp.Body.Close()

	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Could not read body from GET request to URL %v.", URL)
		return
	}
	return
}

func doPostRequest(URL string, data url.Values, headers map[string]string) (result []byte, err error) {
	var resp *http.Response

	postData := strings.NewReader(data.Encode())
	req, err := http.NewRequest("POST", URL, postData)
	if err != nil {
		log.Printf("Error in building POST request for URL %v: %v", URL, err)
		return
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Printf("Received error for URL %v: %v", URL, err)
		return
	}
	defer resp.Body.Close()

	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Could not read body from POST request to URL %v.", URL)
		return
	}
	return
}
