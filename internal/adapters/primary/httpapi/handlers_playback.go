package httpapi

import (
	"net/http"
	"strings"
	"time"

	"fana-karaoke/internal/domain"
)

func (s *Server) handlePlayback(w http.ResponseWriter, r *http.Request) {
	code := roomCode(r)
	if !s.requireHost(w, r, code) {
		return
	}
	var body struct {
		Action string `json:"action"`
	}
	if err := s.decodeJSON(w, r, &body); err != nil {
		s.writeError(w, domain.ErrInvalidInput)
		return
	}
	s.writeSnapshot(w, r, func(device, role string) (*domain.Snapshot, error) {
		return s.svc.Playback(code, body.Action, device, role)
	})
}

func (s *Server) handleStartSkipVote(w http.ResponseWriter, r *http.Request) {
	code := roomCode(r)
	device := s.ensureDevice(w, r)
	if ok, retry := s.limited("vote:"+device, 10, time.Minute); !ok {
		s.writeError(w, rateLimited("Too many votes.", retry))
		return
	}
	snap, err := s.svc.StartSkipVote(code, device, s.roleFor(r, code))
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleCastSkipVote(w http.ResponseWriter, r *http.Request) {
	code := roomCode(r)
	s.writeSnapshot(w, r, func(device, role string) (*domain.Snapshot, error) {
		return s.svc.CastSkipVote(code, device, role)
	})
}

func (s *Server) handleCancelSkipVote(w http.ResponseWriter, r *http.Request) {
	code := roomCode(r)
	if !s.requireHost(w, r, code) {
		return
	}
	device := s.ensureDevice(w, r)
	snap, err := s.svc.CancelSkipVote(code, s.roleFor(r, code), device)
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleHostUnlock(w http.ResponseWriter, r *http.Request) {
	code := roomCode(r)
	if ok, retry := s.limited("unlock:"+s.clientIP(r)+":"+code, 5, 10*time.Minute); !ok {
		s.writeError(w, rateLimited("Too many PIN attempts.", retry))
		return
	}
	pin, exists := s.svc.PIN(code)
	if !exists {
		s.writeError(w, domain.ErrRoomNotFound)
		return
	}
	var body struct {
		PIN string `json:"pin"`
	}
	if err := s.decodeJSON(w, r, &body); err != nil {
		s.writeError(w, domain.ErrInvalidInput)
		return
	}
	if !domain.ConstantTimeEqual(pin, strings.TrimSpace(body.PIN)) {
		s.writeError(w, domain.ErrInvalidPIN)
		return
	}
	device := s.ensureDevice(w, r)
	s.setHostCookie(w, code)
	s.svc.HostSeen(code)
	s.svc.TouchDevice(device, true)
	snap, err := s.svc.Snapshot(code, device, "phone")
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleRotatePIN(w http.ResponseWriter, r *http.Request) {
	code := roomCode(r)
	if !s.requireHost(w, r, code) {
		return
	}
	pin, err := s.svc.RotatePIN(code)
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"pin": pin})
}
