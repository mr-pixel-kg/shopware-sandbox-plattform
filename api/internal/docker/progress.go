package docker

import (
	"encoding/json"
	"io"
	"log/slog"
	"sync"
)

type PullEvent struct {
	Status         string         `json:"status"`
	ID             string         `json:"id"`
	ProgressDetail ProgressDetail `json:"progressDetail"`
	Error          string         `json:"error,omitempty"`
}

type ProgressDetail struct {
	Current int64 `json:"current"`
	Total   int64 `json:"total"`
}

type PullProgress struct {
	Percent int    `json:"percent"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

type PullTracker struct {
	mu       sync.RWMutex
	pulls    map[string]*pullState
	watchers map[string][]chan PullProgress
}

type pullState struct {
	layers            map[string]*layerProgress
	highWaterMark     int
	lastLoggedPercent int
	done              bool
	final             *PullProgress
}

type layerProgress struct {
	downloadCurrent int64
	downloadTotal   int64
	downloadDone    bool
	extractCurrent  int64
	extractTotal    int64
	done            bool
}

func NewPullTracker() *PullTracker {
	return &PullTracker{
		pulls:    make(map[string]*pullState),
		watchers: make(map[string][]chan PullProgress),
	}
}

func (t *PullTracker) Start(imageID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.pulls[imageID] = &pullState{layers: make(map[string]*layerProgress)}
}

func (t *PullTracker) Progress(imageID string) PullProgress {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.computeProgress(imageID)
}

func (t *PullTracker) Watch(imageID string) (<-chan PullProgress, func()) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state := t.pulls[imageID]

	if state == nil || state.done {
		ch := make(chan PullProgress, 1)
		if state != nil && state.final != nil {
			ch <- *state.final
		}
		close(ch)
		return ch, func() {}
	}

	ch := make(chan PullProgress, 16)
	t.watchers[imageID] = append(t.watchers[imageID], ch)

	return ch, func() {
		t.mu.Lock()
		defer t.mu.Unlock()
		watchers := t.watchers[imageID]
		for i, w := range watchers {
			if w == ch {
				t.watchers[imageID] = append(watchers[:i], watchers[i+1:]...)
				break
			}
		}
	}
}

func (t *PullTracker) ConsumePullStream(imageID string, reader io.Reader) error {
	decoder := json.NewDecoder(reader)
	for {
		var event PullEvent
		if err := decoder.Decode(&event); err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		if event.Error != "" {
			return &pullStreamError{msg: event.Error}
		}
		t.updateLayer(imageID, &event)
		t.notifyWatchers(imageID)
	}
}

func (t *PullTracker) Finish(imageID string, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	progress := PullProgress{Percent: 100, Status: "complete"}
	if err != nil {
		progress = t.computeProgress(imageID)
		progress.Status = "failed"
		progress.Error = err.Error()
	}

	if state, ok := t.pulls[imageID]; ok {
		state.done = true
		state.final = &progress
	}

	for _, ch := range t.watchers[imageID] {
		select {
		case ch <- progress:
		default:
		}
		close(ch)
	}
	delete(t.watchers, imageID)
}

func (t *PullTracker) Remove(imageID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.pulls, imageID)
	delete(t.watchers, imageID)
}

func (t *PullTracker) updateLayer(imageID string, event *PullEvent) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state, ok := t.pulls[imageID]
	if !ok || event.ID == "" {
		return
	}

	layer, exists := state.layers[event.ID]
	if !exists {
		layer = &layerProgress{}
		state.layers[event.ID] = layer
	}

	switch event.Status {
	case "Downloading":
		layer.downloadCurrent = event.ProgressDetail.Current
		if event.ProgressDetail.Total > 0 {
			layer.downloadTotal = event.ProgressDetail.Total
		}
	case "Download complete":
		layer.downloadDone = true
		if layer.downloadTotal > 0 {
			layer.downloadCurrent = layer.downloadTotal
		}
	case "Extracting":
		layer.downloadDone = true
		if layer.downloadTotal > 0 {
			layer.downloadCurrent = layer.downloadTotal
		}
		layer.extractCurrent = event.ProgressDetail.Current
		if event.ProgressDetail.Total > 0 {
			layer.extractTotal = event.ProgressDetail.Total
		}
	case "Pull complete", "Already exists":
		layer.done = true
	}
}

func (t *PullTracker) notifyWatchers(imageID string) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	progress := t.computeProgress(imageID)

	if state, ok := t.pulls[imageID]; ok && progress.Percent != state.lastLoggedPercent {
		slog.Info("image pull progress", "image_id", imageID, "percent", progress.Percent)
		state.lastLoggedPercent = progress.Percent
	}

	for _, ch := range t.watchers[imageID] {
		select {
		case ch <- progress:
		default:
		}
	}
}

func (t *PullTracker) computeProgress(imageID string) PullProgress {
	state, ok := t.pulls[imageID]
	if !ok {
		return PullProgress{Status: "pulling"}
	}

	var totalBytes, currentBytes int64
	for _, l := range state.layers {
		layerTotal := l.downloadTotal + l.extractTotal
		if layerTotal == 0 {
			continue
		}
		totalBytes += layerTotal
		if l.done {
			currentBytes += layerTotal
		} else {
			currentBytes += l.downloadCurrent + l.extractCurrent
		}
	}

	var percent int
	if totalBytes > 0 {
		percent = int(currentBytes * 100 / totalBytes)
	}
	if percent > 99 {
		percent = 99
	}

	if percent > state.highWaterMark {
		state.highWaterMark = percent
	} else {
		percent = state.highWaterMark
	}

	return PullProgress{Percent: percent, Status: "pulling"}
}

type pullStreamError struct {
	msg string
}

func (e *pullStreamError) Error() string {
	return e.msg
}
