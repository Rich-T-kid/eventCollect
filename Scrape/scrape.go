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
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	linkChannel := make(chan string, 10)
	//defer close(linkChannel)

	wg.Add(2)
	go func() {
		defer wg.Done()
		s.BeginScrape(ctx, linkChannel)
	}()

	go func() {
		defer wg.Done()
		s.ScrapeSidePages(ctx, linkChannel)
	}()
	go s.startSites()
	wg.Wait()
	print("done waiting !!! \n")
}
func csvReader(filename string) [][]string {
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
	return records

}
func (s *scrape) constructUSlinks(links chan string) {
	// read in top 3000 in order as well as alll  NJ,Ny,PA,Bostan first Start with jusy new jersey for now and locations that u know
	// parse form the base url https://www.eventbrite.com/d/nj--piscataway/all-events/ and place on the channle
	// for now hardcode the state -> nj
	//const baseUrl = "https://www.eventbrite.com/d/nj--piscataway/all-events/"

	records := csvReader("uscities.csv")
	for _, record := range records {
		cityName := strings.ToLower(record[0])
		state_name := strings.ToLower(record[3])
		ValidUrl := fmt.Sprintf("https://www.eventbrite.com/d/%s--%s/all-events/", state_name, cityName)
		links <- ValidUrl

	}
	fmt.Println("Done Proccessing US Links")
}
func (s *scrape) constructnjlinks(links chan string) {
	// read in top 3000 in order as well as alll  NJ,Ny,PA,Bostan first Start with jusy new jersey for now and locations that u know
	// parse form the base url https://www.eventbrite.com/d/nj--piscataway/all-events/ and place on the channle
	// for now hardcode the state -> nj
	const state = "nj"
	records := csvReader("nj-municipalities.csv")
	for _, record := range records {
		cityName := strings.ToLower(record[0])
		ValidUrl := fmt.Sprintf("https://www.eventbrite.com/d/%s--%s/all-events/", state, cityName)
		links <- ValidUrl

	}
	fmt.Println("Done Proccessing New Jersey Links")
}
func (s *scrape) constructInternationalLinks(links chan string) {
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
	mainsites := make(chan string, 100)
	go s.constructnjlinks(mainsites)
	go s.constructInternationalLinks(mainsites)
	go s.constructUSlinks(mainsites)
	for link := range mainsites {
		// for each link proccess all their pages. assumeing there are 11 pages. wont always be true but any failed request arent a major issue
		for i := 1; i < 11; i++ {
			// example link : https://www.eventbrite.com/d/nj--piscataway/all-events/
			// append ?page=i
			pageExtention := fmt.Sprintf("?page=%d", i)
			var completeUrl = link + pageExtention
			print("visiting ", completeUrl, "\n")
			err := s.mainScraper.Visit(completeUrl)
			if err != nil {
				s.logger.ErrorLogger.Println(err)
			}
		}
	}
	for i := 1; i < 2; i++ {
		url := fmt.Sprintf("https://www.eventbrite.com/d/nj--piscataway/all-events/?page=%d", i)
		print("visiting ", url, "\n")
		err := s.mainScraper.Visit(url)
		if err != nil {
			s.logger.ErrorLogger.Println(err)
		}
	}
}
func (s *scrape) BeginScrape(ctx context.Context, link chan string) {
	s.mainScraper.OnHTML("section.SearchPageContent-module__searchPanel___3TunM ul.SearchResultPanelContentEventCardList-module__eventList___2wk-D", func(e *colly.HTMLElement) {
		// Find the `a` tag inside the nested `section` with class `event-card-actions`
		e.ForEach("li", func(i int, li *colly.HTMLElement) {
			//eventTitle := li.ChildText("h3") // Adjust selector as needed
			eventLink := li.ChildAttr("a", "href")
			link <- eventLink
		})
	})
	// Set up callbacks to handle scraping events
	// Visit the URL and start scraping

}
func (s *scrape) ScrapeSidePages(ctx context.Context, source chan string) {
	c := s.sideScraper
	db := DB.GetStorage()
	count := 0
	fmt.Println("got inside side scraper")
	c.OnHTML("span.date-info__full-datetime", func(h *colly.HTMLElement) {
		fmt.Println(h.Text)
	})

	c.OnHTML("body", func(h *colly.HTMLElement) {
		// Date
		//date := h.ChildText("span.date-info__full-datetime")

		// Location
		location := h.ChildText("p.location-info__address-text")
		// bio
		//bio := h.ChildText("p.summary")
		title := h.ChildText("h1.event-title.css-0")
		// Description (all paragraphs)
		var descriptionParts []string
		h.ForEach("div.has-user-generated-content.event-description__content p", func(_ int, el *colly.HTMLElement) {
			descriptionParts = append(descriptionParts, el.Text)
		})
		// Tags
		var tags []string
		h.ForEach("li.tags-item", func(_ int, el *colly.HTMLElement) {
			tags = append(tags, el.Text)
		})
		if count >= 2 {
			return
		}
		location = s.parseAddress(location)
		id := db.AddEvent(title, time.Now(), time.Now().Add(4*time.Hour), 43, strings.Join(descriptionParts, "\n"), 200, 32, "host1", false, tags)
		count++
		lat, long := s.addressToCordnites(location)
		db.AddGeoPoint(title, id, DB.NewGeoPoint(lat, long, location))
		// now store the event in the databse
		// Print or process the extracted information
		///fmt.Println("Title:", title)
		//fmt.Println("Date:", date)
		//fmt.Println("Location:", location)
		//fmt.Println("Description:", strings.Join(descriptionParts, "\n"))
		//fmt.Println("Tags:", tags)
		//fmt.Println("bio:", bio)
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

func paseDate(date string) time.Time {
	return time.Now()
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
