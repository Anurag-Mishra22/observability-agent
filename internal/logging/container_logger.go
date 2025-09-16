package logging

import (
	"bufio"
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// TailContainerLogs streams logs from a container and sends them to Sink
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

	// Demultiplex stdout/stderr
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
			parsed := ParseLog(line, containerID, "", containerID, "")
			Sink(parsed)
		}
	}()

	// Read stderr line by line
	go func() {
		scanner := bufio.NewScanner(stderrReader)
		for scanner.Scan() {
			line := scanner.Text()
			parsed := ParseLog(line, containerID+"-stderr", "", containerID, "")
			Sink(parsed)
		}
	}()

	// Block forever
	select {}
}
