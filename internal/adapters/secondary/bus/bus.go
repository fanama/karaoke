package bus

import (
	"sync"

	"fana-karaoke/internal/ports"
)

// Bus fans room events out to subscribers without coupling them to each other.
type Bus struct {
	mu      sync.RWMutex
	updated []func(string)
	ended   []func(string)
}

// New builds an empty bus.
func New() *Bus { return &Bus{} }

// OnRoomUpdated registers a subscriber for room change events.
func (b *Bus) OnRoomUpdated(fn func(string)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.updated = append(b.updated, fn)
}

// OnRoomEnded registers a subscriber for room deletion events.
func (b *Bus) OnRoomEnded(fn func(string)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ended = append(b.ended, fn)
}

// PublishRoomUpdated notifies subscribers asynchronously.
func (b *Bus) PublishRoomUpdated(code string) {
	b.mu.RLock()
	fns := append([]func(string){}, b.updated...)
	b.mu.RUnlock()
	for _, fn := range fns {
		go fn(code)
	}
}

// PublishRoomEnded notifies subscribers asynchronously.
func (b *Bus) PublishRoomEnded(code string) {
	b.mu.RLock()
	fns := append([]func(string){}, b.ended...)
	b.mu.RUnlock()
	for _, fn := range fns {
		go fn(code)
	}
}

var _ ports.EventPublisher = (*Bus)(nil)
