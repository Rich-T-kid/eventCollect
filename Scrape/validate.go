package scrape

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v9"
)

type Validator interface {
	Validate(string) string
}
type Cache interface {
	// mutex is up to implementation to handle
	Get(key string) (value string, found bool)
	Put(key string, value string) error
	Exist(key string) bool
	Delete(key string)
	IncreaseTTL(key string, extraTime time.Duration) error
	SetTTl(key string, ttl time.Duration) error
	Save() error
}

func newCache() Cache {
	return newRedis()
}

// for now just run on local host
func newRedis() *redCache {
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

func (r *redCache) Get(key string) (value string, found bool) {
	return "", false
}
func (r *redCache) Put(key string, value string) error {
	return nil
}
func (r *redCache) Exist(key string) bool {
	return false
}
func (r *redCache) Delete(key string) {

}
func (r *redCache) IncreaseTTL(key string, extra time.Duration) error {
	return nil
}
func (r *redCache) SetTTl(key string, ttl time.Duration) error {
	return nil
}
func (r *redCache) Save() error {
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
	fmt.Println("Requesting:", url)

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
	f, _ := os.Create("test.html")
	defer f.Close()
	f.Write(bodyBytes)

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
