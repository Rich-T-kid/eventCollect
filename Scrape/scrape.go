package scrape

import (
	"context"
	"encoding/csv"
	"fmt"
	"lite/DB"
	"lite/pkg"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

var colorOutput = pkg.NewTextStyler()

// will write all errors to the console as well as to an error file
type Logger struct {
	ErrorLogger   *log.Logger // Logs errors
	InfoLogger    *log.Logger // Logs general information
	DebugLogger   *log.Logger // Logs debugging information
	RequestLogger *log.Logger // Logs HTTP requests (separate file)
}
type scrape struct {
	mainScraper *colly.Collector
	sideScraper *colly.Collector
	logger      *Logger
	mu          sync.Mutex
}

func NewLogger(prefix string) (*Logger, error) {
	// Open the general log file
	const folderName = "ScapeLogs"
	err := os.MkdirAll(folderName, 0755) // Creates the folder if it doesn't exist
	if err != nil {
		log.Fatalf("error creating folder: %v\n", err)
		return nil, err
	}
	debug_file := pkg.CreateLogFile(folderName + "/" + prefix + "general")
	info_file := pkg.CreateLogFile(folderName + "/" + prefix + "info")
	request_file := pkg.CreateLogFile(folderName + "/" + prefix + "request")
	error_file := pkg.CreateLogFile(folderName + "/" + prefix + "error")
	//multiWriter := io.MultiWriter(info_file, os.Stdout)
	// Create loggers for each purpose
	errorLogger := log.New(error_file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger := log.New(info_file, "INFO: ", log.Ldate|log.Ltime)
	debugLogger := log.New(debug_file, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	requestLogger := log.New(request_file, "REQUEST: ", log.Ldate|log.Ltime)

	return &Logger{
		ErrorLogger:   errorLogger,
		InfoLogger:    infoLogger,
		DebugLogger:   debugLogger,
		RequestLogger: requestLogger,
	}, nil
}

func NewScraper(c *colly.Collector, s *colly.Collector, l *Logger) *scrape {
	return &scrape{
		mainScraper: c,
		sideScraper: s,
		logger:      l,
	}
}
func initScrape() (*scrape, error) {
	mainPage := colly.NewCollector()
	sidePage := colly.NewCollector()
	log, err := NewLogger("__")
	cache := newCache()
	if err != nil {
		return nil, err
	}

	configColly(mainPage, log, "Main Page Scraper", cache)
	configColly(sidePage, log, "Side Page Scraper", cache)
	return NewScraper(mainPage, sidePage, log), nil
}
func Config() *scrape {
	c, err := initScrape()
	if err != nil {
		log.Fatalf("Couldnt start Scraping server %v", err)
	}
	colorOutput.Red("Colly output files set up complete")
	colorOutput.UnderlineGreen("Colly Set up complete")

	return c
}
func configColly(c *colly.Collector, log *Logger, name string, cache Cache) error {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	// Wrap the HTTP client to respect context
	str := fmt.Sprintf("colly configured with max Idle Connections: %d , idle Connection Timeout: %d seconds, TLS Handshake %d seconds", 10, 30, 30)
	colorOutput.Yellow(str)
	c.WithTransport(httpClient.Transport)

	c.OnRequest(func(r *colly.Request) {
		// can add more stuff later but this is just the grounds work right now
		// set random user-agent to not get bot detected
		//r.Headers.Set("User-agent", RandomString())
		r.Headers.Set("Accept-Language", "en-US")
		str := fmt.Sprintf("%s: Requesting %s [Method: %s, Headers: %v, Timestamp: %s]",
			name,
			r.URL,
			r.Method,
			r.Headers,
			time.Now().Format(time.RFC3339))
		log.RequestLogger.Println(str)
	})
	c.OnResponse(func(r *colly.Response) {
		str := fmt.Sprintf("%s: Finished processing site %s [Status: %d, Length: %d bytes, Timestamp: %s]", name,
			r.Request.URL,
			r.StatusCode,
			len(r.Body),
			time.Now().Format(time.RFC3339))
		log.RequestLogger.Println(str)
		if r.StatusCode == 404 { // if URL doesnt Exist never visit it again
			cache.IncreaseTTL(r.Request.URL.String(), time.Hour*24*30*12) // 1 year
		}

	})
	c.OnError(func(r *colly.Response, err error) {
		str := fmt.Sprintf("%s: Error: %v [URL: %s, Status: %d, Timestamp: %s]", name,
			err,
			r.Request.URL,
			r.StatusCode,
			time.Now().Format(time.RFC3339))
		if r.StatusCode == 404 || err == colly.ErrAlreadyVisited { // if URL doesnt Exist never visit it again
			cache.IncreaseTTL(r.Request.URL.String(), time.Hour*24*30*12) // 1 year
			log.InfoLogger.Printf("%s does not exist, blacklisting URL\n", r.Request.URL.String())
			return
		}
		log.ErrorLogger.Println(str)
		if len(r.Body) > 0 {
			log.ErrorLogger.Printf("first 100 bytes of Response Body: %s\n", string(r.Body)[:100])
		}
	})
	return nil
}

func csvReader(filename string) [][]string {
	const foldername = "static_CSV"
	filename = fmt.Sprintf("%s/%s", foldername, filename)
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}
	// first row of these record as just defining the columns
	return records[1:]

}
func (s *scrape) constructUSlinks(links chan string, wg *sync.WaitGroup) {
	defer colorOutput.Red("Done Constructing US Links")
	defer wg.Done()
	records := csvReader("us_cities.csv")
	for _, record := range records {
		cityName := strings.ToLower(record[0])
		state_name := strings.ToLower(record[3])
		ValidUrl := fmt.Sprintf("https://www.eventbrite.com/d/%s--%s/all-events/", state_name, cityName)
		links <- ValidUrl

	}
}
func (s *scrape) constructnjlinks(links chan string, wg *sync.WaitGroup) {
	defer colorOutput.Red("Done Constructing NJ links")
	defer wg.Done()
	const state = "nj"
	records := csvReader("nj.csv")
	for _, record := range records {
		cityName := strings.ToLower(record[0])
		cityName = strings.Replace(cityName, " ", "-", -1)
		ValidUrl := fmt.Sprintf("https://www.eventbrite.com/d/%s--%s/all-events/", state, cityName)
		links <- ValidUrl

	}
}
func (s *scrape) constructInternationalLinks(links chan string, wg *sync.WaitGroup) {
	defer colorOutput.Red("Done Proccessing International Links")
	defer wg.Done()
	records := csvReader("non_us_cities.csv")
	for _, record := range records {
		city := strings.ToLower(record[0])
		country := strings.ToLower(record[4])
		url := fmt.Sprintf("https://www.eventbrite.com/d/%s--%s/events/", country, city)
		links <- url
	}
}

func (s *scrape) Start() error {
	colorOutput.Green("Starting web scrapper .....")
	// Context handling -> for later
	mainCtx, cancle := context.WithTimeout(context.Background(), time.Second*45)
	defer cancle()
	sideCtx, cancle := context.WithTimeout(context.Background(), time.Second*45)
	defer cancle()

	var consumerWG sync.WaitGroup
	cache := newCache()
	cache.Save()
	producerChannel := make(chan string, 33000) // Buffered channel for producers
	SideProducer := make(chan string, 33000)    // Buffered channel for producers
	done := make(chan bool)

	// set call back functions for colly
	s.BeginScrape(SideProducer)
	s.BeginSideScrape(mainCtx, SideProducer)
	// Start Workers that will construct the URL's for main page as well as the side page workers that will proccess the links on the main page
	go s.startSites(producerChannel, done)
	go s.ScrapeSidePages(sideCtx, SideProducer)
	//
	workers := 5
	consumerWG.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer consumerWG.Done()
			for link := range producerChannel {
				select {
				case <-mainCtx.Done():
					msg := "(Main Page Scraper): Context canelled, stopping wokrer"
					colorOutput.BoldRed(msg)
					s.logger.ErrorLogger.Println(msg)
					return
				default:
					s.processLink(mainCtx, link, cache)
				}
			}
		}()
	}
	consumerWG.Wait()
	<-done
	close(SideProducer)
	colorOutput.BoldRed("Done with Init of Web scraper")
	return nil
}

func (s *scrape) startSites(mainsites chan string, done chan bool) {
	colorOutput.Red("Starting to generate links")
	var wg sync.WaitGroup
	wg.Add(3)
	go s.constructnjlinks(mainsites, &wg)
	go s.constructUSlinks(mainsites, &wg)
	go s.constructInternationalLinks(mainsites, &wg)
	colorOutput.UnderlineGreen("Waiting for go routines to finish")
	wg.Wait()
	close(mainsites)
	done <- true

}

func (s *scrape) processLink(ctx context.Context, link string, cache Cache) {
	// This function processes a single link concurrently
	if cache.Exist(link) {
		s.logger.DebugLogger.Printf("Skipping cached link: %s\n", link)
		return
	}
	cache.Put(link, "nil")
	cache.IncreaseTTL(link, time.Hour*24)
	s.logger.InfoLogger.Printf("Added new link to cache: %s with TTL of 24 hours\n", link)
	select {
	case <-ctx.Done():
		s.logger.ErrorLogger.Printf("(Main Scraper): Context canceled, skipping link: %s\n", link)
		return
	default:
		for i := 1; i < 2; i++ {
			pageExtention := fmt.Sprintf("?page=%d", i)
			completeUrl := link + pageExtention
			s.mainScraper.Visit(completeUrl)
			// Error handling is handled in the colly conifg
		}
	}
}

// Grab the  main links
func (s *scrape) BeginScrape(links chan string) {
	colorOutput.Green("Creating callback Function on main page")
	s.mainScraper.OnHTML("section", func(e *colly.HTMLElement) {
		e.ForEach("ul.SearchResultPanelContentEventCardList-module__eventList___2wk-D", func(_ int, el *colly.HTMLElement) {
			//fmt.Println("Found <li> tag:", el.Text)
			el.ForEach("li", func(_ int, el *colly.HTMLElement) {
				event_link := el.ChildAttr("a", "href")
				links <- event_link

			})
		})

	})
}

func (s *scrape) ScrapeSidePages(ctx context.Context, source chan string) {
	var wg sync.WaitGroup
	workerPool := 10
	str := fmt.Sprintf("Starting Side Scraper with %d go-routines", workerPool)
	colorOutput.UnderlineGreen(str)

	wg.Add(workerPool)
	for i := 0; i < workerPool; i++ {
		go func() {
			defer wg.Done()
			for link := range source {
				select {
				case <-ctx.Done():
					msg := "(Side Page Scraper): Context cancelled, stopping side page worker"
					colorOutput.BoldRed(msg)
					s.logger.ErrorLogger.Println(msg)
					return
				default:
					// Process the link
					err := s.sideScraper.Visit(link)
					if err != nil && err != colly.ErrAlreadyVisited {
						s.logger.ErrorLogger.Printf("%s failed with following error %v", link, err)
					}
				}
			}
		}()
	}
	wg.Wait()
	colorOutput.UnderlineGreen("Done proccessing side Pages.")

}

func (s *scrape) BeginSideScrape(ctx context.Context, source chan string) {
	colorOutput.UnderlineGreen("Creating Call back function on side pages")
	const noRefunds = "No Refunds"
	c := s.sideScraper
	db := DB.GetStorage()

	c.OnHTML("body", func(h *colly.HTMLElement) {
		// Extract data using CSS selectors
		var addressFound bool
		var validRefunds bool
		host := h.ChildText("strong.organizer-listing-info-variant-b__name-link")
		date := h.ChildText("span.date-info__full-datetime")
		location := h.ChildText("p.location-info__address-text")
		exactAddress := h.ChildText("div.location-info__address")
		bio := h.ChildText("p.summary")
		title := h.ChildText("h1.event-title.css-0")
		imageURL := h.ChildAttr("img", "src")
		const prefix = "Refund Policy"
		refundPolicy := h.ChildText("section[aria-labelledby='refund-policy-heading'] div")
		// checks if html has certain structre by checking len of parsed string. if long enough removes prefix and check if the refund policy is listed as no refunds. If it isnt then refund flag is Set to true as this means you must contact host for explicit refunds rules.

		if len(refundPolicy) >= len(prefix) {

			policy := refundPolicy[len(prefix):]
			if policy != noRefunds {
				validRefunds = true
			}

		}
		// Description (all paragraphs within a specific div)
		var descriptionParts []string
		h.ForEach("div.has-user-generated-content.event-description__content p", func(_ int, el *colly.HTMLElement) {
			descriptionParts = append(descriptionParts, el.Text)
		})

		// Tags (list items with tag class)
		var tags []string
		h.ForEach("li.tags-item", func(_ int, el *colly.HTMLElement) {
			tags = append(tags, el.Text)
		})

		// Extract additional information from li tags in ul.css-1i6cdnn
		var extraInfo [][]string
		h.ForEach("ul.css-1i6cdnn", func(i int, el *colly.HTMLElement) {
			// Collect the text from each <li> in the <ul> and store it in listItems
			el.ForEach("li", func(_ int, li *colly.HTMLElement) {
				text := []string{li.Text}
				extraInfo = append(extraInfo, text)
			})

		})

		// Print the extracted extra info (this is for debugging)

		// if exact address is present, no need to do geoFinding
		if exactAddress != "" {
			location = exactAddress
			addressFound = true
		}

		// Create an Event struct to store the data
		event := DB.Event{
			ImageUrl:       imageURL,
			Host:           host,
			Title:          title,
			Date:           date,
			Location:       location,
			Description:    strings.Join(descriptionParts, "\n"),
			Tags:           strings.Join(tags, ", "),
			Bio:            bio,
			ExactAddress:   addressFound,
			ExtraInfo:      flattenAndJoin(extraInfo), // Store the extracted extra info
			AcceptsRefunds: validRefunds,
		}

		if !event.ExactAddress {
			location = "Delete this value later"
		}
		id := db.AddEvent(event)
		lat, long := -1.1, -1.1 //s.addressToCordnites(location)
		db.AddGeoPoint(title, id, DB.NewGeoPoint(lat, long, location))
	})

}
func (s *scrape) parseAddress(address string) string {
	var c CLeaner
	address, err := c.ParseAddress(address)
	if err != nil {
		s.logger.ErrorLogger.Println("parsing addrress ", err)
		return ""
	}
	return address
	// do a googlesearch  and then do span.LrzXr colly scapre to parse out the address adn return that
}

func (s *scrape) addressToCordnites(address string) (float64, float64) {
	var geoCoder = geoCoderInstance()
	lat, long, err := geoCoder.streetToCordinates(address)
	if err != nil {
		s.logger.ErrorLogger.Println(err)
		return -1, -1 // if for what ever reason theres an error just log it and have -1 be the placeholders
	}
	return lat, long
}

func printEvent(event DB.Event) {
	fmt.Println("Host:", event.Host)
	fmt.Println("Title:", event.Title)
	fmt.Println("Date:", event.Date)
	fmt.Println("Location:", event.Location)
	fmt.Println("Description:", event.Description)
	fmt.Println("Tags:", event.Tags)
	fmt.Println("Bio:", event.Bio)
	fmt.Println("ExactAddress Found:", event.ExactAddress)
	fmt.Println("Extra Info:", event.ExtraInfo)
	fmt.Println("image Url:", event.ImageUrl)
	fmt.Println("accepts refunds:", event.AcceptsRefunds)

}

func flattenAndJoin(input [][]string) string {
	// Create a slice to hold all the individual strings
	var flatSlice []string

	// Iterate through the 2D array
	for _, outer := range input {
		// For each inner slice, append its elements to flatSlice
		flatSlice = append(flatSlice, outer...)
	}

	// Join the strings into a single comma-separated string
	return strings.Join(flatSlice, ", ")
}
