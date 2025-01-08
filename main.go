package main

import (
	"log"

	"github.com/joho/godotenv"

	"lite/DB"
	scrape "lite/Scrape"
	"lite/metrics"
	"lite/pkg"
	api "lite/server"

	_ "github.com/mattn/go-sqlite3"
)

/*
Add notifications on certian conditoons (start , stop, crash) -> textbelt API
Figure out what to do with the location data we are getting
Figure out what to do with the Invalid Date Format we  are recieveing
*/
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
func main() {
	colorOP := pkg.NewTextStyler()
	db := DB.GetStorage()
	webCrawler := scrape.Config()
	met := &metrics.Metrics{}
	s := api.NewServer()
	pkg.SetUp(met, db, webCrawler, s)
	colorOP.BoldRed("Complete with webscraper")

}
