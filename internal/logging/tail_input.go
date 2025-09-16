package logging

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// TailInput tails a file and streams logs
type TailInput struct {
	Path string
}

func NewTailInput(path string) *TailInput {
	return &TailInput{Path: path}
}

// Start begins tailing and sends LogEvents to the channel
func (t *TailInput) Start(out LogEventChannel) error {
	file, err := os.Open(t.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Move to the end of the file
	file.Seek(0, 2)
	reader := bufio.NewReader(file)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	watcher.Add(t.Path)

	for {
		// Try reading new lines
		line, err := reader.ReadString('\n')
		if err == nil {
			line = strings.TrimSpace(line)
			event := ParseLog(line, t.Path, "", "", "")
			out <- event
		}

		// Watch for changes
		select {
		case <-watcher.Events:
			continue
		case err := <-watcher.Errors:
			println("Watcher error:", err.Error())
		}
		time.Sleep(50 * time.Millisecond)
	}
}
