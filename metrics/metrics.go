package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"runtime"

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

func CollectMetrics(fileName string, interval time.Duration) {
	fmt.Println("Metrics is up and running")
	for {
		// Collect metrics
		cpuPercent, _ := cpu.Percent(0, false)
		memStats, _ := mem.VirtualMemory()
		diskStats, _ := disk.Usage("/")
		var appMem runtime.MemStats
		runtime.ReadMemStats(&appMem)

		metrics := Metrics{
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

		// Open the file in append mode
		file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return
		}

		// Write the JSON object to the file as part of an array
		jsonData, err := json.MarshalIndent(metrics, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling JSON: %v\n", err)
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
		file.Write(jsonData)
		file.WriteString("\n]") // Close the JSON array

		file.Close()

		// Wait for the next interval
		time.Sleep(interval)
	}
}
