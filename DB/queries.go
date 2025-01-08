package DB

import (
	"log"

	"github.com/jmoiron/sqlx"
)

/*
This is where well load the local data to a cloud Database
*/
func connect(databasePath string) *sqlx.DB {
	// Open the SQLite database file
	db, err := sqlx.Open("sqlite3", databasePath)
	if err != nil {
		log.Fatalf("Failed to connect to SQLite database: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping SQLite database: %v", err)
	}

	log.Println("Connected to SQLite database successfully!")
	return db
}

func NewQuery() *Queries {
	Connection := connect("./startup.g")
	return &Queries{db: Connection}
}

func (q *Queries) GetAllEvents(offset, limit uint) ([]Event, error) {
	var events []Event
	query := "SELECT * FROM events limit ? offset ? " // Replace "events" with your table name
	err := q.db.Select(&events, query, limit, offset)
	if err != nil {
		log.Printf("Failed to fetch events: %v", err)
		return nil, err
	}
	return events, nil
}

func (q *Queries) GetAllEventslocations(offset, limit uint) ([]GeoPoint, error) {
	var GeoPoints []GeoPoint
	query := "SELECT * FROM GeoPoint limit ? offset ? " // Replace "events" with your table name
	err := q.db.Select(&GeoPoints, query, limit, offset)
	if err != nil {
		log.Printf("Failed to fetch events: %v", err)
		return nil, err
	}
	return GeoPoints, nil
}

func (q *Queries) EventbyLocation(lat, long float64, offset, limit uint) ([]GeoPoint, error) {
	return nil, nil
}
