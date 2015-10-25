package geopard

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	BASE_URL = "https://maps.googleapis.com/maps/api/geocode/json?"
)

var (
	once     sync.Once
	instance *requestProcessor

	ErrZeroResults    = errors.New("zero results")
	ErrOverLimit      = errors.New("over query limit")
	ErrRequestDenied  = errors.New("request denied")
	ErrInvalidRequest = errors.New("invalid request")
	ErrUnknown        = errors.New("unkown error")
)

//Options contains all required data to create an instance of the request
//processor singleton. Creating an instance with the Instance(..) method
//and leaving the Options object uninitialized will use default options.
type Options struct {
	//ApiKey contains the api key for Google geocoding services.
	//This key is not needed and may be omitted. However usage limits
	//will then be enforced via IP.
	//See: https://developers.google.com/maps/documentation/geocoding/get-api-key
	ApiKey string

	//Lang is the language used for the responses of the
	//geocoding service. For a list of supported languages check:
	//https://developers.google.com/maps/faq#languagesupport
	//This library uses english(en) language and formatting as default.
	Lang string

	//There is a usage limit of 10 requests / second for the google
	//geocoding api. This value usually should not be changed.
	//See: https://developers.google.com/maps/documentation/geocoding/usage-limits
	MaxQueriesPerSec int
}

//GetInstance is a stub method for creating an instance of the request
//processor with default options. In case the singleton already exists
//the instance will just be returned.
func GetInstance() *requestProcessor {
	return Instance(Options{})
}

//Instance creates a request processor instance or returns the instance
//if it already exists. The Options object will only be used for creating
//a new instance.
func Instance(opts Options) *requestProcessor {
	once.Do(func() {
		instance = &requestProcessor{
			apiKey:           opts.ApiKey,
			lang:             "en",
			maxQueriesPerSec: 10,
		}
		if opts.Lang != "" {
			instance.lang = opts.Lang
		}
		if opts.MaxQueriesPerSec > 0 {
			instance.maxQueriesPerSec = opts.MaxQueriesPerSec
		}

		//init the request throttling
		instance.quit = make(chan int)
		instance.throttle = make(chan int, instance.maxQueriesPerSec)
		//allow requests for first time so we don't have to wait for the ticker period
		instance.allowRequests()
		instance.ticker = time.NewTicker(5 * time.Second)
		go instance.multiTick()
	})
	return instance
}

func (r *requestProcessor) Destroy() {
	close(r.quit)
	close(r.throttle)
}

type requestProcessor struct {
	apiKey           string
	lang             string
	maxQueriesPerSec int
	throttle         chan int
	quit             chan int
	ticker           *time.Ticker
}

func (r *requestProcessor) allowRequests() {
	for i := 1; i <= r.maxQueriesPerSec; i++ {
		r.throttle <- i
	}
}

func (r *requestProcessor) multiTick() {
	for {
		select {
		case <-r.quit:
			r.ticker.Stop()
			return
		case <-r.ticker.C:
			r.allowRequests()
		}
	}
}

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

func (r *requestProcessor) processRequest(url string) (GResponse, error) {
	response := GResponse{}

	//wait for throttling to give green light
	//this will block until there are 'free' slots for requests
	<-r.throttle
	//then send request
	resp, err := http.Get(url)

	if err != nil {
		return response, err
	}

	defer resp.Body.Close()

	//parse json response into temporary struct
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, err
	}

	switch response.Status {
	case "OK":
		break
	case "ZERO_RESULTS":
		return response, ErrZeroResults
	case "OVER_QUERY_LIMIT":
		return response, ErrOverLimit
	case "REQUEST_DENIED":
		return response, ErrRequestDenied
	case "INVALID_REQUEST":
		return response, ErrInvalidRequest
	case "UNKOWN_ERROR":
		return response, ErrUnknown
	}

	return response, nil
}

//ReverseGeocode returns a GResponse object for the given latitude, longitude pair.
//It contains all information offered by the google geocoding api.
func (r *requestProcessor) ReverseGeocode(lat, lng float64) (GResponse, error) {
	//query url
	url := BASE_URL +
		"latlng=" + strconv.FormatFloat(lat, 'f', 8, 64) + "," + strconv.FormatFloat(lng, 'f', 8, 64) +
		"&language=" + r.lang +
		"&key=" + r.apiKey

	return r.processRequest(url)
}

//Geocode returns a GResponse object for the given address string.
//It contains all information offered by the google geocoding api.
func (r *requestProcessor) Geocode(address string) (GResponse, error) {
	//query url
	url := BASE_URL +
		"address=" + url.QueryEscape(address) +
		"&language=" + r.lang +
		"&key=" + r.apiKey

	return r.processRequest(url)
}
