package main

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL    = "http://srv.msk01.gigacorp.local/_stats"
	pollInterval = 1 * time.Second
	maxErrors    = 3
)

type ServerStats struct {
	LoadAverage      uint64
	TotalMemory      uint64
	UsedMemory       uint64
	TotalDisk        uint64
	UsedDisk         uint64
	NetworkBandwidth uint64
	NetworkUsage     uint64
}

func main() {
	errorCount := 0

	for {
		stats, err := fetchStats()
		if err != nil {
			errorCount++
			fmt.Printf("Error fetching stats: %v\n", err)

			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic")
				errorCount = 0
			}

			time.Sleep(pollInterval)
			continue
		}

		errorCount = 0

		checkLoadAverage(stats)
		checkMemoryUsage(stats)
		checkDiskSpace(stats)
		checkNetworkUsage(stats)

		time.Sleep(pollInterval)
	}
}

func fetchStats() (*ServerStats, error) {
	resp, err := http.Get(serverURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty response")
	}

	line := scanner.Text()
	values := strings.Split(line, ",")

	if len(values) != 7 {
		return nil, fmt.Errorf("invalid data format: expected 7 values, got %d", len(values))
	}

	stats := &ServerStats{}

	stats.LoadAverage, err = strconv.ParseUint(values[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid load average: %v", err)
	}

	stats.TotalMemory, err = strconv.ParseUint(values[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid total memory: %v", err)
	}

	stats.UsedMemory, err = strconv.ParseUint(values[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid used memory: %v", err)
	}

	stats.TotalDisk, err = strconv.ParseUint(values[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid total disk: %v", err)
	}

	stats.UsedDisk, err = strconv.ParseUint(values[4], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid used disk: %v", err)
	}

	stats.NetworkBandwidth, err = strconv.ParseUint(values[5], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid network bandwidth: %v", err)
	}

	stats.NetworkUsage, err = strconv.ParseUint(values[6], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid network usage: %v", err)
	}

	return stats, nil
}

func checkLoadAverage(stats *ServerStats) {
	if stats.LoadAverage > 30 {
		fmt.Printf("Load Average is too high: %d\n", stats.LoadAverage)
	}
}

func checkMemoryUsage(stats *ServerStats) {
	if stats.TotalMemory == 0 {
		return
	}

	memoryUsagePercent := int64((float64(stats.UsedMemory) / float64(stats.TotalMemory)) * 100)
	if memoryUsagePercent > 80 {
		fmt.Printf("Memory usage too high: %d%%\n", memoryUsagePercent)
	}
}

func checkDiskSpace(stats *ServerStats) {
	if stats.TotalDisk == 0 {
		return
	}

	diskUsagePercent := int64((float64(stats.UsedDisk) / float64(stats.TotalDisk)) * 100)
	if diskUsagePercent > 90 {
		freeSpaceMB := int64(float64(stats.TotalDisk-stats.UsedDisk) / (1024 * 1024))
		fmt.Printf("Free disk space is too low: %d Mb left\n", freeSpaceMB)
	}
}

func checkNetworkUsage(stats *ServerStats) {
	if stats.NetworkBandwidth == 0 {
		return
	}

	networkUsagePercent := int64((float64(stats.NetworkUsage) / float64(stats.NetworkBandwidth)) * 100)
	if networkUsagePercent > 90 {
		availableBandwidthMbps := (float64(stats.NetworkBandwidth-stats.NetworkUsage) / 1_000_000)
		fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", int64(availableBandwidthMbps))
	}
}
