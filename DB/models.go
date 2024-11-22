package DB

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type GeoPoint struct {
	ID        int     `db:"id" json:"id"` // Primary key
	Latitude  float64 `db:"latitude" json:"latitude"`
	Longitude float64 `db:"longitude" json:"longitude"`
	Address   string  `db:"address" json:"address"` // street name, etc.
	EventID   int     `db:"event_id" json:"event_id"`
}

type Event struct {
	ID          int       `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	StartDate   time.Time `db:"start_date" json:"start_date"`
	EndDate     time.Time `db:"end_date" json:"end_date"`
	Price       float32   `db:"price" json:"price"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	LastUpdated time.Time `db:"last_updated" json:"last_updated"`
}

type EventInfo struct {
	ID              int       `db:"id" json:"id"`
	EventID         int       `db:"event_id" json:"event_id"` // associates with Event ID
	Bio             string    `db:"bio" json:"bio"`
	MaxCapacity     int       `db:"max_capacity" json:"max_capacity"`
	CurrentCapacity int       `db:"current_capacity" json:"current_capacity"`
	HostName        string    `db:"host_name" json:"host_name"`
	VipEligible     bool      `db:"vip_eligible" json:"vip_eligible"`
	Tags            string    `db:"free_all" json:"free_all"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	LastUpdated     time.Time `db:"last_updated" json:"last_updated"`
}

// for raw SQl querys
type Queries struct {
	db *sqlx.DB
	// add a logge in later
}
