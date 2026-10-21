package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"fana-karaoke/internal/domain"
)

const securityPolicy = "default-src 'self'; script-src 'self' https://www.youtube.com https://s.ytimg.com; " +
	"frame-src https://www.youtube.com https://www.youtube-nocookie.com; " +
	"img-src 'self' data: https://i.ytimg.com https://img.youtube.com; style-src 'self' 'unsafe-inline'; " +
	"connect-src 'self'; font-src 'self'; object-src 'none'; base-uri 'none'; " +
	"frame-ancestors 'none'; form-action 'self'"

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		s.setSecurityHeaders(w)
		if r.Method != http.MethodGet && r.Method != http.MethodHead && !s.sameOrigin(r) {
			s.writeError(w, domain.ErrForbiddenOrigin)
			return
		}
		defer s.recoverPanic(w, r)
		next.ServeHTTP(w, r)
		s.log.Debug("request", "method", r.Method, "path", logPath(r.URL.Path), "duration_ms", time.Since(start).Milliseconds())
	})
}

func (s *Server) setSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// strict-origin-when-cross-origin keeps the secret /tv/<token> path out of
	// cross-origin referrers while still giving the YouTube iframe a referrer.
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
	w.Header().Set("Content-Security-Policy", securityPolicy)
}

func (s *Server) recoverPanic(w http.ResponseWriter, r *http.Request) {
	if rec := recover(); rec != nil {
		s.log.Error("panic", "error", rec, "path", logPath(r.URL.Path))
		s.writeError(w, errors.New("internal error"))
	}
}

func logPath(p string) string {
	if strings.HasPrefix(p, "/tv/") {
		return "/tv/<secret>"
	}
	return p
}

func (s *Server) sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" || s.dev {
		return true
	}
	if s.baseURL != "" {
		return strings.EqualFold(strings.TrimRight(origin, "/"), s.baseURL)
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return strings.EqualFold(origin, scheme+"://"+r.Host)
}

func (s *Server) origin(r *http.Request) string {
	if s.baseURL != "" {
		return s.baseURL
	}
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}
