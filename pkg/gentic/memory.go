package gentic

import (
	"strings"
	"sync"
)

// ThreadStore maps conversation thread IDs to isolated Memory instances.
// Implementations must be safe for concurrent use from multiple goroutines.
type ThreadStore interface {
	// Get returns the Memory for threadID, creating storage on first access.
	// If threadID is empty after trimming whitespace, Get returns nil.
	Get(threadID string) Memory
}

// InMemoryThreadStore is a thread-safe registry of per-thread InMemoryStorage.
// It uses sync.Map for the thread-ID index; each InMemoryStorage has its own mutex.
type InMemoryThreadStore struct {
	m sync.Map // string -> *InMemoryStorage
}

// NewInMemoryThreadStore creates an empty thread store.
func NewInMemoryThreadStore() *InMemoryThreadStore {
	return &InMemoryThreadStore{}
}

// Get implements [ThreadStore].
func (s *InMemoryThreadStore) Get(threadID string) Memory {
	tid := strings.TrimSpace(threadID)
	if tid == "" {
		return nil
	}
	v, _ := s.m.LoadOrStore(tid, NewInMemoryStorage())
	return v.(Memory)
}

// Memory is an interface for storing and retrieving conversation messages.
// Implementations can be in-memory, database-backed, or custom.
type Memory interface {
	// Append adds a message to storage.
	Append(msg Message) error

	// Messages returns all stored messages in chronological order.
	Messages() ([]Message, error)

	// Clear removes all stored messages.
	Clear() error
}

// InMemoryStorage is a thread-safe, in-memory implementation of Memory.
// Messages are stored in a slice and lost when the process exits.
type InMemoryStorage struct {
	mu       sync.RWMutex
	messages []Message
}

// NewInMemoryStorage creates a new in-memory message store.
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		messages: []Message{},
	}
}

// Append adds a message to the in-memory store.
func (s *InMemoryStorage) Append(msg Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, msg)
	return nil
}

// Messages returns a copy of all stored messages.
func (s *InMemoryStorage) Messages() ([]Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Return a copy to prevent external mutation
	result := make([]Message, len(s.messages))
	copy(result, s.messages)
	return result, nil
}

// Clear removes all stored messages.
func (s *InMemoryStorage) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = []Message{}
	return nil
}
