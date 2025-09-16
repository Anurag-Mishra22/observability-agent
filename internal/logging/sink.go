package logging

import (
	"encoding/json"
	"fmt"
	"os"
)

// Sink is the interface that all log sinks implement
type Sink interface {
	Write(event LogEvent) error
}

// --------------------
// StdoutSink
// --------------------
type StdoutSink struct{}

func (s *StdoutSink) Write(event LogEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// --------------------
// FileSink
// --------------------
type FileSink struct {
	Path string
	file *os.File
}

func NewFileSink(path string) (*FileSink, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &FileSink{Path: path, file: f}, nil
}

func (s *FileSink) Write(event LogEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = s.file.WriteString(string(data) + "\n")
	return err
}
