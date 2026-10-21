package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"

	"fana-karaoke/internal/domain"
)

// wsMessage is a client-to-server telemetry envelope.
type wsMessage struct {
	Type       string  `json:"type"`
	EntryID    string  `json:"entryId"`
	Generation uint64  `json:"generation"`
	Position   float64 `json:"positionSeconds"`
	Duration   float64 `json:"durationSeconds"`
	Code       string  `json:"code"`
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(r.URL.Query().Get("room"))
	if code == "" {
		s.writeError(w, domain.ErrRoomNotFound)
		return
	}
	if ok, _ := s.limited("ws:"+s.clientIP(r), 10, time.Minute); !ok {
		s.writeError(w, rateLimited("Too many connections.", 0))
		return
	}
	role := s.roleFor(r, code)
	device := s.ensureDevice(w, r)
	if role != "guest" {
		s.svc.HostSeen(code)
	}
	snap, err := s.svc.Snapshot(code, device, role)
	if err != nil {
		s.writeError(w, err)
		return
	}
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: s.dev})
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
	go c.writePump()
	s.sendInitial(c, snap)
	s.readWS(c)
}

func (s *Server) sendInitial(c *wsClient, snap *domain.Snapshot) {
	data, err := json.Marshal(struct {
		Type string `json:"type"`
		*domain.Snapshot
	}{Type: "snapshot", Snapshot: snap})
	if err == nil {
		s.hub.enqueue(c, data)
	}
}

func (s *Server) readWS(c *wsClient) {
	defer func() {
		c.close()
		s.hub.remove(c)
		_ = c.conn.Close(websocket.StatusNormalClosure, "")
	}()
	for {
		_, data, err := c.conn.Read(context.Background())
		if err != nil {
			return
		}
		var msg wsMessage
		if json.Unmarshal(data, &msg) != nil {
			continue
		}
		s.handleWSMessage(c, msg)
	}
}

func (s *Server) handleWSMessage(c *wsClient, msg wsMessage) {
	switch msg.Type {
	case "heartbeat":
		s.svc.TouchDevice(c.device, c.role != "guest")
		if c.role != "guest" {
			s.svc.HostSeen(c.code)
		}
	case "tvProgress":
		if c.role == "tv" {
			s.svc.TVProgress(c.code, msg.EntryID, msg.Generation, msg.Position, msg.Duration)
		}
	case "tvEnded":
		if c.role == "tv" {
			s.svc.TVEnded(c.code, msg.EntryID, msg.Generation, c.device)
		}
	case "tvError":
		if c.role == "tv" {
			reason := msg.Code
			if reason == "" {
				reason = "error"
			}
			s.svc.TVError(c.code, msg.EntryID, reason, msg.Generation)
		}
	}
}
