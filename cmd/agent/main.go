package main

import (
	"fmt"

	"github.com/Anurag-Mishra22/observability-agent/internal/logging"
)

func main() {
	pipeline := logging.NewPipeline(1000)

	// Inputs
	pipeline.AddInput(logging.NewTailInput("/var/log/containers/test.log"))

	// Filters
	pipeline.AddFilter(&logging.NamespaceFilter{Excluded: []string{"kube-system"}})

	// Sinks
	pipeline.AddSink(&logging.StdoutSink{})

	fmt.Println("Observability Agent running...")
	pipeline.Run()

	select {} // block forever
}
