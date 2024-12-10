package main

import (
	"lite/DB"
	scrape "lite/Scrape"
	"lite/metrics"
	"lite/pkg"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	colorOP := pkg.NewTextStyler()
	db := DB.GetStorage()
	webCrawler := scrape.Config()
	met := &metrics.Metrics{}
	pkg.SetUp(met, db, webCrawler)
	colorOP.BoldRed("Complete with webscraper")

}
