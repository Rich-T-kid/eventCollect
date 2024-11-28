package main

import (
	"fmt"
	"lite/DB"
	scrape "lite/Scrape"
	"lite/metrics"
	"log"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

type Collector struct {
	ID         uint
	Name       string
	Age        int
	SuperPower string
}

func init() {
	DB.InitDB()
	metrics.StartMetricsJob(time.Second*15, "metrics/metrics.json")
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	fmt.Println("env variables set up")
}
func main() {
	webCrawler := scrape.Config()
	fmt.Println("configed")
	webCrawler.Start()
	fmt.Println("Done here")
}

/*
https://www.eventbrite.com/d/nj--newark/all-events/?page=2
url -> https://www.eventbrite.com/d/nj--edison/all-events/
generic ulr (USA based) -? https://www.eventbrite.com/d/{state}--{city}/all-events/?page=int
generic ulr (non USA based) -? https://www.eventbrite.com/d/{country}--{city}/all-events/

*/
