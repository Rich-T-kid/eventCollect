package scrape

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"lite/DB"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

// will write all errors to the console as well as to an error file
type Logger struct {
	ErrorLogger   *log.Logger // Logs errors
	InfoLogger    *log.Logger // Logs general information
	DebugLogger   *log.Logger // Logs debugging information
	ScrapeLogger  *log.Logger // Logs scraping-specific messages
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

	Generalfilename := fmt.Sprintf("%s/%s_general.log", folderName, prefix)
	generalLogFile, err := os.OpenFile(Generalfilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open general log file: %w", err)
	}

	Requestfilename := fmt.Sprintf("%s/%s_request.log", folderName, prefix)
	// Open the request log file
	requestLogFile, err := os.OpenFile(Requestfilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open request log file: %w", err)
	}
	Debugfilename := fmt.Sprintf("%s/%s_debug.log", folderName, prefix)
	debugFile, err := os.OpenFile(Debugfilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open general log file: %w", err)
	}
	multiWriter := io.MultiWriter(debugFile, os.Stdout)
	// Create loggers for each purpose
	errorLogger := log.New(generalLogFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger := log.New(generalLogFile, "INFO: ", log.Ldate|log.Ltime)
	debugLogger := log.New(multiWriter, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	scrapeLogger := log.New(generalLogFile, "SCRAPER: ", log.Ldate|log.Ltime)
	requestLogger := log.New(requestLogFile, "REQUEST: ", log.Ldate|log.Ltime)

	return &Logger{
		ErrorLogger:   errorLogger,
		InfoLogger:    infoLogger,
		DebugLogger:   debugLogger,
		ScrapeLogger:  scrapeLogger,
		RequestLogger: requestLogger,
	}, nil
}

// for constructuing a random user agent to not get blocked
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandomString() string {
	b := make([]byte, rand.Intn(10)+10)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
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
	log, err := NewLogger("")
	if err != nil {
		return nil, err
	}

	configColly(mainPage, log)
	configColly(sidePage, log)
	return NewScraper(mainPage, sidePage, log), nil
}
func Config() *scrape {
	c, err := initScrape()
	if err != nil {
		log.Fatalf("Couldnt start Scraping server %v", err)
	}

	return c
}
func configColly(c *colly.Collector, log *Logger) error {

	c.OnRequest(func(r *colly.Request) {
		// can add more stuff later but this is just the grounds work right now
		// set random user-agent to not get bot detected
		//r.Headers.Set("User-agent", RandomString())
		r.Headers.Set("Accept-Language", "en-US")
		str := fmt.Sprintf("Requesting %s [Method: %s, Headers: %v, Timestamp: %s]",
			r.URL,
			r.Method,
			r.Headers,
			time.Now().Format(time.RFC3339))
		log.RequestLogger.Println(str)
	})
	c.OnResponse(func(r *colly.Response) {
		str := fmt.Sprintf("Finished processing site %s [Status: %d, Length: %d bytes, Timestamp: %s]",
			r.Request.URL,
			r.StatusCode,
			len(r.Body),
			time.Now().Format(time.RFC3339))
		log.RequestLogger.Println(str)

	})
	c.OnError(func(r *colly.Response, err error) {
		str := fmt.Sprintf("Error: %v [URL: %s, Status: %d, Timestamp: %s]",
			err,
			r.Request.URL,
			r.StatusCode,
			time.Now().Format(time.RFC3339))
		log.ErrorLogger.Println(str)
		if len(r.Body) > 0 {
			log.ErrorLogger.Printf("Response Body: %s\n", string(r.Body))
		}
	})
	return nil
}
func (s *scrape) Start() {
	s.logger.DebugLogger.Println("Starting web scraper ........")
	var wg sync.WaitGroup
	ctx := context.Background()
	//defer cancel()

	linkChannel := make(chan string, 100)

	wg.Add(3)
	go func() {
		defer wg.Done()
		s.startSites()
	}()
	go func() {
		defer wg.Done()
		s.BeginScrape(ctx, linkChannel)
	}()

	go func() {
		defer wg.Done()
		s.ScrapeSidePages(ctx, linkChannel)
	}()
	wg.Wait()
	s.logger.DebugLogger.Println("Closing link channle")
	close(linkChannel)
	print("done waiting !!! \n")
}
func csvReader(filename string) [][]string {
	const foldername = "static_CSV"
	filename = fmt.Sprintf("%s/%s", foldername, filename)
	fmt.Printf("opening %s to retrive records\n", filename)
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
	// read in top 3000 in order as well as alll  NJ,Ny,PA,Bostan first Start with jusy new jersey for now and locations that u know
	// parse form the base url https://www.eventbrite.com/d/nj--piscataway/all-events/ and place on the channle
	// for now hardcode the state -> nj
	//const baseUrl = "https://www.eventbrite.com/d/nj--piscataway/all-events/"
	defer wg.Done()
	records := csvReader("us_cities.csv")
	for _, record := range records {
		cityName := strings.ToLower(record[0])
		state_name := strings.ToLower(record[3])
		ValidUrl := fmt.Sprintf("https://www.eventbrite.com/d/%s--%s/all-events/", state_name, cityName)
		links <- ValidUrl

	}
	fmt.Println("Done Proccessing US Links")
}
func (s *scrape) constructnjlinks(links chan string, wg *sync.WaitGroup) {
	// read in top 3000 in order as well as alll  NJ,Ny,PA,Bostan first Start with jusy new jersey for now and locations that u know
	// parse form the base url https://www.eventbrite.com/d/nj--piscataway/all-events/ and place on the channle
	// for now hardcode the state -> nj
	defer wg.Done()
	const state = "nj"
	records := csvReader("nj_cities.csv")
	for _, record := range records {
		cityName := strings.ToLower(record[0])
		ValidUrl := fmt.Sprintf("https://www.eventbrite.com/d/%s--%s/all-events/", state, cityName)
		links <- ValidUrl

	}
	fmt.Println("Done Proccessing New Jersey Links")
}
func (s *scrape) constructInternationalLinks(links chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	records := csvReader("non_us_cities.csv")
	// https://www.eventbrite.com/d/Bangladesh--Dhaka/events/
	//https://www.eventbrite.com/d/country--city/events/
	for _, record := range records {
		city := strings.ToLower(record[0])
		country := strings.ToLower(record[4])
		url := fmt.Sprintf("https://www.eventbrite.com/d/%s--%s/events/", country, city)
		links <- url
	}
	// this will read in from a csv file of the top 2000 citys and countrys
	// Checkk what regions areny valid and modify the csv file
}
func (s *scrape) startSites() {
	// Top Level URl will be constructed and then passed onto the main links queue -> for each link go through first 20 pages as is already being done
	// then  the main scrapper will get all the links on those pages and pass that onto a side scrapper queue
	// sider scrapers will proccess these sites for all their info and save this is  a data
	// Data Cleaning neds to be done , As well as using  a custom implementation of a bloom filter to check if a site has already been seen
	s.logger.DebugLogger.Println("Starting to generate Links")
	s.mainScraper.Visit("https://www.eventbrite.com/d/nj--piscataway/all-events/")
	var wg sync.WaitGroup
	mainsites := make(chan string, 100)
	wg.Add(3)
	go s.constructnjlinks(mainsites, &wg)
	go s.constructInternationalLinks(mainsites, &wg)
	go s.constructUSlinks(mainsites, &wg)
	go func() {
		wg.Wait()
		s.logger.DebugLogger.Println("All producers are done. Closing channel.")
		close(mainsites)
	}()
	cache := newCache()
	defer cache.Save() // always use save operation after calling caching. This handles all clean up
	for link := range mainsites {
		// for each link proccess all their pages. assumeing there are 11 pages. wont always be true but any failed request arent a major issue
		if cache.Exist(link) {
			s.logger.InfoLogger.Printf("Skipping cached link: %s\n", link)
			continue
		}
		cache.Put(link, "nil")
		cache.IncreaseTTL(link, time.Hour*24)
		s.logger.InfoLogger.Printf("Added new link to cache: %s with TTL of 24 hours\n", link)
		for i := 1; i < 5; i++ {
			// example link : https://www.eventbrite.com/d/nj--piscataway/all-events/
			// append ?page=i
			pageExtention := fmt.Sprintf("?page=%d", i)
			var completeUrl = link + pageExtention
			s.logger.DebugLogger.Printf("Main page Processing page: %s\n", completeUrl)
			err := s.mainScraper.Visit(completeUrl)
			if err != nil {
				s.logger.ErrorLogger.Println(err)
			}
		}
	}
}

// Grab the  main links
func (s *scrape) BeginScrape(ctx context.Context, link chan string) {
	fmt.Println("got inside begin scraper ")
	defer fmt.Println("exiting web scraper")
	s.mainScraper.OnHTML("section", func(e *colly.HTMLElement) {
		e.ForEach("ul.SearchResultPanelContentEventCardList-module__eventList___2wk-D", func(_ int, el *colly.HTMLElement) {
			//fmt.Println("Found <li> tag:", el.Text)
			el.ForEach("li", func(_ int, el *colly.HTMLElement) {
				event_link := el.ChildAttr("a", "href")
				link <- event_link

			})
		})

	})
	// Set up callbacks to handle scraping events
	// Visit the URL and start scraping

}
func (s *scrape) ScrapeSidePages(ctx context.Context, source chan string) {
	const noRefunds = "No Refunds"
	c := s.sideScraper
	db := DB.GetStorage()
	count := 0
	fmt.Println("got inside side scraper")
	defer fmt.Println("exiting side scraper")

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
		if len(refundPolicy) <= len(prefix) {
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

		// Print the collected data
		fmt.Println("Proccesed Event : - >")
		printEvent(event)
		if count >= 3 {
			return
		}
		if !event.ExactAddress {
			location = s.parseAddress(location)
		}
		id := db.AddEvent(event)
		count++
		lat, long := s.addressToCordnites(location)
		db.AddGeoPoint(title, id, DB.NewGeoPoint(lat, long, location))
	})

	for {
		select {
		case link, ok := <-source:
			if !ok {
				fmt.Println("got to not ok case")
				// Channel is closed and empty
				return
			}

			// Process the link
			fmt.Println("proccessing " + link)
			err := s.sideScraper.Visit(link)
			if err != nil {
				s.logger.ErrorLogger.Println(err)
			}
			//LrzXr
		case <-ctx.Done():
			// Context cancelled
			fmt.Println("complete with context cancled")

			return
		}
	}

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
