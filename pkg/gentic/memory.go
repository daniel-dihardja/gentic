package gentic

import "sync"

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
