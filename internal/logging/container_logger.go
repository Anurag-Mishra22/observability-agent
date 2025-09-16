package logging

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

func TailContainerLogs(containerID string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
		Tail:       "0",
	}

	out, err := cli.ContainerLogs(context.Background(), containerID, options)
	if err != nil {
		return err
	}
	defer out.Close()

	// Create pipes for stdout and stderr
	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	// Demultiplex in a goroutine
	go func() {
		_, _ = stdcopy.StdCopy(stdoutWriter, stderrWriter, out)
		stdoutWriter.Close()
		stderrWriter.Close()
	}()

	// Read stdout line by line
	go func() {
		scanner := bufio.NewScanner(stdoutReader)
		for scanner.Scan() {
			line := scanner.Text()
			event := LogEvent{
				Timestamp: time.Now(),
				Line:      line,
				Source:    containerID,
			}
			data, _ := json.Marshal(event)
			fmt.Println(string(data))
		}
	}()

	// Read stderr line by line
	go func() {
		scanner := bufio.NewScanner(stderrReader)
		for scanner.Scan() {
			line := scanner.Text()
			event := LogEvent{
				Timestamp: time.Now(),
				Line:      line,
				Source:    containerID + "-stderr",
				Container: containerID,
			}
			data, _ := json.Marshal(event)
			fmt.Println(string(data))
		}
	}()

	// Block forever so tail keeps running
	select {}
}
