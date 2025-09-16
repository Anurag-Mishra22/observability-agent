package main

import (
	"fmt"

	"github.com/Anurag-Mishra22/observability-agent/internal/logging"
)

func main() {
	pipeline := logging.NewPipeline(1000)

	// -------------------
	// Inputs
	// -------------------
	// Option 1: Tail a single file
	// pipeline.AddInput(logging.NewTailInput("/var/log/containers/test.log"))

	// Option 2: DirectoryInput (Fluent Bit style)
	pipeline.AddInput(logging.NewDirectoryInput("/var/log/containers"))

	// -------------------
	// Filters
	// -------------------
	pipeline.AddFilter(&logging.NamespaceFilter{Excluded: []string{"kube-system"}})

	// -------------------
	// Sinks
	// -------------------
	pipeline.AddSink(&logging.StdoutSink{})

	fmt.Println("Observability Agent running...")
	pipeline.Run()

	select {} // block forever
}
