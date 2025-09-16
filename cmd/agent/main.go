package main

import (
	"fmt"

	"github.com/Anurag-Mishra22/observability-agent/internal/logging"
)

func main() {
	logChannel := make(logging.LogEventChannel, 1000)

	// Start pipeline processor
	go func() {
		for event := range logChannel {
			// Filter logs
			if event.Namespace == "kube-system" {
				continue
			}
			logging.Sink(event)
		}
	}()

	// Watch /var/log/containers for all log files dynamically
	go func() {
		err := logging.WatchContainerLogs("/var/log/containers", logChannel)
		if err != nil {
			fmt.Println("Error watching container logs:", err)
		}
	}()

	fmt.Println("Observability Agent running...")
	select {} // block forever
}
