package logging

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// TailFileFS tails a log file and prints new lines as JSON
func TailFileFS(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Move to the end of the file so we only read new lines
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
		if err == nil {
			line = strings.TrimSpace(line) // remove \n and spaces
			event := LogEvent{
				Timestamp: time.Now(),
				Line:      line,
				Source:    filePath,
			}
			data, _ := json.Marshal(event)
			fmt.Println(string(data))
		}

		// Watch for file changes
		select {
		case <-watcher.Events:
			// File changed, continue reading new lines
			continue
		case err := <-watcher.Errors:
			fmt.Println("Watcher error:", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
