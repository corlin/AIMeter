package common

import (
	"encoding/json"
	"os"
	"sync"
)

// ReadSeedBytes attempts to locate and read the raw bytes of a configuration or seed file
// across multiple candidate directory depths (current, ../, ../../, ../../../).
func ReadSeedBytes(relativePath string) ([]byte, error) {
	candidates := []string{
		relativePath,
		"../" + relativePath,
		"../../" + relativePath,
		"../../../" + relativePath,
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
	}
	return nil, os.ErrNotExist
}

// LoadSeedFile attempts to locate and deserialize a JSON seed file by searching
// across multiple candidate directory depths (current, ../, ../../, ../../../).
func LoadSeedFile[T any](relativePath string, target *T) error {
	data, err := ReadSeedBytes(relativePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// RingBuffer is a thread-safe circular buffer that keeps the most recent N items.
type RingBuffer[T any] struct {
	mu       sync.RWMutex
	items    []T
	capacity int
}

// NewRingBuffer initializes a thread-safe ring buffer with given capacity.
func NewRingBuffer[T any](capacity int) *RingBuffer[T] {
	if capacity <= 0 {
		capacity = 100
	}
	return &RingBuffer[T]{
		items:    make([]T, 0, capacity),
		capacity: capacity,
	}
}

// Push adds an item to the buffer, evicting the oldest item if full.
func (r *RingBuffer[T]) Push(item T) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.items) >= r.capacity {
		r.items = r.items[1:]
	}
	r.items = append(r.items, item)
}

// List returns the latest items in reverse chronological order (newest first).
func (r *RingBuffer[T]) List(limit int) []T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	n := len(r.items)
	if limit <= 0 || limit > n {
		limit = n
	}

	result := make([]T, limit)
	for i := 0; i < limit; i++ {
		result[i] = r.items[n-1-i]
	}
	return result
}

// All returns all stored items in chronological order.
func (r *RingBuffer[T]) All() []T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]T, len(r.items))
	copy(result, r.items)
	return result
}

// Len returns the current number of items.
func (r *RingBuffer[T]) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items)
}
