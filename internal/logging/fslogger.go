package logging

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type LogEventChannel chan LogEvent

// TailFileFS tails a single file and sends events to the central channel
func TailFileFS(filePath string, out LogEventChannel) error {
	file, err := os.Open(filePath)
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
	watcher.Add(filePath)

	for {
		// Read new lines
		line, err := reader.ReadString('\n')
		if err == nil {
			line = strings.TrimSpace(line)
			event := ParseLog(line, filePath, "", "", "") // Parser from parser.go
			out <- event
		}

		// Watch for file changes
		select {
		case <-watcher.Events:
			continue
		case err := <-watcher.Errors:
			// log error and continue
			println("Watcher error:", err.Error())
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// WatchContainerLogs watches /var/log/containers and tails new log files
func WatchContainerLogs(dir string, out LogEventChannel) error {
	// Tail existing files first
	files, err := filepath.Glob(filepath.Join(dir, "*.log"))
	if err != nil {
		return err
	}

	for _, file := range files {
		go TailFileFS(file, out)
	}

	// Watch the directory for new files
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	err = watcher.Add(dir)
	if err != nil {
		return err
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Create == fsnotify.Create {
				// New file created, start tailing
				fmt.Println("New log file detected:", event.Name)
				go TailFileFS(event.Name, out)
			}
		case err := <-watcher.Errors:
			fmt.Println("Watcher error:", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
