package logging

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hpcloud/tail"
)

type LogEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	Line      string                 `json:"line"`
	Pod       string                 `json:"pod,omitempty"`
	Container string                 `json:"container,omitempty"`
	Namespace string                 `json:"namespace,omitempty"`
	Source    string                 `json:"source,omitempty"` // file or docker id
	Extra     map[string]interface{} `json:"extra,omitempty"`  // parsed JSON log
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
