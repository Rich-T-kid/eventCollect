package main

import (
	"lite/DB"
	scrape "lite/Scrape"
	"lite/metrics"
	"lite/pkg"

	_ "github.com/mattn/go-sqlite3"
)

/*
Add notifications on certian conditoons (start , stop, crash) -> textbelt API
Figure out what to do with the location data we are getting
Figure out what to do with the Invalid Date Format we  are recieveing
*/
func main() {
	colorOP := pkg.NewTextStyler()
	db := DB.GetStorage()
	webCrawler := scrape.Config()
	met := &metrics.Metrics{}
	pkg.SetUp(met, db, webCrawler)
	colorOP.BoldRed("Complete with webscraper")

}
