package DB

import (
	"fmt"
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

var (
	// Singleton instance of Storage
	storageInstance *Storage
	// Mutex to synchronize creation
	once           sync.Once
	gormConnection *gorm.DB
)

// Storage struct encapsulates the database connection and log file

// GetStorage returns the singleton instance of Storage
func GetStorage() *Storage {
	// Ensure the instance is created only once
	once.Do(func() {
		// Initialize the Storage instance
		storageInstance = &Storage{
			Database: createDatabaseConnection(),
			logFile:  createLogFile("DB/Database"),
		}
	})
	return storageInstance
}

// createDatabaseConnection initializes the gorm.DB connection
func createDatabaseConnection() *gorm.DB {
	// Replace this with your actual database configuration
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}
	updateModels(db)
	return db
}

// createLogFile initializes the log file
func createLogFile(prefix string) *os.File {
	fileName := fmt.Sprintf("%s_%s", prefix, ".log")
	logFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to create/open log file: %v", err)
	}
	return logFile
}

func InitDB() {
	db, err := gorm.Open(sqlite.Open("startup.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error opening database: %v\n", err)
	}

	// Optionally, auto-migrate your schema
	err = updateModels(db)
	if err != nil {
		log.Fatal("Error migrating Database tables -> Repo.go in database file")
	}
	gormConnection = db
	fmt.Println("Database Connection is ready to go")
}

func updateModels(db *gorm.DB) error {
	// very easy to just add them in here
	return db.AutoMigrate(&Event{}, &EventInfo{}, &GeoPoint{})
}
func newEvent(title string, start, end time.Time, price float32) *Event {
	return &Event{
		Name:        title,
		StartDate:   start,
		EndDate:     end,
		Price:       price,
		CreatedAt:   time.Now(),
		LastUpdated: time.Now(),
	}
}
func newEventInfo(EventId int, bio string, maxCapacity, currentCap int, hostname string, eligibal bool, tags string) *EventInfo {
	return &EventInfo{
		EventID:         EventId,
		Bio:             bio,
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
	var constMessage = fmt.Sprintf("Created Event %s at %v", event.Name, time.Now())
	s.logFile.Write([]byte(constMessage))
}

func (s *Storage) createEventInfo(title string, eventInfo *EventInfo) {
	s.Database.Create(eventInfo)
	var constMessage = fmt.Sprintf("Created EventInfo %s at %v", title, time.Now())
	s.logFile.Write([]byte(constMessage))
}

func (s *Storage) createEventGeo(title string, Geo *GeoPoint) {
	s.Database.Create(Geo)
	var constMessage = fmt.Sprintf("Created EventGeo Point %s: %v at %v \n", title, Geo, time.Now())
	s.logFile.Write([]byte(constMessage))
}
func (s *Storage) AddEvent(name string, start, end time.Time, price float32, bio string, maxCapacity, currentCap int, hostname string, eligibal bool, tags []string) int {
	newEvent := newEvent(name, start, end, price)
	s.createEvent(newEvent)
	s.createEventInfo(newEvent.Name, newEventInfo(newEvent.ID, bio, maxCapacity, currentCap, hostname, eligibal, ""))
	return newEvent.ID
}
func (s *Storage) AddGeoPoint(title string, eventId int, Geo *GeoPoint) {
	Geo.EventID = eventId
	s.createEventGeo(title, Geo)
}
