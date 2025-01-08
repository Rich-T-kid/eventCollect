package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	db "lite/DB"
)

type Server struct {
	disk *db.Queries
}

var (
	defaultOffset = 0
	defaultLimit  = 200
)

func handleAndClean(offset, limit string) (int, int, error) {
	// Clean offset
	cleanOffset := defaultOffset
	if offset != "" {
		parsedOffset, err := strconv.Atoi(offset)
		if err != nil {
			return -1, -1, fmt.Errorf("invalid offset: %w", err)
		}
		cleanOffset = parsedOffset
	}

	// Clean limit
	cleanLimit := defaultLimit
	if limit != "" {
		parsedLimit, err := strconv.Atoi(limit)
		if err != nil {
			return -1, -1, fmt.Errorf("invalid limit: %w", err)
		}
		cleanLimit = parsedLimit
	}

	return cleanOffset, cleanLimit, nil
}
func (s *Server) events(w http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	offset := queryParams.Get("offset")
	limit := queryParams.Get("limit")

	// Handle and clean parameters
	cleanOffset, cleanLimit, err := handleAndClean(offset, limit)
	if err != nil {
		http.Error(w, "Invalid offset or limit passed in request: "+err.Error(), http.StatusBadRequest)
		return
	}
	events, err := s.disk.GetAllEvents(uint(cleanOffset), uint(cleanLimit))
	if err != nil {
		http.Error(w, "Database Operation to fetch events has failed: "+err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	response := eventResponse{
		Total:   len(events),
		Payload: events,
	}
	json.NewEncoder(w).Encode(response)

	// rest of this is just  a simple databse call
}
func (s *Server) eventLocation(w http.ResponseWriter, req *http.Request) {

	queryParams := req.URL.Query()
	offset := queryParams.Get("offset")
	limit := queryParams.Get("limit")
	//from := queryParams.Get("from")

	// Handle and clean parameters
	cleanOffset, cleanLimit, err := handleAndClean(offset, limit)
	if err != nil {
		http.Error(w, "Invalid offset or limit passed in request: "+err.Error(), http.StatusBadRequest)
		return
	}
	locations, err := s.disk.GetAllEventslocations(uint(cleanOffset), uint(cleanLimit))
	if err != nil {
		http.Error(w, "Database Operation to fetch Event Locations has failed: "+err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	response := eventResponse{
		Total:   len(locations),
		Payload: locations,
	}
	json.NewEncoder(w).Encode(response)
}

func (s *Server) Start() error {
	http.HandleFunc("/life", s.life)
	http.HandleFunc("/events", s.events)
	http.HandleFunc("/eventLocation", s.eventLocation)

	// Run the server in a goroutine
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Block forever to keep the program running
	select {}
}

func NewServer() *Server {
	return &Server{
		disk: db.NewQuery(),
	}
}

func (s *Server) life(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Hello world"))
	w.WriteHeader(200)
}

type eventResponse struct {
	Total   int
	Payload interface{}
}
