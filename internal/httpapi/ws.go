package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"

	"fana-karaoke/internal/app"
	"fana-karaoke/internal/domain"
)

type Hub struct {
	mgr   *app.Manager
	log   *slog.Logger
	mu    sync.Mutex
	rooms map[string]map[*wsClient]struct{}
}

type wsClient struct {
	code        string
	device      string
	role        string
	conn        *websocket.Conn
	send        chan []byte
	done        chan struct{}
	once        sync.Once
	closeCode   websocket.StatusCode
	closeReason string
}

func NewHub(mgr *app.Manager, log *slog.Logger) *Hub {
	return &Hub{mgr: mgr, log: log, rooms: make(map[string]map[*wsClient]struct{})}
}

func (h *Hub) add(c *wsClient) {
	h.mu.Lock()
	if h.rooms[c.code] == nil {
		h.rooms[c.code] = make(map[*wsClient]struct{})
	}
	h.rooms[c.code][c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) remove(c *wsClient) {
	h.mu.Lock()
	if set := h.rooms[c.code]; set != nil {
		delete(set, c)
		if len(set) == 0 {
			delete(h.rooms, c.code)
		}
	}
	h.mu.Unlock()
}

func (h *Hub) Broadcast(code string) {
	h.mu.Lock()
	clients := make([]*wsClient, 0, len(h.rooms[code]))
	for c := range h.rooms[code] {
		clients = append(clients, c)
	}
	h.mu.Unlock()
	for _, c := range clients {
		snap, err := h.mgr.Snapshot(code, c.device, c.role)
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

func (h *Hub) Ended(code string) {
	h.mu.Lock()
	clients := make([]*wsClient, 0, len(h.rooms[code]))
	for c := range h.rooms[code] {
		clients = append(clients, c)
	}
	delete(h.rooms, code)
	h.mu.Unlock()
	msg := []byte(`{"type":"roomEnded"}`)
	for _, c := range clients {
		c.closeCode = websocket.StatusCode(4001)
		c.closeReason = "room ended"
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

func (c *wsClient) close() {
	c.once.Do(func() { close(c.done) })
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(r.URL.Query().Get("room"))
	if code == "" {
		s.writeError(w, domain.ErrRoomNotFound)
		return
	}
	if ok, _ := s.limited("ws:"+s.clientIP(r), 10, time.Minute); !ok {
		s.writeError(w, &domain.AppError{Code: "rate_limited", Message: "Too many connections.", Status: 429})
		return
	}
	role := "guest"
	if s.isHost(r, code) {
		if r.URL.Query().Get("client") == "tv" {
			role = "tv"
		} else {
			role = "phone"
		}
	}
	device := s.ensureDevice(w, r)
	if role != "guest" {
		s.mgr.HostSeen(code)
	}
	snap, err := s.mgr.Snapshot(code, device, role)
	if err != nil {
		s.writeError(w, err)
		return
	}
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: s.dev,
	})
	if err != nil {
		s.log.Debug("ws accept", "error", err)
		return
	}
	conn.SetReadLimit(4096)
	c := &wsClient{
		code:   code,
		device: device,
		role:   role,
		conn:   conn,
		send:   make(chan []byte, 16),
		done:   make(chan struct{}),
	}
	s.hub.add(c)
	go s.writeWS(c)
	h, err := json.Marshal(struct {
		Type string `json:"type"`
		*domain.Snapshot
	}{Type: "snapshot", Snapshot: snap})
	if err == nil {
		s.hub.enqueue(c, h)
	}
	s.readWS(c)
}

func (s *Server) readWS(c *wsClient) {
	ctx := context.Background()
	defer func() {
		c.close()
		s.hub.remove(c)
		_ = c.conn.Close(websocket.StatusNormalClosure, "")
	}()
	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			return
		}
		var msg struct {
			Type       string  `json:"type"`
			EntryID    string  `json:"entryId"`
			Generation uint64  `json:"generation"`
			Position   float64 `json:"positionSeconds"`
			Duration   float64 `json:"durationSeconds"`
			Code       string  `json:"code"`
		}
		if json.Unmarshal(data, &msg) != nil {
			continue
		}
		switch msg.Type {
		case "heartbeat":
			s.mgr.TouchDevice(c.device, c.role != "guest")
			if c.role != "guest" {
				s.mgr.HostSeen(c.code)
			}
		case "tvProgress":
			if c.role == "tv" {
				s.mgr.TVProgress(c.code, msg.EntryID, msg.Generation, msg.Position, msg.Duration)
			}
		case "tvEnded":
			if c.role == "tv" {
				s.mgr.TVEnded(c.code, msg.EntryID, msg.Generation, c.device)
			}
		case "tvError":
			if c.role == "tv" {
				reason := msg.Code
				if reason == "" {
					reason = "error"
				}
				s.mgr.TVError(c.code, msg.EntryID, reason, msg.Generation)
			}
		}
	}
}

func (s *Server) writeWS(c *wsClient) {
	ctx := context.Background()
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.done:
			code := c.closeCode
			reason := c.closeReason
			if code == 0 {
				code = websocket.StatusNormalClosure
				reason = "bye"
			}
			_ = c.conn.Close(code, reason)
			return
		case data := <-c.send:
			writeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := c.conn.Write(writeCtx, websocket.MessageText, data)
			cancel()
			if err != nil {
				c.close()
				return
			}
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := c.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				c.close()
				return
			}
		}
	}
}
