package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
)

type ExecAttachOptions struct {
	Cmd  []string
	Cols uint
	Rows uint
}

type ExecSession struct {
	execID string
	conn   io.WriteCloser
	reader io.Reader
	client *DockerClient
}

func (s *ExecSession) Read(p []byte) (int, error) {
	return s.reader.Read(p)
}

func (s *ExecSession) Write(p []byte) (int, error) {
	return s.conn.Write(p)
}

func (s *ExecSession) Resize(ctx context.Context, cols, rows uint) error {
	return s.client.client.ContainerExecResize(ctx, s.execID, container.ResizeOptions{
		Width:  cols,
		Height: rows,
	})
}

func (s *ExecSession) Close() error {
	return s.conn.Close()
}

func (s *ExecSession) Wait(ctx context.Context) (int, error) {
	for {
		inspect, err := s.client.client.ContainerExecInspect(ctx, s.execID)
		if err != nil {
			return -1, fmt.Errorf("inspect exec: %w", err)
		}
		if !inspect.Running {
			return inspect.ExitCode, nil
		}
		select {
		case <-ctx.Done():
			return -1, ctx.Err()
		default:
		}
	}
}

func (c *DockerClient) CreateExecSession(ctx context.Context, containerID string, opts ExecAttachOptions) (*ExecSession, error) {
	cmd := opts.Cmd
	if len(cmd) == 0 {
		cmd = []string{"/bin/bash"}
	}

	cols := opts.Cols
	if cols == 0 {
		cols = 80
	}
	rows := opts.Rows
	if rows == 0 {
		rows = 24
	}

	execResp, err := c.client.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          cmd,
		Tty:          true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("create exec: %w", err)
	}

	attachResp, err := c.client.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{
		Tty: true,
	})
	if err != nil {
		return nil, fmt.Errorf("attach exec: %w", err)
	}

	if err := c.client.ContainerExecResize(ctx, execResp.ID, container.ResizeOptions{
		Width:  cols,
		Height: rows,
	}); err != nil {
		attachResp.Close()
		return nil, fmt.Errorf("resize exec: %w", err)
	}

	return &ExecSession{
		execID: execResp.ID,
		conn:   attachResp.Conn,
		reader: attachResp.Reader,
		client: c,
	}, nil
}
