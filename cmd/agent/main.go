package main

import (
	"fmt"
	"log"

	"github.com/Anurag-Mishra22/observability-agent/internal/logging"
)

func main() {
	logChannel := make(logging.LogEventChannel, 1000)

	// 1. Setup sinks
	stdoutSink := &logging.StdoutSink{}
	fileSink, err := logging.NewFileSink("/var/log/agent-logs.json")
	if err != nil {
		log.Fatal("Error creating file sink:", err)
	}
	sinks := []logging.Sink{stdoutSink, fileSink}

	// 2. Setup filters
	filters := []logging.Filter{}

	// Example: drop kube-system logs
	// Drop kube-system logs
	filters = append(filters, &logging.NamespaceFilter{Excluded: []string{"kube-system"}})

	// Add Kubernetes metadata enrichment
	k8sFilter, err := logging.NewKubernetesMetadataFilter()
	if err != nil {
		log.Fatal("Failed to init Kubernetes metadata filter:", err)
	}
	filters = append(filters, k8sFilter)

	// 3. Pipeline processor
	go func() {
		for event := range logChannel {
			pass := true
			for _, f := range filters {
				if e, ok := f.Apply(event); ok {
					event = *e
				} else {
					pass = false
					break
				}
			}
			if pass {
				for _, s := range sinks {
					s.Write(event)
				}
			}
		}
	}()

	// 4. Input: Watch container logs
	go func() {
		err := logging.WatchContainerLogs("/var/log/containers", logChannel)
		if err != nil {
			fmt.Println("Error watching container logs:", err)
		}
	}()

	fmt.Println("Observability Agent running...")
	select {}
}
