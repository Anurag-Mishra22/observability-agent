package logging

import "fmt"

// Input defines a source of logs (like TailInput, KubeInput)
type Input interface {
	Start(out LogEventChannel) error
}

// Pipeline ties together inputs, filters, and sinks
type Pipeline struct {
	Inputs  []Input
	Filters []Filter
	Sinks   []Sink
	Events  LogEventChannel
}

// NewPipeline creates a new pipeline with a buffered channel
func NewPipeline(bufferSize int) *Pipeline {
	return &Pipeline{
		Inputs:  []Input{},
		Filters: []Filter{},
		Sinks:   []Sink{},
		Events:  make(LogEventChannel, bufferSize),
	}
}

// AddInput registers an input plugin
func (p *Pipeline) AddInput(input Input) {
	p.Inputs = append(p.Inputs, input)
}

// AddFilter registers a filter plugin
func (p *Pipeline) AddFilter(filter Filter) {
	p.Filters = append(p.Filters, filter)
}

// AddSink registers a sink plugin
func (p *Pipeline) AddSink(sink Sink) {
	p.Sinks = append(p.Sinks, sink)
}

// Run starts the pipeline
func (p *Pipeline) Run() {
	// Start inputs
	for _, input := range p.Inputs {
		go func(in Input) {
			if err := in.Start(p.Events); err != nil {
				fmt.Println("Input error:", err)
			}
		}(input)
	}

	// Process events
	go func() {
		for event := range p.Events {
			// Apply filters
			keep := true
			current := &event
			for _, f := range p.Filters {
				var ok bool
				current, ok = f.Apply(*current)
				if !ok {
					keep = false
					break
				}
			}
			if !keep || current == nil {
				continue
			}

			// Send to sinks
			for _, s := range p.Sinks {
				if err := s.Write(*current); err != nil {
					fmt.Println("Sink error:", err)
				}
			}
		}
	}()
}
