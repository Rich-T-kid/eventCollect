package DB

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type DataStore interface {
	Insert(event) error
	Update(event) error
}
type event interface {
	isEvent()
}
type GeoPoint struct {
	ID        int     `db:"id" json:"id"` // Primary key
	Latitude  float64 `db:"latitude" json:"latitude"`
	Longitude float64 `db:"longitude" json:"longitude"`
	Address   string  `db:"address" json:"address"` // street name, etc.
	EventID   int     `db:"event_id" json:"event_id"`
}

type Event struct {
	ID             int    `db:"id" json:"id"`
	ImageUrl       string `json:"image_url" db:"image_url"`
	Host           string `json:"host" db:"host"`
	Title          string `json:"title" db:"title"`
	Date           string `json:"date" db:"date"`
	Location       string `json:"location" db:"location"`
	Description    string `json:"description" db:"description"`
	Tags           string `json:"tags" db:"tags"`
	ExtraInfo      string `json:"extra_info" db:"extra_info"`
	Bio            string `json:"bio" db:"bio"`
	ExactAddress   bool   `json:"exact_address" db:"exact_address"`
	AcceptsRefunds bool   `json:"accepts_refunds" db:"accepts_refunds"`
}

type EventInfo struct {
	ID              int       `db:"id" json:"id"`
	EventID         int       `db:"event_id" json:"event_id"` // associates with Event ID
	MaxCapacity     int       `db:"max_capacity" json:"max_capacity"`
	CurrentCapacity int       `db:"current_capacity" json:"current_capacity"`
	HostName        string    `db:"host_name" json:"host_name"`
	VipEligible     bool      `db:"vip_eligible" json:"vip_eligible"`
	Tags            string    `db:"free_all" json:"free_all"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	LastUpdated     time.Time `db:"last_updated" json:"last_updated"`
}

func (e *EventInfo) isEvent() {}
func (e *Event) isEvent()     {}
func (e *GeoPoint) isEvent()  {}

// for raw SQl querys
type Queries struct {
	db *sqlx.DB
	// add a logge in later
}
