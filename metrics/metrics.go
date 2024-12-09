package metrics

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"runtime"

	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type Metrics struct {
	Time            string  `json:"time"`
	CPUUsage        float64 `json:"cpu_usage_percent"`
	TotalMemory     uint64  `json:"total_memory_mb"`
	UsedMemory      uint64  `json:"used_memory_mb"`
	MemoryUsage     float64 `json:"memory_usage_percent"`
	DiskTotal       uint64  `json:"disk_total_gb"`
	DiskUsed        uint64  `json:"disk_used_gb"`
	DiskUsage       float64 `json:"disk_usage_percent"`
	AllocatedMemory uint64  `json:"allocated_memory_kb"`
	NumGC           uint32  `json:"num_gc"`
}

var currentMetrics Metrics

func CollectMetrics(fileName string, interval time.Duration) {
	fmt.Println("Metrics is up and running")
	go func() {
		for {
			// Collect metrics
			cpuPercent, _ := cpu.Percent(0, false)
			memStats, _ := mem.VirtualMemory()
			diskStats, _ := disk.Usage("/")
			var appMem runtime.MemStats
			runtime.ReadMemStats(&appMem)

			currentMetrics = Metrics{
				Time:            time.Now().Format(time.RFC3339),
				CPUUsage:        cpuPercent[0],
				TotalMemory:     memStats.Total / 1024 / 1024,
				UsedMemory:      memStats.Used / 1024 / 1024,
				MemoryUsage:     memStats.UsedPercent,
				DiskTotal:       diskStats.Total / 1024 / 1024 / 1024,
				DiskUsed:        diskStats.Used / 1024 / 1024 / 1024,
				DiskUsage:       diskStats.UsedPercent,
				AllocatedMemory: appMem.Alloc / 1024,
				NumGC:           appMem.NumGC,
			}
			saveMetricsToFile(fileName)

			// Wait for the next interval
			time.Sleep(interval)
		}
	}()
}
func saveMetricsToFile(fileName string) {
	// Open the file in append mode
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Marshal the metrics to JSON
	metricsJSON, err := json.MarshalIndent(currentMetrics, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	// Append the JSON object with a comma if the file already contains data
	fileStat, _ := file.Stat()
	if fileStat.Size() == 0 {
		// File is empty; start a new JSON array
		file.WriteString("[\n")
	} else {
		// Add a comma and newline before appending
		file.Seek(-2, os.SEEK_END) // Move back 2 bytes to replace "]\n"
		file.WriteString(",\n")
	}

	// Write the current JSON object
	file.Write(metricsJSON)
	file.WriteString("\n]") // Close the JSON array

	log.Printf("Metrics saved to file: %s", currentMetrics.Time)
}

// API handler to serve the latest metrics as JSON
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Set the response header to indicate JSON
	w.Header().Set("Content-Type", "application/json")

	// Return the latest metrics as JSON
	err := json.NewEncoder(w).Encode(currentMetrics)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	var message = fmt.Sprintf("Sever is Live %v", time.Now())
	w.Write([]byte(message))
}
func StartMetricsJob(interval time.Duration, fileName string) {
	// Start collecting metrics every 5 seconds in a separate goroutine
	go CollectMetrics(fileName, interval)

	go func() {
		http.HandleFunc("/metrics", metricsHandler)
		http.HandleFunc("/Life", healthCheck)

		// Start the HTTP server (blocking operation)
		port := ":8080"
		fmt.Printf("Starting server on http://localhost%s...\n", port)
		log.Fatal(http.ListenAndServe(port, nil))
	}()

}

func (m *Metrics) Start() error {
	StartMetricsJob(time.Second*15, "metrics/metrics.json")
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	return nil // this cant fail but must fufil the interface
}
