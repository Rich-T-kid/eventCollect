package scrape

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

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
