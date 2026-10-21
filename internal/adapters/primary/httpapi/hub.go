package httpapi

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

// Hub tracks connected clients per room and fans snapshots out to them.
type Hub struct {
	svc   ports.KaraokeService
	log   *slog.Logger
	mu    sync.Mutex
	rooms map[string]map[*wsClient]struct{}
}

// NewHub builds an empty hub over the karaoke service.
func NewHub(svc ports.KaraokeService, log *slog.Logger) *Hub {
	return &Hub{svc: svc, log: log, rooms: map[string]map[*wsClient]struct{}{}}
}

func (h *Hub) add(c *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[c.code] == nil {
		h.rooms[c.code] = map[*wsClient]struct{}{}
	}
	h.rooms[c.code][c] = struct{}{}
}

func (h *Hub) remove(c *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set := h.rooms[c.code]; set != nil {
		delete(set, c)
		if len(set) == 0 {
			delete(h.rooms, c.code)
		}
	}
}

func (h *Hub) clients(code string) []*wsClient {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]*wsClient, 0, len(h.rooms[code]))
	for c := range h.rooms[code] {
		out = append(out, c)
	}
	return out
}

// PublishRoomUpdated implements ports.EventPublisher.
func (h *Hub) PublishRoomUpdated(code string) { h.Broadcast(code) }

// PublishRoomEnded implements ports.EventPublisher.
func (h *Hub) PublishRoomEnded(code string) { h.Ended(code) }

// Broadcast sends the latest snapshot to every client in a room.
func (h *Hub) Broadcast(code string) {
	for _, c := range h.clients(code) {
		snap, err := h.svc.Snapshot(code, c.device, c.role)
		if err != nil {
			continue
		}
		data, err := json.Marshal(struct {
			Type string `json:"type"`
			*domain.Snapshot
		}{Type: "snapshot", Snapshot: snap})
		if err != nil {
			continue
		}
		h.enqueue(c, data)
	}
}

// Ended tells every client that the room is gone and closes their sockets.
func (h *Hub) Ended(code string) {
	clients := h.clients(code)
	h.mu.Lock()
	delete(h.rooms, code)
	h.mu.Unlock()
	msg := []byte(`{"type":"roomEnded"}`)
	for _, c := range clients {
		c.closeWith(websocket.StatusCode(4001), "room ended")
		h.enqueue(c, msg)
		go func(c *wsClient) {
			time.Sleep(250 * time.Millisecond)
			c.close()
		}(c)
	}
}

func (h *Hub) enqueue(c *wsClient, data []byte) {
	select {
	case c.send <- data:
	default:
		c.close()
	}
}

var _ ports.EventPublisher = (*Hub)(nil)
