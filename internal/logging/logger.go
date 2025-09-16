package logging

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hpcloud/tail"
)

type LogEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Line      string    `json:"line"`
	// Optional fields for different log sources
	Source    string `json:"source,omitempty"`    // file or docker ID
	Pod       string `json:"pod,omitempty"`       // for Kubernetes
	Container string `json:"container,omitempty"` // docker or pod container
	Namespace string `json:"namespace,omitempty"` // pod namespace

}

// TailFile tails a given log file and prints logs as JSON to stdout
func TailFile(filePath string) error {
	t, err := tail.TailFile(filePath, tail.Config{Follow: true, ReOpen: true})
	if err != nil {
		return err
	}

	for line := range t.Lines {
		event := LogEvent{
			Timestamp: time.Now(),
			Line:      line.Text,
			Source:    filePath,
		}

		data, _ := json.Marshal(event)
		fmt.Println(string(data)) // JSON logs to stdout
	}

	return nil
}
