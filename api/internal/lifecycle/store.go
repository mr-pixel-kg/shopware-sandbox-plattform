package lifecycle

import (
	"sync"
	"time"
)

type Store struct {
	mu       sync.RWMutex
	buffers  map[string]*Buffer
	capacity int
}

func NewStore(capacity int) *Store {
	if capacity <= 0 {
		capacity = DefaultCapacity
	}
	return &Store{
		buffers:  make(map[string]*Buffer),
		capacity: capacity,
	}
}

func (s *Store) Get(containerID string) *Buffer {
	s.mu.RLock()
	b := s.buffers[containerID]
	s.mu.RUnlock()
	return b
}

func (s *Store) GetOrCreate(containerID string) *Buffer {
	s.mu.Lock()
	defer s.mu.Unlock()
	if b, ok := s.buffers[containerID]; ok {
		return b
	}
	b := NewBuffer(s.capacity)
	s.buffers[containerID] = b
	return b
}

func (s *Store) Remove(containerID string) {
	s.mu.Lock()
	delete(s.buffers, containerID)
	s.mu.Unlock()
}

func (s *Store) Log(containerID, phase string, level Level, msg string) {
	s.GetOrCreate(containerID).Write(Entry{
		Time:    time.Now(),
		Phase:   phase,
		Level:   level,
		Message: msg,
	})
}
