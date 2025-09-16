package logging

import (
	"bufio"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// TailFileFS tails a log file and sends new lines to Sink
func TailFileFS(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Move to the end of the file to only read new lines
	file.Seek(0, io.SeekEnd)
	reader := bufio.NewReader(file)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	watcher.Add(filePath)

	for {
		// Try reading new lines
		line, err := reader.ReadString('\n')
		if err == nil || err == io.EOF {
			line = strings.TrimSpace(line)
			if line != "" {
				parsed := ParseLog(line, "file", "", "", filePath)
				Sink(parsed)
			}
		}

		// Watch for file changes
		select {
		case <-watcher.Events:
			// File changed, continue reading new lines
			continue
		case err := <-watcher.Errors:
			parsed := ParseLog(err.Error(), "watcher-error", "", "", filePath)
			Sink(parsed)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
