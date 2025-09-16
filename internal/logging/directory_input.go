package logging

import (
	"bufio"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DirectoryInput tails all files in a directory
type DirectoryInput struct {
	Dir     string
	stopCh  chan struct{}
	wg      sync.WaitGroup
	started bool
}

func NewDirectoryInput(dir string) *DirectoryInput {
	return &DirectoryInput{
		Dir:    dir,
		stopCh: make(chan struct{}),
	}
}

// Start now matches the Input interface
func (d *DirectoryInput) Start(out LogEventChannel) error {
	if d.started {
		return nil
	}
	d.started = true

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for {
			select {
			case <-d.stopCh:
				return
			default:
				d.scanAndTail(out)
				time.Sleep(5 * time.Second) // rescan for new files
			}
		}
	}()

	return nil
}

func (d *DirectoryInput) scanAndTail(out LogEventChannel) {
	files, err := filepath.Glob(filepath.Join(d.Dir, "*.log"))
	if err != nil {
		return
	}

	for _, f := range files {
		go d.tailFile(f, out) // tail each file concurrently
	}
}

func (d *DirectoryInput) tailFile(path string, out LogEventChannel) {
	d.wg.Add(1)
	defer d.wg.Done()

	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		select {
		case <-d.stopCh:
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				time.Sleep(500 * time.Millisecond) // wait for new lines
				continue
			}

			event := ParseLog(line, path, "", "", "") // use parser
			out <- event
		}
	}
}

func (d *DirectoryInput) Stop() {
	close(d.stopCh)
	d.wg.Wait()
}
