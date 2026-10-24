package httpapi

import (
	"context"
	"crypto/subtle"
	"net/http"
	"time"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

func (s *Server) handleAdminSession(w http.ResponseWriter, r *http.Request) {
	s.writeAdminSession(w, r, s.isAdmin(r))
}

func (s *Server) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	if !s.adminEnabled() {
		s.writeError(w, domain.ErrAdminDisabled)
		return
	}
	if ok, retry := s.limited("admin-login:"+s.clientIP(r), 5, 15*time.Minute); !ok {
		s.writeError(w, rateLimited("Too many login attempts.", retry))
		return
	}
	var body struct {
		Token string `json:"token"`
	}
	if err := s.decodeJSON(w, r, &body); err != nil {
		s.writeError(w, domain.ErrInvalidInput)
		return
	}
	if subtle.ConstantTimeCompare([]byte(body.Token), []byte(s.adminToken)) != 1 {
		s.writeError(w, domain.ErrInvalidToken)
		return
	}
	s.setAdminCookie(w)
	s.writeAdminSession(w, r, true)
}

func (s *Server) handleAdminLogout(w http.ResponseWriter, _ *http.Request) {
	s.clearAdminCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAdminExport(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	format, err := s.libraryFormat(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeLibraryExport(w, r, format)
}

func (s *Server) handleAdminImport(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	if ok, retry := s.limited("admin-library:"+s.clientIP(r), 10, time.Hour); !ok {
		s.writeError(w, rateLimited("Too many library imports.", retry))
		return
	}
	format, err := s.libraryFormat(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.importLibrary(w, r, format)
}

// adminSessionView is the admin page state; the count is only sent after login.
type adminSessionView struct {
	Enabled       bool              `json:"enabled"`
	Authenticated bool              `json:"authenticated"`
	Library       bool              `json:"library"`
	Songs         int               `json:"songs"`
	Storage       *adminStorageView `json:"storage,omitempty"`
}

type adminStorageView struct {
	Backend    string `json:"backend"`
	Connected  bool   `json:"connected"`
	CacheOnly  bool   `json:"cacheOnly"`
	Songs      int    `json:"songs"`
	LastSyncAt string `json:"lastSyncAt"`
	LastError  string `json:"lastError"`
}

func (s *Server) writeAdminSession(w http.ResponseWriter, r *http.Request, authenticated bool) {
	view := adminSessionView{Enabled: s.adminEnabled(), Authenticated: authenticated}
	if authenticated {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if status, err := s.svc.LibraryInfo(ctx); err == nil {
			view.Library = true
			view.Songs = status.Songs
			view.Storage = storageView(status)
		}
	}
	s.writeJSON(w, http.StatusOK, view)
}

func storageView(status ports.LibraryStatus) *adminStorageView {
	last := ""
	if !status.LastSyncAt.IsZero() {
		last = status.LastSyncAt.UTC().Format(time.RFC3339)
	}
	return &adminStorageView{
		Backend: status.Backend, Connected: status.Connected, CacheOnly: status.CacheOnly,
		Songs: status.Songs, LastSyncAt: last, LastError: status.LastError,
	}
}
