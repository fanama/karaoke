package memory

import (
	"errors"
	"strings"
	"sync"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

// errGone is returned when an update targets a room that no longer exists.
var errGone = errors.New("room gone")

// RoomRepo keeps rooms in memory behind an atomic read-modify-write API.
type RoomRepo struct {
	mu     sync.RWMutex
	rooms  map[string]domain.Room
	tokens map[string]string
}

// NewRoomRepo builds an empty repository.
func NewRoomRepo() *RoomRepo {
	return &RoomRepo{rooms: map[string]domain.Room{}, tokens: map[string]string{}}
}

// Count reports how many rooms exist.
func (r *RoomRepo) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.rooms)
}

// Create stores a room unless its code is already taken.
func (r *RoomRepo) Create(room domain.Room) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.rooms[room.Code]; exists {
		return domain.ErrServerFull
	}
	r.rooms[room.Code] = room
	r.tokens[room.HostToken] = room.Code
	return nil
}

// Get returns a copy of a room.
func (r *RoomRepo) Get(code string) (domain.Room, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	room, ok := r.rooms[code]
	return room, ok
}

// Exists reports whether a room code is live.
func (r *RoomRepo) Exists(code string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.rooms[code]
	return ok
}

// Resolve accepts a canonical code or the same code without separators.
func (r *RoomRepo) Resolve(input string) (string, bool) {
	input = strings.ToLower(strings.TrimSpace(input))
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, ok := r.rooms[input]; ok {
		return input, true
	}
	compact := domain.NormalizeCompactCode(input)
	for code := range r.rooms {
		if domain.NormalizeCompactCode(code) == compact {
			return code, true
		}
	}
	return "", false
}

// ByHostToken maps a secret TV token onto its room code.
func (r *RoomRepo) ByHostToken(token string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	code, ok := r.tokens[token]
	return code, ok
}

// Update applies fn atomically and stores the result. If fn returns an error
// nothing is written and the error is returned to the caller.
func (r *RoomRepo) Update(code string, fn func(domain.Room) (domain.Room, error)) (domain.Room, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	room, ok := r.rooms[code]
	if !ok {
		return domain.Room{}, domain.ErrRoomNotFound
	}
	next, err := fn(room.Clone())
	if err != nil {
		return room, err
	}
	if next.Code == "" {
		next.Code = code
	}
	r.rooms[code] = next
	return next, nil
}

// Delete removes a room and frees its host token.
func (r *RoomRepo) Delete(code string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	room, ok := r.rooms[code]
	if !ok {
		return false
	}
	delete(r.rooms, code)
	delete(r.tokens, room.HostToken)
	return true
}

// Codes lists every live room code.
func (r *RoomRepo) Codes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.rooms))
	for code := range r.rooms {
		out = append(out, code)
	}
	return out
}

var _ ports.RoomRepository = (*RoomRepo)(nil)
