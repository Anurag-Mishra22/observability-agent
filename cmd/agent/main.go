package main

import (
	"fmt"

	"github.com/Anurag-Mishra22/observability-agent/internal/logging"
)

func main() {
	logChannel := make(logging.LogEventChannel, 1000)

	// Choose sink (stdout for now)
	sinker := &logging.StdoutSink{}

	// Define filters
	filters := []logging.Filter{
		&logging.NamespaceFilter{Excluded: []string{"kube-system", "kube-public"}},
		&logging.KeywordFilter{Keyword: "error"},
	}
	// Pipeline
	go func() {
		for event := range logChannel {
			keep := true
			var ev *logging.LogEvent = &event

			for _, f := range filters {
				var ok bool
				ev, ok = f.Apply(*ev)
				if !ok {
					keep = false
					break
				}
			}

			if keep && ev != nil {
				_ = sinker.Write(*ev)
			}
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
