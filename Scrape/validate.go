package scrape

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Validator interface {
	Validate(string) string
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
