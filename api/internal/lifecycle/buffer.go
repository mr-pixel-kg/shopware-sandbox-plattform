package lifecycle

import (
	"sync"
	"time"
)

const DefaultCapacity = 10_000

type Level string

const (
	LevelInfo    Level = "info"
	LevelSuccess Level = "success"
	LevelError   Level = "error"
	LevelOutput  Level = "output"
	LevelDetail  Level = "detail"
	LevelWait    Level = "wait"
)

type Entry struct {
	Time    time.Time
	Phase   string
	Level   Level
	Message string
}

type Buffer struct {
	mu       sync.RWMutex
	entries  []Entry
	capacity int
	pos      int
	count    int
	subs     map[int]chan Entry
	nextSub  int
}

func NewBuffer(capacity int) *Buffer {
	if capacity <= 0 {
		capacity = DefaultCapacity
	}
	return &Buffer{
		entries:  make([]Entry, capacity),
		capacity: capacity,
		subs:     make(map[int]chan Entry),
	}
}

func (b *Buffer) Write(e Entry) {
	b.mu.Lock()
	b.entries[b.pos] = e
	b.pos = (b.pos + 1) % b.capacity
	if b.count < b.capacity {
		b.count++
	}

	subs := make([]chan Entry, 0, len(b.subs))
	for _, ch := range b.subs {
		subs = append(subs, ch)
	}
	b.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- e:
		default:
		}
	}
}

func (b *Buffer) Snapshot() []Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.snapshotLocked()
}

func (b *Buffer) SnapshotAndSubscribe() ([]Entry, <-chan Entry, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	snapshot := b.snapshotLocked()

	ch := make(chan Entry, 256)
	id := b.nextSub
	b.nextSub++
	b.subs[id] = ch

	cancel := func() {
		b.mu.Lock()
		delete(b.subs, id)
		b.mu.Unlock()
	}

	return snapshot, ch, cancel
}

func (b *Buffer) snapshotLocked() []Entry {
	result := make([]Entry, 0, b.count)
	if b.count < b.capacity {
		result = append(result, b.entries[:b.count]...)
	} else {
		result = append(result, b.entries[b.pos:]...)
		result = append(result, b.entries[:b.pos]...)
	}
	return result
}
