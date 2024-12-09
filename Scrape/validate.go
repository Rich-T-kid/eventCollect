package scrape

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v9"
)

const linkCooldown = time.Hour * 24

type Validator interface {
	Validate(string) string
}
type Cache interface {
	// mutex is up to implementation to handle
	Get(key string) (value string, found bool)
	Put(key string, value string) error
	Exist(key string) bool
	Delete(key string) error
	IncreaseTTL(key string, extraTime time.Duration) error
	SetTTl(key string, ttl time.Duration) error
	Valid(key string) bool
	Flush()
	Save() error
}

func newCache() Cache {
	return newRedis()
}

// for now just run on local host
func newRedis() *redCache {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default to localhost for local development
	}
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // No password set
		DB:       0,  // Use default DB
	})

	fileName := fmt.Sprintf("%s_%s", "DB/cache", ".log")
	logFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to create/open log file: %v", err)
	}
	return &redCache{
		errorLog: logFile,
		client:   client,
	}
}

type redCache struct {
	errorLog *os.File
	mu       sync.Mutex
	client   *redis.Client
}

// TODO: add better error handling to the error log file. This is where well need a mutext to handle the writing operations

func (r *redCache) contextTimeout(seconds int) (context.Context, context.CancelFunc) {
	return context.WithDeadline(context.Background(), time.Now().Add(time.Second*time.Duration(seconds)))
}

func (r *redCache) Valid(key string) bool {
	return false
}
func (r *redCache) Flush() {
	ctx, _ := r.contextTimeout(2)
	_, err := r.client.FlushAll(ctx).Result()
	if err != nil {
		log.Fatalf("Error flushing all keys: %v", err)
	}
	fmt.Println("Just FLushed all keys")

}

func (r *redCache) Get(key string) (value string, found bool) {
	ctx, cancle := r.contextTimeout(3)
	defer cancle()
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", false
	}
	return val, true
}
func (r *redCache) Put(key string, value string) error {
	ctx, cancle := r.contextTimeout(3)
	defer cancle()
	err := r.client.Set(ctx, key, value, linkCooldown).Err()
	return err
}
func (r *redCache) Exist(key string) bool {
	ctx, cancel := r.contextTimeout(3)
	defer cancel()

	count, err := r.client.Exists(ctx, key).Result()
	if err != nil || count == 0 {
		return false
	}
	return true
}
func (r *redCache) Delete(key string) error {
	ctx, cancel := r.contextTimeout(3)
	defer cancel()

	_, err := r.client.Del(ctx, key).Result()
	return err
}
func (r *redCache) IncreaseTTL(key string, extra time.Duration) error {
	ctx, cancel := r.contextTimeout(3)
	defer cancel()

	// Check if the key exists
	if !r.Exist(key) {
		// Set the key with the given extra duration
		if _, err := r.client.Set(ctx, key, "", extra).Result(); err != nil {
			return fmt.Errorf("unable to set key with duration: %v", err)
		}
		return nil // Key is now set with the given duration
	}
	// Retrieve current TTL
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil || ttl <= 0 {
		return fmt.Errorf("unable to retrieve TTL or key has no TTL")
	}

	// Extend the TTL
	newTTL := ttl + extra
	if _, err := r.client.Expire(ctx, key, newTTL).Result(); err != nil {
		return err
	}
	return nil
}

func (r *redCache) SetTTl(key string, ttl time.Duration) error {
	ctx, cancel := r.contextTimeout(3)
	defer cancel()

	if !r.Exist(key) {
		return fmt.Errorf("key doesn't exist")
	}

	_, err := r.client.Expire(ctx, key, ttl).Result()
	return err
}
func (r *redCache) Save() error {
	r.errorLog.Close()
	return nil
}

type CustomCache struct {
	// You can add fields like map, TTL handling, or a mutex here.
	data map[string]string
	ttl  map[string]time.Time // Tracks key expiration times
}

// Ensure CustomCache implements the Cache interface
func (c *CustomCache) Get(key string) (value string, found bool) {
	return
}

func (c *CustomCache) Put(key string, value string) error {
	return nil
}

func (c *CustomCache) Valid(key string) bool { return false }
func (c *CustomCache) Exist(key string) bool {
	return false
}

func (c *CustomCache) Delete(key string) {
}

func (c *CustomCache) IncreaseTTL(key string, extraTime time.Duration) error {
	return nil
}

func (c *CustomCache) SetTTl(key string, ttl time.Duration) error {
	return nil
}

func (c *CustomCache) Save() error {
	return nil
}

func (c *CustomCache) Flush() {}

type CLeaner struct {
}

func (c *CLeaner) extractElementsWithClass(html string) []string {
	// Regular expression to match any element with the class "LrzXr"
	re := regexp.MustCompile(`<span class="[^"]*BNeawe tAd8D AP7Wnd[^"]*">(.*?)</span>`)
	// Find all matches
	matches := re.FindAllStringSubmatch(html, -1)

	// Collect the inner content of each match
	var results []string
	for _, match := range matches {
		if len(match) > 1 {
			results = append(results, match[1]) // Append the captured inner content
		}
	}

	return results
}

func (c *CLeaner) fetchAndExtractAddress(address string) (string, error) {
	const googleURL = "https://www.google.com/search?q="

	// Replace spaces with '+' for the search query
	newAddress := strings.ReplaceAll(address, " ", "+")
	url := googleURL + newAddress

	// Make HTTP GET request
	response, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer response.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	// Save response body to a file for debugging (optional)

	// Convert response body to string
	bodyString := string(bodyBytes)

	// Extract elements with the class "LrzXr"
	addresses := c.extractElementsWithClass(bodyString)
	if len(addresses) > 0 {
		return addresses[0], nil // Return the first matched address
	}

	return "", fmt.Errorf("no address found in the response")
}

func (c *CLeaner) ParseAddress(address string) (string, error) {
	return c.fetchAndExtractAddress(address)
}

type Geospatial struct {
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func BatchCoordinates(addresses []string) *[]Geospatial {
	apiKey := os.Getenv("GELOKY_KEY")

	// Convert addresses slice to a JSON array string
	addressesJSON, err := json.Marshal(addresses)
	if err != nil {
		log.Fatal("Error marshalling addresses:", err)
	}

	// Escape the JSON string for use in the query parameter
	escapedAddresses := url.QueryEscape(string(addressesJSON))
	apiURL := fmt.Sprintf("https://geloky.com/api/geo/geocode-batch?addresses=%s&key=%s&format=geloky", escapedAddresses, apiKey)

	// Make the HTTP GET request
	response, err := http.Get(apiURL)
	if err != nil {
		log.Fatal("Error making HTTP request:", err)
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Error reading response body:", err)
	}

	// Debug: Print raw response body

	// Parse the response into the Geospatial struct
	var geoResponse []Geospatial
	err = json.Unmarshal(body, &geoResponse)
	if err != nil {
		log.Fatal("Error unmarshalling response:", err)
	}

	return &geoResponse
}
func Cordniates(address string) (*Geospatial, error) {
	apiKey := os.Getenv("GELOKY_KEY")
	normalizedAddress := strings.ReplaceAll(address, ",", "")

	// Escape spaces and format the address for the API
	escapedAddress := url.QueryEscape(normalizedAddress)
	apiURL := fmt.Sprintf("https://geloky.com/api/geo/geocode?address=%s&key=%s&format=geloky", escapedAddress, apiKey)
	response, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("issue making request to %s response returned %v", apiURL, err)
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("did not recieve 200 status code from deocoding api")
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	var geoResponse []Geospatial
	err = json.Unmarshal(body, &geoResponse)
	if err != nil {
		return nil, fmt.Errorf("issue with geocoding service response %v", err)
	}
	if len(geoResponse) < 1 {
		return nil, fmt.Errorf("issue with geocoding service response %v", err)
	}

	return &geoResponse[0], nil
}

func GeoPoints(address string) (float64, float64) {
	geoObj, err := Cordniates(address)
	if err != nil {
		return -1, -1
	}
	return geoObj.Latitude, geoObj.Longitude
}
