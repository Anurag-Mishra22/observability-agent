package logging

import (
	"github.com/hpcloud/tail"
)

// TailFile tails a given log file and sends logs to Sink
func TailFile(filePath string) error {
	t, err := tail.TailFile(filePath, tail.Config{Follow: true, ReOpen: true})
	if err != nil {
		return err
	}

	for line := range t.Lines {
		// Parse the log line into a LogEvent using parser.go
		parsed := ParseLog(line.Text, filePath, "", "", "")
		// Send the parsed log to Sink (stdout or other configured outputs)
		Sink(parsed)
	}

	return nil
}
