package httpapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

const (
	maxSongOffset = 1_000_000
	maxSongLimit  = ports.MaxSongPageSize
	maxSongFilter = 120
)

// adminSongView is one row of the admin song list.
type adminSongView struct {
	VideoID         string `json:"videoId"`
	Title           string `json:"title"`
	Channel         string `json:"channel"`
	DurationSeconds int    `json:"durationSeconds"`
}

// adminSongsView is one page of the admin song list.
type adminSongsView struct {
	Songs  []adminSongView `json:"songs"`
	Total  int             `json:"total"`
	Offset int             `json:"offset"`
	Limit  int             `json:"limit"`
	Query  string          `json:"query"`
}

func (s *Server) handleAdminSongs(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	if ok, retry := s.limited("admin-songs:"+s.clientIP(r), 240, time.Minute); !ok {
		s.writeError(w, rateLimited("Too many admin requests.", retry))
		return
	}
	query, err := songQuery(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	page, err := s.svc.LibrarySongs(ctx, query)
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, songsView(page))
}

// songQuery reads offset, limit and q, rejecting values outside the range.
func songQuery(r *http.Request) (ports.SongQuery, error) {
	values := r.URL.Query()
	offset, err := intParam(values.Get("offset"), 0, 0, maxSongOffset)
	if err != nil {
		return ports.SongQuery{}, err
	}
	limit, err := intParam(values.Get("limit"), ports.DefaultSongPageSize, 1, maxSongLimit)
	if err != nil {
		return ports.SongQuery{}, err
	}
	query := strings.TrimSpace(values.Get("q"))
	if len([]rune(query)) > maxSongFilter {
		return ports.SongQuery{}, domain.ErrInvalidInput
	}
	return ports.SongQuery{Offset: offset, Limit: limit, Query: query}, nil
}

// intParam parses a numeric query parameter inside an inclusive range.
func intParam(raw string, fallback, low, high int) (int, error) {
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < low || value > high {
		return 0, domain.ErrInvalidInput
	}
	return value, nil
}

func songsView(page ports.SongPage) adminSongsView {
	songs := make([]adminSongView, 0, len(page.Songs))
	for _, song := range page.Songs {
		songs = append(songs, adminSongView{
			VideoID: song.VideoID, Title: song.Title, Channel: song.Channel,
			DurationSeconds: song.DurationSeconds,
		})
	}
	return adminSongsView{
		Songs: songs, Total: page.Total, Offset: page.Offset, Limit: page.Limit, Query: page.Query,
	}
}
