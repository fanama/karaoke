package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"

	"fana-karaoke/internal/app"
	"fana-karaoke/internal/domain"
)

type Searcher interface {
	Search(ctx context.Context, query string) ([]domain.SearchResult, bool, error)
}

type Config struct {
	Dev        bool
	BaseURL    string
	Secret     []byte
	TrustProxy bool
	Version    string
}

type Server struct {
	mgr        *app.Manager
	searcher   Searcher
	hub        *Hub
	limiter    *Limiter
	static     fs.FS
	index      []byte
	baseURL    string
	dev        bool
	secret     []byte
	trustProxy bool
	log        *slog.Logger
}

func New(mgr *app.Manager, searcher Searcher, static fs.FS, cfg Config, log *slog.Logger) (*Server, error) {
	if log == nil {
		log = slog.Default()
	}
	s := &Server{
		mgr:        mgr,
		searcher:   searcher,
		limiter:    NewLimiter(),
		static:     static,
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		dev:        cfg.Dev,
		secret:     cfg.Secret,
		trustProxy: cfg.TrustProxy,
		log:        log,
	}
	if static != nil {
		if b, err := fs.ReadFile(static, "index.html"); err == nil {
			s.index = b
		}
	}
	if len(s.index) == 0 {
		s.index = []byte("<!doctype html><title>Fana Karaoke</title><p>The frontend has not been built.</p>")
	}
	s.hub = NewHub(mgr, log)
	mgr.SetCallbacks(s.hub.Broadcast, s.hub.Ended)
	return s, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /api/config", s.handleConfig)
	mux.HandleFunc("POST /api/rooms", s.handleCreateRoom)
	mux.HandleFunc("GET /api/tv/session", s.handleTVSession)
	mux.HandleFunc("GET /api/rooms/resolve", s.handleResolveRoom)
	mux.HandleFunc("GET /api/rooms/{code}", s.handleSnapshot)
	mux.HandleFunc("GET /api/rooms/{code}/qr.svg", s.handleQR)
	mux.HandleFunc("GET /api/search", s.handleSearch)
	mux.HandleFunc("POST /api/rooms/{code}/songs", s.handleAddSong)
	mux.HandleFunc("POST /api/rooms/{code}/songs/{entryID}/vote", s.handleVote)
	mux.HandleFunc("POST /api/rooms/{code}/skip-vote", s.handleStartSkipVote)
	mux.HandleFunc("POST /api/rooms/{code}/skip-vote/cast", s.handleCastSkipVote)
	mux.HandleFunc("POST /api/rooms/{code}/host/unlock", s.handleHostUnlock)
	mux.HandleFunc("POST /api/rooms/{code}/playback", s.handlePlayback)
	mux.HandleFunc("DELETE /api/rooms/{code}/songs/{entryID}", s.handleRemoveSong)
	mux.HandleFunc("PUT /api/rooms/{code}/order", s.handleReorder)
	mux.HandleFunc("POST /api/rooms/{code}/queue/clear", s.handleClearQueue)
	mux.HandleFunc("POST /api/rooms/{code}/queue/lock", s.handleLockQueue)
	mux.HandleFunc("POST /api/rooms/{code}/skip-vote/cancel", s.handleCancelSkipVote)
	mux.HandleFunc("POST /api/rooms/{code}/pin/rotate", s.handleRotatePIN)
	mux.HandleFunc("POST /api/rooms/{code}/end", s.handleEndRoom)
	mux.HandleFunc("GET /ws", s.handleWS)
	mux.HandleFunc("GET /assets/", s.handleAsset)
	mux.HandleFunc("GET /favicon.svg", s.handleFavicon)
	mux.HandleFunc("GET /", s.handleSPA)
	return s.middleware(mux)
}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// strict-origin-when-cross-origin keeps the secret /tv/<token> path out of
		// cross-origin referrers (only the origin is sent) while still giving the
		// YouTube iframe a referrer. YouTube rejects embeds without one (error 153).
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' https://www.youtube.com https://s.ytimg.com; frame-src https://www.youtube.com https://www.youtube-nocookie.com; img-src 'self' data: https://i.ytimg.com https://img.youtube.com; style-src 'self' 'unsafe-inline'; connect-src 'self'; font-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			if !s.sameOrigin(r) {
				s.writeError(w, domain.Err("forbidden_origin", "Cross-origin request refused.", http.StatusForbidden))
				return
			}
		}
		defer func() {
			if rec := recover(); rec != nil {
				s.log.Error("panic", "error", rec, "path", logPath(r.URL.Path))
				s.writeError(w, errors.New("internal error"))
			}
		}()
		next.ServeHTTP(w, r)
		s.log.Debug("request", "method", r.Method, "path", logPath(r.URL.Path), "duration_ms", time.Since(start).Milliseconds())
	})
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

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	cfg := s.mgr.Config()
	s.writeJSON(w, http.StatusOK, map[string]any{
		"maxGuests": cfg.MaxGuests,
		"maxQueue":  cfg.MaxQueue,
		"languages": []string{"en", "fr"},
	})
}

func (s *Server) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	if ok, retry := s.limited("create:"+s.clientIP(r), 5, time.Hour); !ok {
		s.writeError(w, &domain.AppError{Code: "rate_limited", Message: "Too many rooms created.", Status: 429, RetryAfter: retry})
		return
	}
	code, token, pin, err := s.mgr.CreateRoom()
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
		s.writeError(w, &domain.AppError{Code: "rate_limited", Message: "Too many requests.", Status: 429, RetryAfter: retry})
		return
	}
	code, ok := s.mgr.ResolveCode(r.URL.Query().Get("code"))
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
	pin, exists := s.mgr.PIN(code)
	if !exists {
		s.writeError(w, domain.ErrRoomNotFound)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"code": code, "pin": pin})
}

func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(r.PathValue("code"))
	role := s.roleFor(r, code)
	device := s.ensureDevice(w, r)
	if role != "guest" {
		s.mgr.HostSeen(code)
	}
	snap, err := s.mgr.Snapshot(code, device, role)
	if err != nil {
		s.writeError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	s.writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if ok, retry := s.limited("search:"+s.clientIP(r), 30, time.Minute); !ok {
		s.writeError(w, &domain.AppError{Code: "rate_limited", Message: "Too many searches.", Status: 429, RetryAfter: retry})
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len([]rune(q)) < 2 {
		s.writeJSON(w, http.StatusOK, map[string]any{"results": []domain.SearchResult{}, "stale": false})
		return
	}
	if s.searcher == nil {
		s.writeError(w, domain.ErrSearchFailed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	results, stale, err := s.searcher.Search(ctx, q)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if results == nil {
		results = []domain.SearchResult{}
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"results": results, "stale": stale})
}

func (s *Server) handleAddSong(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(r.PathValue("code"))
	role := s.roleFor(r, code)
	device := s.ensureDevice(w, r)
	if role != "guest" {
		s.mgr.HostSeen(code)
	}
	if ok, retry := s.limited("add:"+device, 10, time.Minute); !ok {
		s.writeError(w, &domain.AppError{Code: "rate_limited", Message: "Too many songs added.", Status: 429, RetryAfter: retry})
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
	snap, err := s.mgr.AddSong(ctx, code, device, role, body.Nickname, body.VideoID, body.URL)
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusCreated, snap)
}

func (s *Server) handleVote(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(r.PathValue("code"))
	role := s.roleFor(r, code)
	device := s.ensureDevice(w, r)
	snap, err := s.mgr.ToggleVote(code, device, role, r.PathValue("entryID"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleStartSkipVote(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(r.PathValue("code"))
	role := s.roleFor(r, code)
	device := s.ensureDevice(w, r)
	if ok, retry := s.limited("vote:"+device, 10, time.Minute); !ok {
		s.writeError(w, &domain.AppError{Code: "rate_limited", Message: "Too many votes.", Status: 429, RetryAfter: retry})
		return
	}
	snap, err := s.mgr.StartSkipVote(code, device, role)
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleCastSkipVote(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(r.PathValue("code"))
	role := s.roleFor(r, code)
	device := s.ensureDevice(w, r)
	snap, err := s.mgr.CastSkipVote(code, device, role)
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleCancelSkipVote(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(r.PathValue("code"))
	if !s.requireHost(w, r, code) {
		return
	}
	device := s.ensureDevice(w, r)
	snap, err := s.mgr.CancelSkipVote(code, s.roleFor(r, code), device)
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleHostUnlock(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(r.PathValue("code"))
	if ok, retry := s.limited("unlock:"+s.clientIP(r)+":"+code, 5, 10*time.Minute); !ok {
		s.writeError(w, &domain.AppError{Code: "pin_locked", Message: "Too many PIN attempts.", Status: 429, RetryAfter: retry})
		return
	}
	pin, exists := s.mgr.PIN(code)
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
	if !constantTimeEqual(pin, strings.TrimSpace(body.PIN)) {
		s.writeError(w, domain.ErrInvalidPIN)
		return
	}
	device := s.ensureDevice(w, r)
	s.setHostCookie(w, code)
	s.mgr.HostSeen(code)
	s.mgr.TouchDevice(device, true)
	snap, err := s.mgr.Snapshot(code, device, "phone")
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handlePlayback(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(r.PathValue("code"))
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
	device := s.ensureDevice(w, r)
	snap, err := s.mgr.Playback(code, body.Action, device, s.roleFor(r, code))
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleRemoveSong(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(r.PathValue("code"))
	if !s.requireHost(w, r, code) {
		return
	}
	device := s.ensureDevice(w, r)
	snap, err := s.mgr.RemoveSong(code, r.PathValue("entryID"), device, s.roleFor(r, code))
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleReorder(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(r.PathValue("code"))
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
	device := s.ensureDevice(w, r)
	snap, err := s.mgr.Reorder(code, body.EntryIDs, device, s.roleFor(r, code))
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleClearQueue(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(r.PathValue("code"))
	if !s.requireHost(w, r, code) {
		return
	}
	device := s.ensureDevice(w, r)
	snap, err := s.mgr.ClearQueue(code, device, s.roleFor(r, code))
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleLockQueue(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(r.PathValue("code"))
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
	device := s.ensureDevice(w, r)
	snap, err := s.mgr.LockQueue(code, body.Locked, device, s.roleFor(r, code))
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleRotatePIN(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(r.PathValue("code"))
	if !s.requireHost(w, r, code) {
		return
	}
	pin, err := s.mgr.RotatePIN(code)
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"pin": pin})
}

func (s *Server) handleEndRoom(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(r.PathValue("code"))
	if !s.requireHost(w, r, code) {
		return
	}
	device := s.ensureDevice(w, r)
	if err := s.mgr.EndRoom(code, device, s.roleFor(r, code)); err != nil {
		s.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) requireHost(w http.ResponseWriter, r *http.Request, code string) bool {
	if !s.isHost(r, code) {
		s.writeError(w, domain.ErrHostRequired)
		return false
	}
	s.mgr.HostSeen(code)
	return true
}

func (s *Server) handleSPA(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	if p == "/" {
		s.writeIndex(w)
		return
	}
	if strings.HasPrefix(p, "/r/") && strings.Count(strings.Trim(p, "/"), "/") == 1 {
		s.ensureDevice(w, r)
		s.writeIndex(w)
		return
	}
	if strings.HasPrefix(p, "/tv/") && strings.Count(strings.Trim(p, "/"), "/") == 1 {
		token := strings.TrimPrefix(p, "/tv/")
		code, ok := s.mgr.RoomByHostToken(token)
		if !ok {
			http.NotFound(w, r)
			return
		}
		s.mgr.HostSeen(code)
		device := s.ensureDevice(w, r)
		s.setHostCookie(w, code)
		s.setTVRoomCookie(w, code)
		s.mgr.TouchDevice(device, true)
		s.writeIndex(w)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) writeIndex(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(s.index)
}

func (s *Server) handleFavicon(w http.ResponseWriter, r *http.Request) {
	if s.static == nil {
		http.NotFound(w, r)
		return
	}
	data, err := fs.ReadFile(s.static, "favicon.svg")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(data)
}

func (s *Server) handleAsset(w http.ResponseWriter, r *http.Request) {
	if s.static == nil {
		http.NotFound(w, r)
		return
	}
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	data, err := fs.ReadFile(s.static, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	switch path.Ext(name) {
	case ".js":
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".woff2":
		w.Header().Set("Content-Type", "font/woff2")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	_, _ = w.Write(data)
}

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
		appErr = &domain.AppError{Code: "internal_error", Message: "Something went wrong.", Status: 500}
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
