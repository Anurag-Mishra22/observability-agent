package main

import (
	"fmt"

	"github.com/Anurag-Mishra22/observability-agent/internal/logging"
)

func main() {
	logChannel := make(logging.LogEventChannel, 1000)

	// Choose sink (stdout for now)
	sinker := &logging.StdoutSink{}

	// Start pipeline processor
	go func() {
		for event := range logChannel {
			// Example filter: skip kube-system logs
			if event.Namespace == "kube-system" {
				continue
			}
			_ = sinker.Write(event)
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
