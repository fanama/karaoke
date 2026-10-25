package httpapi

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

const (
	defaultSuggestionCount = 6
	maxSuggestionCount     = 20
)

// handleRandomSongs serves the guest search screen's default song vignettes.
func (s *Server) handleRandomSongs(w http.ResponseWriter, r *http.Request) {
	if ok, retry := s.limited("random:"+s.clientIP(r), 60, time.Minute); !ok {
		s.writeError(w, rateLimited("Too many suggestion requests.", retry))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	results, err := s.svc.RandomSongs(ctx, suggestionLimit(r))
	if err != nil {
		s.writeError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	s.writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

// suggestionLimit reads the optional limit, clamped to the supported range.
func suggestionLimit(r *http.Request) int {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return defaultSuggestionCount
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return defaultSuggestionCount
	}
	if value > maxSuggestionCount {
		return maxSuggestionCount
	}
	return value
}
