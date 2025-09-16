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
	// go func() {
	// 	logFile := "app.log"
	// 	if _, err := os.Stat(logFile); err == nil {
	// 		fmt.Printf("Tailing log file: %s\n", logFile)
	// 		if err := logging.TailFile(logFile); err != nil {
	// 			fmt.Println("Error tailing log file:", err)
	// 		}
	// 	} else {
	// 		fmt.Println("No log file found, skipping logging")
	// 	}
	// }()
	// Start fsnotify log tailing
	go func() {
		logFile := "app.log"
		if _, err := os.Stat(logFile); err == nil {
			fmt.Printf("Tailing log file using fsnotify: %s\n", logFile)
			if err := logging.TailFileFS(logFile); err != nil {
				fmt.Println("Error tailing log file:", err)
			}
		} else {
			fmt.Println("No log file found, skipping logging")
		}
	}()
	// Start container log tailing
	// go func() {
	// 	// Replace with the name or ID of your running container
	// 	containerName := "testapp"
	// 	fmt.Printf("Tailing container logs: %s\n", containerName)
	// 	if err := logging.TailContainerLogs(containerName); err != nil {
	// 		fmt.Println("Error tailing container logs:", err)
	// 	}
	// }()

	// Tail logs from all pods in specified namespace(s)
	go func() {
		namespace := os.Getenv("WATCH_NAMESPACE")
		if namespace == "" {
			fmt.Println("WATCH_NAMESPACE not set, tailing all namespaces")
		} else {
			fmt.Printf("Tailing pods in namespace: %s\n", namespace)
		}
		logging.TailAllPods(namespace)
	}()

	// Expose /metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	fmt.Println("Observability Agent running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
