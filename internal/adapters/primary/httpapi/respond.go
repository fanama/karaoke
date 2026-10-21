package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"fana-karaoke/internal/domain"
)

func (s *Server) decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	return nil
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		s.log.Debug("write response", "error", err)
	}
}

func (s *Server) writeError(w http.ResponseWriter, err error) {
	var appErr *domain.AppError
	if !errors.As(err, &appErr) {
		appErr = domain.Err("internal_error", "Something went wrong.", 500)
	}
	body := map[string]any{
		"error": map[string]any{
			"code":    appErr.Code,
			"message": appErr.Message,
		},
	}
	if appErr.RetryAfter > 0 {
		body["error"].(map[string]any)["retry_after"] = appErr.RetryAfter
		w.Header().Set("Retry-After", fmt.Sprintf("%d", appErr.RetryAfter))
	}
	s.writeJSON(w, appErr.Status, body)
}

func (s *Server) limited(key string, limit int, window time.Duration) (bool, int) {
	if s.dev {
		return true, 0
	}
	return s.limiter.Allow(key, limit, window)
}

// roomCode normalises the path parameter used by every room route.
func roomCode(r *http.Request) string { return strings.ToLower(r.PathValue("code")) }

// rateLimited builds the throttling error returned by every guarded route.
func rateLimited(message string, retryAfter int) *domain.AppError {
	err := domain.Err("rate_limited", message, http.StatusTooManyRequests)
	err.RetryAfter = retryAfter
	return err
}

// requireHost rejects a request that lacks a valid host cookie.
func (s *Server) requireHost(w http.ResponseWriter, r *http.Request, code string) bool {
	if !s.isHost(r, code) {
		s.writeError(w, domain.ErrHostRequired)
		return false
	}
	s.svc.HostSeen(code)
	return true
}
