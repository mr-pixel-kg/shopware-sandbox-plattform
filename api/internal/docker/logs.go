package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
)

func (c *DockerClient) ContainerLogs(ctx context.Context, containerID string) (io.ReadCloser, error) {
	reader, err := c.client.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return nil, fmt.Errorf("container logs %s: %w", containerID, err)
	}
	return reader, nil
}

type execReadCloser struct {
	reader io.Reader
	conn   io.Closer
}

func (e *execReadCloser) Read(p []byte) (int, error) {
	return e.reader.Read(p)
}

func (e *execReadCloser) Close() error {
	return e.conn.Close()
}

func (c *DockerClient) ExecFollow(ctx context.Context, containerID string, cmd []string) (io.ReadCloser, error) {
	execResp, err := c.client.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          cmd,
		Tty:          false,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("create exec: %w", err)
	}

	attachResp, err := c.client.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{
		Tty: false,
	})
	if err != nil {
		return nil, fmt.Errorf("attach exec: %w", err)
	}

	return &execReadCloser{reader: attachResp.Reader, conn: attachResp.Conn}, nil
}
