package httpapi

import (
	"context"
	"crypto/subtle"
	"net/http"
	"time"

	"fana-karaoke/internal/domain"
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
	s.writeLibraryExport(w, r)
}

func (s *Server) handleAdminImport(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	if ok, retry := s.limited("admin-library:"+s.clientIP(r), 10, time.Hour); !ok {
		s.writeError(w, rateLimited("Too many library imports.", retry))
		return
	}
	s.importLibrary(w, r)
}

// adminSessionView is the admin page state; the count is only sent after login.
type adminSessionView struct {
	Enabled       bool `json:"enabled"`
	Authenticated bool `json:"authenticated"`
	Library       bool `json:"library"`
	Songs         int  `json:"songs"`
}

func (s *Server) writeAdminSession(w http.ResponseWriter, r *http.Request, authenticated bool) {
	view := adminSessionView{Enabled: s.adminEnabled(), Authenticated: authenticated}
	if authenticated {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if total, err := s.svc.LibraryStats(ctx); err == nil {
			view.Library = true
			view.Songs = total
		}
	}
	s.writeJSON(w, http.StatusOK, view)
}
