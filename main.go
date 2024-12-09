package main

import (
	"fmt"
	scrape "lite/Scrape"

	_ "github.com/mattn/go-sqlite3"
)

func init() {
	fmt.Println("Configing project")
}
func main() {
	webCrawler := scrape.Config()
	fmt.Println("configed")
	webCrawler.Start()
}

/*
https://www.eventbrite.com/d/nj--newark/all-events/?page=2
url -> https://www.eventbrite.com/d/nj--edison/all-events/
generic ulr (USA based) -? https://www.eventbrite.com/d/{state}--{city}/all-events/?page=int
generic ulr (non USA based) -? https://www.eventbrite.com/d/{country}--{city}/all-events/

*/
