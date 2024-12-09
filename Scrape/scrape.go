package scrape

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"lite/DB"
	"lite/pkg"
	"log"
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
	multiWriter := io.MultiWriter(info_file, os.Stdout)
	// Create loggers for each purpose
	errorLogger := log.New(error_file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger := log.New(multiWriter, "INFO: ", log.Ldate|log.Ltime)
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
	defer wg.Done()
	records := csvReader("us_cities.csv")
	for _, record := range records {
		cityName := strings.ToLower(record[0])
		state_name := strings.ToLower(record[3])
		ValidUrl := fmt.Sprintf("https://www.eventbrite.com/d/%s--%s/all-events/", state_name, cityName)
		fmt.Sprintf(ValidUrl)
		//links <- ValidUrl

	}
	fmt.Println("Done Proccessing US Links")
}
func (s *scrape) constructnjlinks(links chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	const state = "nj"
	records := csvReader("nj_cities.csv")
	for _, record := range records {
		cityName := strings.ToLower(record[0])
		cityName = strings.Replace(cityName, " ", "-", -1)
		ValidUrl := fmt.Sprintf("https://www.eventbrite.com/d/%s--%s/all-events/", state, cityName)
		links <- ValidUrl

	}
	fmt.Println("Done Proccessing New Jersey Links")
}
func (s *scrape) constructInternationalLinks(links chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	records := csvReader("non_us_cities.csv")
	for _, record := range records {
		city := strings.ToLower(record[0])
		country := strings.ToLower(record[4])
		url := fmt.Sprintf("https://www.eventbrite.com/d/%s--%s/events/", country, city)
		fmt.Sprint(url)
		//links <- url
	}
	fmt.Println("Done with Internatinal links")
}

func (s *scrape) Start() error {
	s.logger.DebugLogger.Println("Starting web scraper ........")
	ctx, cancle := context.WithTimeout(context.Background(), time.Second*45)
	defer cancle()
	//	var producerWG sync.WaitGroup
	var consumerWG sync.WaitGroup
	cache := newCache()
	cache.Flush()
	cache.Save()
	producerChannel := make(chan string, 33000) // Buffered channel for producers
	SideProducer := make(chan string, 33000)    // Buffered channel for producers
	done := make(chan bool)
	s.BeginScrape(SideProducer)
	go s.startSites(producerChannel, done)
	go s.ScrapeSidePages(ctx, SideProducer)
	fmt.Println("consuming now")
	consumerWG.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			defer consumerWG.Done()
			for link := range producerChannel {
				s.processLink(link, cache)
			}
		}()
	}
	consumerWG.Wait()
	fmt.Println("waiting to get messaeg to close channel")
	<-done
	fmt.Println("just go the message to close the channel")
	close(SideProducer)
	s.logger.DebugLogger.Println("All scraping tasks completed")
	return nil
}

func (s *scrape) startSites(mainsites chan string, done chan bool) {
	s.logger.DebugLogger.Println("Starting to generate Links")

	var wg sync.WaitGroup
	wg.Add(3)
	go s.constructnjlinks(mainsites, &wg)
	go s.constructUSlinks(mainsites, &wg)
	go s.constructInternationalLinks(mainsites, &wg)

	wg.Wait()
	close(mainsites)
	//time.Sleep(time.Second * 15)
	fmt.Println("sending close the other channel message")
	done <- true
	s.logger.DebugLogger.Println("Finished processing all links")
}

func (s *scrape) processLink(link string, cache Cache) {
	// This function processes a single link concurrently
	if cache.Exist(link) {
		s.logger.InfoLogger.Printf("Skipping cached link: %s\n", link)
		return
	}
	cache.Put(link, "nil")
	cache.IncreaseTTL(link, time.Hour*24)
	//s.logger.InfoLogger.Printf("Added new link to cache: %s with TTL of 24 hours\n", link)

	workerPool := 50
	// Loop through pages (assuming 2 pages for now)
	var wg sync.WaitGroup
	wg.Add(workerPool)
	for x := 0; x < workerPool; x++ {
		go func() {
			defer wg.Done()
			for i := 1; i < 4; i++ {
				pageExtention := fmt.Sprintf("?page=%d", i)
				completeUrl := link + pageExtention
				s.logger.DebugLogger.Printf("Main page Processing page: %s\n", completeUrl)

				//fmt.Println("visiting", completeUrl)
				err := s.mainScraper.Visit(completeUrl)
				if err != nil {
					s.logger.ErrorLogger.Println(err)
				}
			}
		}()
	}
}

// Grab the  main links
func (s *scrape) BeginScrape(links chan string) {

	s.mainScraper.OnHTML("section", func(e *colly.HTMLElement) {
		e.ForEach("ul.SearchResultPanelContentEventCardList-module__eventList___2wk-D", func(_ int, el *colly.HTMLElement) {
			//fmt.Println("Found <li> tag:", el.Text)
			el.ForEach("li", func(_ int, el *colly.HTMLElement) {
				event_link := el.ChildAttr("a", "href")
				links <- event_link

			})
		})

	})
	fmt.Println("set up call back all done!")
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
		fmt.Println("Set up the webscraper behavoir")
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
			location = s.parseAddress(location)
		}
		id := db.AddEvent(event)
		count++
		lat, long := s.addressToCordnites(location)
		db.AddGeoPoint(title, id, DB.NewGeoPoint(lat, long, location))
	})
	var counter int
	var mu sync.Mutex
	var wg sync.WaitGroup
	workerPool := 100

	for i := 0; i < workerPool; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for link := range source {
				// Process the link
				mu.Lock()
				if counter%10 == 0 {
					fmt.Println("processing link -> ", link)
				}
				counter++
				mu.Unlock()

				err := s.sideScraper.Visit(link)
				if err != nil {
					s.logger.ErrorLogger.Println(err)
				}
			}
			fmt.Println("Worker exiting")
		}()
	}
	defer fmt.Println("Done proccessing links")

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
