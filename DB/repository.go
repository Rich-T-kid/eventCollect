package DB

import (
	"fmt"
	"lite/pkg"
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Storage struct {
	Database *gorm.DB
	logFile  *os.File
}

func (s Storage) Start() error {
	GetStorage()
	return nil
}

var (
	colorOutput *pkg.TextStyler
	// Singleton instance of Storage
	storageInstance *Storage
	// Mutex to synchronize creation
	once sync.Once
)

// Storage struct encapsulates the database connection and log file

// GetStorage returns the singleton instance of Storage
func GetStorage() *Storage {
	// Ensure the instance is created only once
	once.Do(func() {
		// Initialize the Storage instance
		colorOutput = pkg.NewTextStyler()
		colorOutput.Red("Configed Color Ouput")
		storageInstance = &Storage{
			Database: createDatabaseConnection(),
			logFile:  pkg.CreateLogFile("DB/_Database"),
		}
	})
	return storageInstance
}

// createDatabaseConnection initializes the gorm.DB connection
func createDatabaseConnection() *gorm.DB {
	// Replace this with your actual database configuration
	db, err := gorm.Open(sqlite.Open("DataStore.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}
	err = updateModels(db)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// createLogFile initializes the log file

func updateModels(db *gorm.DB) error {
	// very easy to just add them in here
	return db.AutoMigrate(&Event{}, &EventInfo{}, &GeoPoint{})
}
func newEventInfo(EventId int, bio string, maxCapacity, currentCap int, hostname string, eligibal bool, tags string) *EventInfo {
	return &EventInfo{
		EventID:         EventId,
		MaxCapacity:     maxCapacity,
		CurrentCapacity: currentCap,
		HostName:        hostname,
		VipEligible:     eligibal,
		Tags:            tags,
		CreatedAt:       time.Now(),
		LastUpdated:     time.Now(),
	}
}
func NewGeoPoint(Lat, Long float64, streetName string) *GeoPoint {
	return &GeoPoint{
		Latitude:  Lat,
		Longitude: Long,
		Address:   streetName,
	}
}

// Handle insert statments for the data first and formost we can query the data very easily later
func (s *Storage) createEvent(event *Event) {
	s.Database.Create(event)
	var constMessage = fmt.Sprintf("Created Event %s at %v\n", event.Title, time.Now())
	s.logFile.Write([]byte(constMessage))
}

func (s *Storage) createEventInfo(title string, eventInfo *EventInfo) {
	s.Database.Create(eventInfo)
	var constMessage = fmt.Sprintf("Created EventInfo %s at %v\n", title, time.Now())
	s.logFile.Write([]byte(constMessage))
}

func (s *Storage) createEventGeo(title string, Geo *GeoPoint) {
	s.Database.Create(Geo)
	var constMessage = fmt.Sprintf("Created EventGeo Point %s: %v at %v \n", title, Geo, time.Now())
	s.logFile.Write([]byte(constMessage))
}
func (s *Storage) AddEvent(event Event) int {
	s.createEvent(&event)
	//s.createEventInfo(newEvent.Name, newEventInfo(newEvent.ID, bio, maxCapacity, currentCap, hostname, eligibal, ""))
	return event.ID
}
func (s *Storage) AddGeoPoint(title string, eventId int, Geo *GeoPoint) {
	Geo.EventID = eventId
	s.createEventGeo(title, Geo)
}
