package main

import (
	"fmt"
	"net/http"

	"github.com/Anurag-Mishra22/observability-agent/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Initialize system metrics
	metrics.Init()

	// Expose /metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	fmt.Println("Observability Agent running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
