package logging

import (
	"encoding/json"
	"time"
)

// LogEvent represents a structured log
type LogEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	Line      string                 `json:"line"`
	Pod       string                 `json:"pod,omitempty"`
	Container string                 `json:"container,omitempty"`
	Namespace string                 `json:"namespace,omitempty"`
	Source    string                 `json:"source,omitempty"` // file, container, pod
	Extra     map[string]interface{} `json:"extra,omitempty"`  // parsed JSON content
}

// ParseLog takes a raw log line and converts it into structured JSON
func ParseLog(line string, source, pod, container, namespace string) LogEvent {
	event := LogEvent{
		Timestamp: time.Now(),
		Line:      line,
		Pod:       pod,
		Container: container,
		Namespace: namespace,
		Source:    source,
	}

	// Try parsing JSON
	var extra map[string]interface{}
	if json.Unmarshal([]byte(line), &extra) == nil {
		event.Extra = extra
	}

	return event
}
