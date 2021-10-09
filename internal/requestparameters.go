package internal

import (
	"net/url"
	"strconv"
)

const defaultMediaWidth = 1920
const defaultMinLengthInSeconds = 0

type RequestParameters struct {
	Width                  int
	MinimumLengthInSeconds int
}

func CreateRequestParametersFromURL(URL *url.URL) RequestParameters {
	return RequestParameters{
		Width:                  getRequestedWidth(URL),
		MinimumLengthInSeconds: getRequestedMinimumLength(URL),
	}
}

func getRequestedWidth(URL *url.URL) int {
	return getRequestedIntegerParameter(URL, "width", defaultMediaWidth)
}

func getRequestedMinimumLength(URL *url.URL) int {
	return getRequestedIntegerParameter(URL, "minLength", defaultMinLengthInSeconds)
}

func getRequestedIntegerParameter(URL *url.URL, parameterName string, defaultValue int) int {
	result := defaultValue
	parameterValue := URL.Query().Get(parameterName)
	if parameterValue != "" {
		tmp, err := strconv.ParseInt(parameterValue, 10, 0)
		if err == nil {
			result = int(tmp)
		}
	}
	return result
}
