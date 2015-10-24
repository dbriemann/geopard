package geopard

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const (
	GOOGLE_GEOCODE_BASE_URL = "https://maps.googleapis.com/maps/api/geocode/json?"
)

var (
	googleApiKey = ""
	googleLang   = "en"

	ErrZeroResults    = errors.New("zero results")
	ErrOverLimit      = errors.New("over query limit")
	ErrRequestDenied  = errors.New("request denied")
	ErrInvalidRequest = errors.New("invalid request")
	ErrUnknown        = errors.New("unkown error")
)

//The following structs are for parsing the json response from
//the google geocoding service.
type (
	GResponse struct {
		Status  string    `json:"status"`
		Results []GResult `json:"results"`
	}
	GResult struct {
		PlaceId        string           `json:"place_id"`
		FormattedAddr  string           `json:"formatted_address"`
		Geometry       GGeometry        `json:"geometry"`
		PartialMatch   bool             `json:"partial_match"`
		AddrComponents []GAddrComponent `json:"address_components"`
		Types          []string         `json:"types"`
	}
	GGeometry struct {
		Location     GPoint `json:"location"`
		Viewport     GArea  `json:"viewport"`
		Bounds       GArea  `json:"bounds"`
		LocationType string `json:"location_type"`
	}
	GPoint struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	GArea struct {
		NorthEast GPoint `json:"northeast"`
		SouthWest GPoint `json:"southwest"`
	}
	GAddrComponent struct {
		Long  string   `json:"long_name"`
		Short string   `json:"short_name"`
		Types []string `json:"types"`
	}
)

//SetGoogleApiKey sets the api key for Google geocoding services.
//This key is not needed and may be omitted. Howevery usage limits
//will then be enforced via IP.
func SetGoogleApiKey(key string) {
	googleApiKey = key
}

//SetGoogleLanguage sets the language for the responses of the
//geocoding service. For a list of supported languages check:
//https://developers.google.com/maps/faq#languagesupport
//This library uses english(en) language and formatting as default.
func SetGoogleLanguage(l string) {
	googleLang = l
}

//GoogleGeocode returns a Location object for the given address string.
//The address string should be in the format used by the national
//postal service of the country concerned.
func GoogleGeocode(address string) (GResponse, error) {
	response := GResponse{}

	//query google service for a json response
	url := GOOGLE_GEOCODE_BASE_URL +
		"address=" + url.QueryEscape(address) +
		"&language=" + googleLang +
		"&key=" + googleApiKey
	resp, err := http.Get(url)

	if err != nil {
		return response, fmt.Errorf("geocoding address '%s' failed with error '%v'", address, err)
	}

	defer resp.Body.Close()

	//parse json response into temporary struct
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, fmt.Errorf("parsing geocoding result for address '%s' failed with error '%v'", address, err)
	}

	switch response.Status {
	case "OK":
		break
	case "ZERO_RESULTS":
		return response, fmt.Errorf("google replied with error '%v' for address '%s'", ErrZeroResults, address)
	case "OVER_QUERY_LIMIT":
		return response, fmt.Errorf("google replied with error '%v' for address '%s'", ErrOverLimit, address)
	case "REQUEST_DENIED":
		return response, fmt.Errorf("google replied with error '%v' for address '%s'", ErrRequestDenied, address)
	case "INVALID_REQUEST":
		return response, fmt.Errorf("google replied with error '%v' for address '%s'", ErrInvalidRequest, address)
	case "UNKOWN_ERROR":
		return response, fmt.Errorf("google replied with error '%v' for address '%s'", ErrUnknown, address)
	}

	return response, nil
}
