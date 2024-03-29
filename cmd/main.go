package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/seiferma/docker_mediathek2rss/internal"
	"github.com/seiferma/docker_mediathek2rss/internal/ardapi"
	"github.com/seiferma/docker_mediathek2rss/internal/ardfeed"
	"github.com/seiferma/docker_mediathek2rss/internal/zdfapi"
	"github.com/seiferma/docker_mediathek2rss/internal/zdffeed"
)

// Constants (maybe make this configurable)
const listenAddress = ":8080"
const cacheDuration = 5 * time.Minute
const maxEpisodes = 50
const ardShowByIDPathPrefix = "/ard/show/"
const zdfShowByPathPrefix = "/zdf/show/byPath/"

// Global state
var feedCache internal.Cache

func main() {
	feedCache = internal.CreateCache(cacheDuration)
	http.HandleFunc(ardShowByIDPathPrefix, ardShowByIDServer)
	http.HandleFunc(zdfShowByPathPrefix, zdfShowByPathServer)
	log.Printf("Starting HTTP server on %v", listenAddress)
	http.ListenAndServe(listenAddress, nil)
}

func ardShowByIDServer(w http.ResponseWriter, r *http.Request) {
	// extract show id from request
	urlSegments := strings.Split(r.URL.Path, "/")
	if len(urlSegments) < 1 {
		return
	}
	showID := urlSegments[len(urlSegments)-1]
	if !isValidArdShowID(showID) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "The given show ID is not valid.")
		log.Print("Received a request for invalid show ID.")
		return
	}

	// extract request parameters
	requestParameters := internal.CreateRequestParametersFromURL(r.URL)
	log.Printf("Received a request for show ID %v with parameters %v.", showID, requestParameters)

	// create ARD API
	ardAPI := ardapi.CreateArdAPI(maxEpisodes)

	// create RSS feed
	fnCreateRss := func(showID string, parameters internal.RequestParameters) (string, error) {
		return ardfeed.CreateArdRssFeed(showID, parameters, &ardAPI)
	}
	rssFeedString, error := internal.CreateRssFeedCached(showID, requestParameters, &feedCache, fnCreateRss)

	// report an error
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, error)
		log.Printf("There was an error while processing request for %v: %v", showID, error)
	}

	// return produced feed
	w.Header().Add("Content-Type", "application/rss+xml")
	fmt.Fprint(w, rssFeedString)
	log.Printf("Successfully returing RSS feed for %v.", showID)
}

func isValidArdShowID(showID string) bool {
	idRegex, _ := regexp.Compile("^[a-zA-Z0-9]+$")
	showIDBytes := []byte(showID)
	return idRegex.Match(showIDBytes)
}

func zdfShowByPathServer(w http.ResponseWriter, r *http.Request) {
	// extract show path from URL
	showPath := strings.Replace(r.URL.Path, zdfShowByPathPrefix, "", -1)
	if !isValidZdfPath(showPath) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "The given show path is not valid.")
		log.Print("Received a request for invalid show path.")
		return
	}

	// extract request parameters
	requestParameters := internal.CreateRequestParametersFromURL(r.URL)
	log.Printf("Received a request for show path %v with parameters %v.", showPath, requestParameters)

	// create ARD API
	zdfAPI, err := zdfapi.CreateZDFApi(maxEpisodes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		log.Printf("There was an error while initializing the ZDF API: %v", err)
	}

	// create RSS feed
	fnCreateRss := func(showPath string, parameters internal.RequestParameters) (string, error) {
		return zdffeed.CreateZdfRssFeed(showPath, parameters, &zdfAPI)
	}
	rssFeedString, err := internal.CreateRssFeedCached(showPath, requestParameters, &feedCache, fnCreateRss)

	// report an error
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		log.Printf("There was an error while processing request for %v: %v", showPath, err)
	}

	// return produced feed
	w.Header().Add("Content-Type", "application/rss+xml")
	fmt.Fprint(w, rssFeedString)
	log.Printf("Successfully returing RSS feed for %v.", showPath)
}

func isValidZdfPath(path string) bool {
	regex := regexp.MustCompile("^([a-zA-Z0-9-]+/)*[a-zA-Z0-9-]+$")
	return regex.Match([]byte(path))
}
