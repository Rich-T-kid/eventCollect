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

func newQuery() *Queries {
	Connection := connect("./startup.g")
	return &Queries{db: Connection}
}

func (q *Queries) GetAllEvents() ([]string, error) {
	var events []string
	query := "SELECT name FROM events" // Replace "events" with your table name
	err := q.db.Select(&events, query)
	if err != nil {
		log.Printf("Failed to fetch events: %v", err)
		return nil, err
	}
	return events, nil
}
