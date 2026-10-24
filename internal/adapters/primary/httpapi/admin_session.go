package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"fana-karaoke/internal/domain"
)

const (
	adminCookie = "fana_admin"
	adminTTL    = 24 * time.Hour
)

func (s *Server) adminEnabled() bool { return s.adminToken != "" }

func (s *Server) setAdminCookie(w http.ResponseWriter) {
	value := "admin|" + strconv.FormatInt(time.Now().Add(adminTTL).Unix(), 10)
	s.setCookie(w, adminCookie, s.signed(value), adminTTL)
}

func (s *Server) clearAdminCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     adminCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.secureCookies(),
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Server) isAdmin(r *http.Request) bool {
	if !s.adminEnabled() {
		return false
	}
	c, err := r.Cookie(adminCookie)
	if err != nil {
		return false
	}
	value, ok := s.signer.Verify(c.Value)
	if !ok {
		return false
	}
	parts := strings.Split(value, "|")
	if len(parts) != 2 || parts[0] != "admin" {
		return false
	}
	exp, err := strconv.ParseInt(parts[1], 10, 64)
	return err == nil && time.Now().Unix() <= exp
}

func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if !s.adminEnabled() {
		s.writeError(w, domain.ErrAdminDisabled)
		return false
	}
	if !s.isAdmin(r) {
		s.writeError(w, domain.ErrAdminRequired)
		return false
	}
	return true
}
