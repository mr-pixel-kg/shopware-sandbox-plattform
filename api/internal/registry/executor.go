package registry

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/mr-pixel-kg/shopshredder/api/internal/lifecycle"
)

const defaultExecTimeout = 5 * time.Minute

type postStartState struct {
	done   chan struct{}
	failed bool
}

type Executor struct {
	Client    *client.Client
	Lifecycle *lifecycle.Store

	mu        sync.Mutex
	postStart map[string]*postStartState
}

func (e *Executor) PostStartDone(containerID string) bool {
	e.mu.Lock()
	state, ok := e.postStart[containerID]
	e.mu.Unlock()
	if !ok {
		return true
	}
	select {
	case <-state.done:
		return true
	default:
		return false
	}
}

func (e *Executor) PostStartFailed(containerID string) bool {
	e.mu.Lock()
	state, ok := e.postStart[containerID]
	e.mu.Unlock()
	if !ok {
		return false
	}
	select {
	case <-state.done:
		return state.failed
	default:
		return false
	}
}

func (e *Executor) PostStartWait(containerID string) <-chan struct{} {
	e.mu.Lock()
	defer e.mu.Unlock()
	if state, ok := e.postStart[containerID]; ok {
		return state.done
	}
	return nil
}

func (e *Executor) RunPostStart(ctx context.Context, containerID string, commands []ExecCommand) {
	state := &postStartState{done: make(chan struct{})}
	e.mu.Lock()
	if e.postStart == nil {
		e.postStart = make(map[string]*postStartState)
	}
	e.postStart[containerID] = state
	e.mu.Unlock()

	defer func() {
		close(state.done)
	}()

	buf := e.getBuffer(containerID)
	failed := e.runCommands(ctx, containerID, commands, buf, "post_start")

	e.mu.Lock()
	state.failed = failed > 0
	e.mu.Unlock()
}

func (e *Executor) RunPreStop(ctx context.Context, containerID string, commands []ExecCommand) {
	buf := e.getBuffer(containerID)
	e.runCommands(ctx, containerID, commands, buf, "pre_stop")
}

func (e *Executor) runCommands(ctx context.Context, containerID string, commands []ExecCommand, buf *lifecycle.Buffer, phase string) int {
	total := len(commands)
	totalStart := time.Now()
	slog.Debug(phase+" starting", "container_id", containerID, "commands", total)

	failed := 0
	for i, cmd := range commands {
		step := fmt.Sprintf("[%d/%d]", i+1, total)
		name := cmd.Label
		rawCmd := strings.Join(cmd.Command, " ")

		if name != "" {
			buf.Write(lifecycle.Entry{Time: time.Now(), Phase: phase, Level: lifecycle.LevelInfo, Message: step + " " + name})
			buf.Write(lifecycle.Entry{Time: time.Now(), Phase: phase, Level: lifecycle.LevelDetail, Message: rawCmd})
		} else {
			name = rawCmd
			buf.Write(lifecycle.Entry{Time: time.Now(), Phase: phase, Level: lifecycle.LevelInfo, Message: step + " " + name})
		}

		slog.Debug(phase+" executing", "container_id", containerID, "index", i, "command", cmd.Command)

		start := time.Now()
		if err := e.execWithRetry(ctx, containerID, cmd, buf, phase); err != nil {
			elapsed := time.Since(start).Truncate(100 * time.Millisecond)
			buf.Write(lifecycle.Entry{Time: time.Now(), Phase: phase, Level: lifecycle.LevelError, Message: fmt.Sprintf("%s failed (%s) — %v", step, elapsed, err)})
			slog.Warn(phase+" command failed", "container_id", containerID, "index", i, "command", cmd.Command, "error", err)
			failed++
		} else {
			elapsed := time.Since(start).Truncate(100 * time.Millisecond)
			buf.Write(lifecycle.Entry{Time: time.Now(), Phase: phase, Level: lifecycle.LevelSuccess, Message: fmt.Sprintf("%s done (%s)", step, elapsed)})
			slog.Debug(phase+" command succeeded", "container_id", containerID, "index", i)
		}
	}

	totalElapsed := time.Since(totalStart).Truncate(100 * time.Millisecond)
	if failed == 0 {
		buf.Write(lifecycle.Entry{Time: time.Now(), Phase: phase, Level: lifecycle.LevelInfo, Message: fmt.Sprintf("All %d commands completed (%s)", total, totalElapsed)})
	} else {
		buf.Write(lifecycle.Entry{Time: time.Now(), Phase: phase, Level: lifecycle.LevelError, Message: fmt.Sprintf("%d/%d commands failed (%s)", failed, total, totalElapsed)})
	}
	slog.Debug(phase+" finished", "container_id", containerID)
	return failed
}

func (e *Executor) execWithRetry(ctx context.Context, containerID string, cmd ExecCommand, buf *lifecycle.Buffer, phase string) error {
	if cmd.Delay.Duration > 0 {
		buf.Write(lifecycle.Entry{Time: time.Now(), Phase: phase, Level: lifecycle.LevelWait, Message: fmt.Sprintf("Waiting %s", cmd.Delay.Duration)})
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
			buf.Write(lifecycle.Entry{Time: time.Now(), Phase: phase, Level: lifecycle.LevelWait, Message: fmt.Sprintf("Retry %d/%d in %s", attempt, cmd.Retries, cmd.RetryDelay.Duration)})
			slog.Debug("exec retry waiting", "container_id", containerID, "attempt", attempt+1, "retry_delay", cmd.RetryDelay.Duration)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(cmd.RetryDelay.Duration):
			}
		}

		start := time.Now()
		lastErr = e.exec(ctx, containerID, cmd, buf, phase)
		elapsed := time.Since(start)

		if lastErr == nil {
			slog.Debug("exec succeeded", "container_id", containerID, "attempt", attempt+1, "elapsed", elapsed)
			return nil
		}

		slog.Debug("exec attempt failed", "container_id", containerID, "attempt", attempt+1, "max_attempts", attempts, "elapsed", elapsed, "error", lastErr)
	}

	return lastErr
}

func (e *Executor) exec(ctx context.Context, containerID string, cmd ExecCommand, buf *lifecycle.Buffer, phase string) error {
	timeout := cmd.Timeout.Duration
	if timeout <= 0 {
		timeout = defaultExecTimeout
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp, err := e.Client.ContainerExecCreate(execCtx, containerID, container.ExecOptions{
		Cmd:          cmd.Command,
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return fmt.Errorf("create exec: %w", err)
	}

	attachResp, err := e.Client.ContainerExecAttach(execCtx, resp.ID, container.ExecAttachOptions{
		Tty: true,
	})
	if err != nil {
		return fmt.Errorf("attach exec: %w", err)
	}

	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		scanner := bufio.NewScanner(attachResp.Reader)
		scanner.Buffer(make([]byte, 0, 64*1024), 64*1024)
		for scanner.Scan() {
			if line := scanner.Text(); line != "" {
				buf.Write(lifecycle.Entry{Time: time.Now(), Phase: phase, Level: lifecycle.LevelOutput, Message: line})
			}
		}
		if err := scanner.Err(); err != nil && ctx.Err() == nil {
			slog.Debug("exec scanner error", "container_id", containerID, "error", err)
		}
	}()

	select {
	case <-readDone:
		attachResp.Close()
	case <-execCtx.Done():
		attachResp.Close()
		<-readDone
		return fmt.Errorf("timed out after %s", timeout)
	}

	inspect, err := e.Client.ContainerExecInspect(context.Background(), resp.ID)
	if err != nil {
		return fmt.Errorf("inspect exec: %w", err)
	}
	if inspect.ExitCode != 0 {
		return fmt.Errorf("exit code %d", inspect.ExitCode)
	}
	return nil
}

func (e *Executor) getBuffer(containerID string) *lifecycle.Buffer {
	return e.Lifecycle.GetOrCreate(containerID)
}
