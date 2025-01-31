package scrape

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

/*
Main concern right now is that the location data is there for a street name but we need the Lat and lOng of this
This will be handled using batch jobs. Every couple of seconds or minutes we iterate through the database and we translate the street Name into Lat Long
Very simple
Keep everything general and only accpet interfaces as we will be changing out our datastores as we continue to tesr
*/
var (
	once     sync.Once
	instance *Geocoder
	apikey   = "6740b9d3ea16b460848865roa6225f6" //
	baseUrl  = "https://geocode.maps.co/search"  //
)

type GeoAPIResponse []struct {
	PlaceID     int      `json:"place_id"`
	Licence     string   `json:"licence"`
	OsmType     string   `json:"osm_type"`
	OsmID       int64    `json:"osm_id"`
	Boundingbox []string `json:"boundingbox"`
	Lat         string   `json:"lat"`
	Lon         string   `json:"lon"`
	DisplayName string   `json:"display_name"`
	Class       string   `json:"class"`
	Type        string   `json:"type"`
	Importance  float64  `json:"importance"`
}
type GeoAPI struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Geocoder struct {
	apiKey  string
	baseUrl string
}

func newGeoCoder(apiKey string, baseUrl string) *Geocoder {
	return &Geocoder{
		apiKey:  apiKey,
		baseUrl: baseUrl,
	}
}
func (g *Geocoder) streetToCordinates(address string) (float64, float64, error) {
	address = strings.TrimSpace(address)

	// Step 2: Remove punctuation (e.g., commas, extra symbols)
	address = g.removePunctuation(address)

	// Step 3: Replace spaces with "+" for URL encoding
	address = strings.ReplaceAll(address, " ", "+")
	// Step 4: Generate the request URL
	requestURL := fmt.Sprintf("%s?q=%s&api_key=%s", g.baseUrl, address, g.apiKey)
	response, err := http.Get(requestURL)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}
	if response.StatusCode != 200 {
		var geoAPIError GeoAPI
		err = json.Unmarshal(body, &geoAPIError)
		if err != nil {
			fmt.Println(err)
		}
		response := fmt.Sprintf("Api Response %s , api Code %d", geoAPIError.Message, geoAPIError.Code)
		fmt.Println(response)
		fmt.Printf("Error: %s\n", geoAPIError.Message)
		return -1, -1, errors.New(geoAPIError.Message)
	}
	if len(body) < 10 {
		return -1, -1, fmt.Errorf("api service doesnt have geocoding for this street")
	}
	var geoAPIResponse GeoAPIResponse
	err = json.Unmarshal(body, &geoAPIResponse)
	if err != nil {
		return -1, -1, fmt.Errorf("error unmarshalling JSON: %v", err)
	}
	Latitude, err := strconv.ParseFloat(geoAPIResponse[0].Lat, 64)
	if err != nil {
		fmt.Println("Error converting string to float64:", err)
		return -1, -1, fmt.Errorf("error converting string latidude to float64: %v", err)
	}
	longitude, err := strconv.ParseFloat(geoAPIResponse[0].Lon, 64)
	if err != nil {
		fmt.Println("Error converting string to float64:", err)
		return -1, -1, fmt.Errorf("error converting string longitude to float64: %v", err)
	}
	return Latitude, longitude, nil

}

func (g *Geocoder) removePunctuation(input string) string {
	// Regex to match and remove punctuation
	re := regexp.MustCompile(`[^\w\s]`) // Matches anything that's not a word or space
	return re.ReplaceAllString(input, "")
}

func geoCoderInstance() *Geocoder {
	once.Do(func() {
		instance = newGeoCoder(apikey, baseUrl)
	})
	return instance
}

var (
	gelokyApiKey string = "xZhll25Ktga6TpfHAd1uZMjZqrF06oWq"
)

type geoResponse struct {
	Address   string      `json:"address"`
	Latitude  StringOrInt `json:"latitude"`
	Longitude StringOrInt `json:"longitude"`
}

type EventLocation struct {
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Custom type to handle Latitude/Longitude that can be string or int
type StringOrInt string

func (s *StringOrInt) UnmarshalJSON(data []byte) error {
	// Check if the value is a string
	if data[0] == '"' {
		*s = StringOrInt(data[1 : len(data)-1]) // Remove quotes
		return nil
	}

	// Otherwise, it's a number; convert to string

	// Otherwise, it's a number; convert to string
	str := string(data) // Convert the raw bytes to a string
	*s = StringOrInt(str)
	return nil
}

type addressCleaner struct {
	logger *log.Logger
}

func newAddressCleaner(l *log.Logger) *addressCleaner {
	return &addressCleaner{
		logger: l,
	}
}

// The caller of this function must handle checking for empty values
func (a *addressCleaner) ReverseGeoCode(streetName string) EventLocation {
	if streetName == "" {
		return EventLocation{
			Address:   "",
			Latitude:  -1,
			Longitude: -1,
		}
	}
	apiResponse := streetToCord(streetName, a.logger)
	return geoReponseParse(apiResponse, a.logger)
}

func streetToCord(streetName string, logger *log.Logger) geoResponse {
	defualtResponse := geoResponse{Address: "", Latitude: "1", Longitude: "1"}
	escapedStreet := strings.ReplaceAll(streetName, " ", "%20")
	url := fmt.Sprintf("https://geloky.com/api/geo/geocode?address=%s&key=%s&format=geloky", escapedStreet, gelokyApiKey)
	fmt.Println("GeoCode api endpoint : ", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("HTTP Get error %e \n", err)
		return defualtResponse
	}
	defer resp.Body.Close()
	fmt.Println(resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Println("io read error ", err)
		return defualtResponse
	}
	var location []geoResponse
	err = json.Unmarshal(body, &location)
	if err != nil {
		fmt.Printf("error unmarshaling response %e , current geoReponse %v \n", err, location)
		return defualtResponse
	}

	if len(location) < 1 {
		logger.Printf("Invalid Response url: %s  response json %v\n", url, location)
	}
	instance := location[0]
	return instance
}

func geoReponseParse(g geoResponse, logger *log.Logger) EventLocation {
	// In case where geoResponse passed to this isn't valid, use default values
	lat, err := strconv.ParseFloat(string(g.Latitude), 64)
	if err != nil {
		lat = -1
		logger.Println("Error parsing lat value", err)
	}
	long, err := strconv.ParseFloat(string(g.Longitude), 64)
	if err != nil {
		long = -1
		logger.Println("Error parsing long value", err)
	}
	return EventLocation{Address: g.Address, Latitude: lat, Longitude: long}
}
