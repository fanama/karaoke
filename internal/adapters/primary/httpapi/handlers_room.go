package httpapi

import (
	"net/http"
	"time"

	"fana-karaoke/internal/domain"
)

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleConfig(w http.ResponseWriter, _ *http.Request) {
	cfg := s.svc.Config()
	s.writeJSON(w, http.StatusOK, map[string]any{
		"maxGuests": cfg.MaxGuests,
		"maxQueue":  cfg.MaxQueue,
		"languages": []string{"en", "fr"},
	})
}

func (s *Server) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	if ok, retry := s.limited("create:"+s.clientIP(r), 5, time.Hour); !ok {
		s.writeError(w, rateLimited("Too many rooms created.", retry))
		return
	}
	code, token, pin, err := s.svc.CreateRoom(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	base := s.origin(r)
	s.writeJSON(w, http.StatusCreated, map[string]string{
		"code":    code,
		"joinUrl": base + "/r/" + code,
		"tvUrl":   base + "/tv/" + token,
		"pin":     pin,
	})
}

func (s *Server) handleResolveRoom(w http.ResponseWriter, r *http.Request) {
	if ok, retry := s.limited("resolve:"+s.clientIP(r), 60, time.Minute); !ok {
		s.writeError(w, rateLimited("Too many requests.", retry))
		return
	}
	code, ok := s.svc.ResolveCode(r.URL.Query().Get("code"))
	if !ok {
		s.writeError(w, domain.ErrRoomNotFound)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"code": code})
}

func (s *Server) handleTVSession(w http.ResponseWriter, r *http.Request) {
	code, ok := s.tvRoomFrom(r)
	if !ok {
		s.writeError(w, domain.ErrRoomNotFound)
		return
	}
	pin, exists := s.svc.PIN(code)
	if !exists {
		s.writeError(w, domain.ErrRoomNotFound)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"code": code, "pin": pin})
}

func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	code := roomCode(r)
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
	w.Header().Set("Cache-Control", "no-store")
	s.writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleQR(w http.ResponseWriter, r *http.Request) {
	code := roomCode(r)
	if !s.svc.Exists(code) {
		s.writeError(w, domain.ErrRoomNotFound)
		return
	}
	svg, err := s.qr.EncodeSVG(s.origin(r) + "/r/" + code)
	if err != nil {
		s.writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(svg)
}

func (s *Server) handleEndRoom(w http.ResponseWriter, r *http.Request) {
	code := roomCode(r)
	if !s.requireHost(w, r, code) {
		return
	}
	if err := s.svc.EndRoom(code); err != nil {
		s.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
