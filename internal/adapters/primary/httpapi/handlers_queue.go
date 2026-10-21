package httpapi

import (
	"context"
	"net/http"
	"time"

	"fana-karaoke/internal/domain"
)

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if ok, retry := s.limited("search:"+s.clientIP(r), 30, time.Minute); !ok {
		s.writeError(w, rateLimited("Too many searches.", retry))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	results, stale, err := s.svc.Search(ctx, r.URL.Query().Get("q"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"results": results, "stale": stale})
}

func (s *Server) handleAddSong(w http.ResponseWriter, r *http.Request) {
	code := roomCode(r)
	role := s.roleFor(r, code)
	device := s.ensureDevice(w, r)
	if role != "guest" {
		s.svc.HostSeen(code)
	}
	if ok, retry := s.limited("add:"+device, 10, time.Minute); !ok {
		s.writeError(w, rateLimited("Too many songs added.", retry))
		return
	}
	var body struct {
		VideoID  string `json:"videoId"`
		URL      string `json:"url"`
		Nickname string `json:"nickname"`
	}
	if err := s.decodeJSON(w, r, &body); err != nil {
		s.writeError(w, domain.ErrInvalidInput)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()
	snap, err := s.svc.AddSong(ctx, code, device, role, body.Nickname, body.VideoID, body.URL)
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusCreated, snap)
}

func (s *Server) handleVote(w http.ResponseWriter, r *http.Request) {
	code := roomCode(r)
	s.writeSnapshot(w, r, func(device, role string) (*domain.Snapshot, error) {
		return s.svc.ToggleVote(code, device, role, r.PathValue("entryID"))
	})
}

func (s *Server) handleRemoveSong(w http.ResponseWriter, r *http.Request) {
	code := roomCode(r)
	if !s.requireHost(w, r, code) {
		return
	}
	s.writeSnapshot(w, r, func(device, role string) (*domain.Snapshot, error) {
		return s.svc.RemoveSong(code, r.PathValue("entryID"), device, role)
	})
}

func (s *Server) handleReorder(w http.ResponseWriter, r *http.Request) {
	code := roomCode(r)
	if !s.requireHost(w, r, code) {
		return
	}
	var body struct {
		EntryIDs []string `json:"entryIds"`
	}
	if err := s.decodeJSON(w, r, &body); err != nil {
		s.writeError(w, domain.ErrInvalidInput)
		return
	}
	s.writeSnapshot(w, r, func(device, role string) (*domain.Snapshot, error) {
		return s.svc.Reorder(code, body.EntryIDs, device, role)
	})
}

func (s *Server) handleClearQueue(w http.ResponseWriter, r *http.Request) {
	code := roomCode(r)
	if !s.requireHost(w, r, code) {
		return
	}
	s.writeSnapshot(w, r, func(device, role string) (*domain.Snapshot, error) {
		return s.svc.ClearQueue(code, device, role)
	})
}

func (s *Server) handleLockQueue(w http.ResponseWriter, r *http.Request) {
	code := roomCode(r)
	if !s.requireHost(w, r, code) {
		return
	}
	var body struct {
		Locked bool `json:"locked"`
	}
	if err := s.decodeJSON(w, r, &body); err != nil {
		s.writeError(w, domain.ErrInvalidInput)
		return
	}
	s.writeSnapshot(w, r, func(device, role string) (*domain.Snapshot, error) {
		return s.svc.LockQueue(code, body.Locked, device, role)
	})
}

// writeSnapshot resolves the device and role, runs a mutation, and responds.
func (s *Server) writeSnapshot(w http.ResponseWriter, r *http.Request, call func(device, role string) (*domain.Snapshot, error)) {
	code := roomCode(r)
	device := s.ensureDevice(w, r)
	snap, err := call(device, s.roleFor(r, code))
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, snap)
}
