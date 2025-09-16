package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Anurag-Mishra22/observability-agent/internal/logging"
	"github.com/Anurag-Mishra22/observability-agent/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Initialize system metrics
	metrics.Init()

	// Start log tailing (async)
	go func() {
		logFile := "app.log"
		if _, err := os.Stat(logFile); err == nil {
			fmt.Printf("Tailing log file: %s\n", logFile)
			if err := logging.TailFile(logFile); err != nil {
				fmt.Println("Error tailing log file:", err)
			}
		} else {
			fmt.Println("No log file found, skipping logging")
		}
	}()

	// Expose /metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	fmt.Println("Observability Agent running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
