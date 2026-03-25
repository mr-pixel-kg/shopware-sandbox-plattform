package registry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

const defaultExecTimeout = 5 * time.Minute

type Executor struct {
	Client *client.Client
}

func (e *Executor) RunPostStart(ctx context.Context, containerID string, commands []ExecCommand) {
	slog.Debug("post-start starting", "container_id", containerID, "commands", len(commands))
	for i, cmd := range commands {
		slog.Debug("post-start executing", "container_id", containerID, "index", i, "command", cmd.Command, "delay", cmd.Delay.Duration, "timeout", cmd.Timeout.Duration)
		if err := e.execWithRetry(ctx, containerID, cmd); err != nil {
			slog.Warn("post-start command failed", "container_id", containerID, "index", i, "command", cmd.Command, "error", err)
		} else {
			slog.Debug("post-start command succeeded", "container_id", containerID, "index", i)
		}
	}
	slog.Debug("post-start finished", "container_id", containerID)
}

func (e *Executor) RunPreStop(ctx context.Context, containerID string, commands []ExecCommand) {
	slog.Debug("pre-stop starting", "container_id", containerID, "commands", len(commands))
	for i, cmd := range commands {
		slog.Debug("pre-stop executing", "container_id", containerID, "index", i, "command", cmd.Command)
		if err := e.execWithRetry(ctx, containerID, cmd); err != nil {
			slog.Warn("pre-stop command failed", "container_id", containerID, "index", i, "command", cmd.Command, "error", err)
		} else {
			slog.Debug("pre-stop command succeeded", "container_id", containerID, "index", i)
		}
	}
	slog.Debug("pre-stop finished", "container_id", containerID)
}

func (e *Executor) execWithRetry(ctx context.Context, containerID string, cmd ExecCommand) error {
	if cmd.Delay.Duration > 0 {
		slog.Debug("exec waiting for delay", "container_id", containerID, "delay", cmd.Delay.Duration)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(cmd.Delay.Duration):
		}
	}

	attempts := 1 + cmd.Retries
	var lastErr error

	for attempt := range attempts {
		if attempt > 0 && cmd.RetryDelay.Duration > 0 {
			slog.Debug("exec retry waiting", "container_id", containerID, "attempt", attempt+1, "retry_delay", cmd.RetryDelay.Duration)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(cmd.RetryDelay.Duration):
			}
		}

		start := time.Now()
		lastErr = e.exec(ctx, containerID, cmd)
		elapsed := time.Since(start)

		if lastErr == nil {
			slog.Debug("exec succeeded", "container_id", containerID, "attempt", attempt+1, "elapsed", elapsed)
			return nil
		}

		slog.Debug("exec attempt failed", "container_id", containerID, "attempt", attempt+1, "max_attempts", attempts, "elapsed", elapsed, "error", lastErr)
	}

	return lastErr
}

func (e *Executor) exec(ctx context.Context, containerID string, cmd ExecCommand) error {
	timeout := cmd.Timeout.Duration
	if timeout <= 0 {
		timeout = defaultExecTimeout
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp, err := e.Client.ContainerExecCreate(execCtx, containerID, container.ExecOptions{
		Cmd:          cmd.Command,
		AttachStdout: false,
		AttachStderr: false,
	})
	if err != nil {
		return fmt.Errorf("create exec: %w", err)
	}

	if err := e.Client.ContainerExecStart(execCtx, resp.ID, container.ExecStartOptions{}); err != nil {
		return fmt.Errorf("start exec: %w", err)
	}

	time.Sleep(50 * time.Millisecond)

	for {
		inspect, err := e.Client.ContainerExecInspect(execCtx, resp.ID)
		if err != nil {
			return fmt.Errorf("inspect exec: %w", err)
		}
		if !inspect.Running {
			if inspect.ExitCode != 0 {
				return fmt.Errorf("exec exited with code %d", inspect.ExitCode)
			}
			return nil
		}

		select {
		case <-execCtx.Done():
			return execCtx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
}
