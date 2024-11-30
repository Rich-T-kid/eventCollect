package main

import (
	"lite/DB"
	scrape "lite/Scrape"
	"lite/metrics"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	boldRed := color.New(color.FgRed, color.Bold)
	DB.InitDB()
	metrics.StartMetricsJob(time.Second*15, "metrics/metrics.json")
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	x := os.Getenv("REDIS_ADDR")
	x2 := os.Getenv("GEOCODE_API_KEY")
	x3 := os.Getenv("GELOKY_KEY")
	boldRed.Printf("env Variables reddisaddr : %s , geoCodeAPi : %s , gelokKey : %s", x, x2, x3)
	boldRed.Println("Complete setting up enviroment")
}
func main() {
	underlineGreen := color.New(color.FgGreen, color.Underline)
	webCrawler := scrape.Config()
	underlineGreen.Println("Scraper is configed")
	webCrawler.Start()
	underlineGreen.Println("Done here")
}

/*
https://www.eventbrite.com/d/nj--newark/all-events/?page=2
url -> https://www.eventbrite.com/d/nj--edison/all-events/
generic ulr (USA based) -? https://www.eventbrite.com/d/{state}--{city}/all-events/?page=int
generic ulr (non USA based) -? https://www.eventbrite.com/d/{country}--{city}/all-events/

*/
