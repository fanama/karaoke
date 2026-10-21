package httpapi

import (
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"fana-karaoke/internal/ports"
)

// Config holds transport-level settings.
type Config struct {
	Dev        bool
	BaseURL    string
	TrustProxy bool
	Version    string
}

// Server is the HTTP + WebSocket driving adapter.
type Server struct {
	svc        ports.KaraokeService
	signer     ports.SessionSigner
	limiter    ports.RateLimiter
	qr         ports.QRGenerator
	hub        *Hub
	static     fs.FS
	index      []byte
	baseURL    string
	dev        bool
	trustProxy bool
	log        *slog.Logger
}

// New builds the adapter. The hub is passed in so the composition root can
// subscribe it to the event bus without any setters.
func New(svc ports.KaraokeService, hub *Hub, signer ports.SessionSigner, limiter ports.RateLimiter, qr ports.QRGenerator, static fs.FS, cfg Config, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	s := &Server{
		svc:        svc,
		signer:     signer,
		limiter:    limiter,
		qr:         qr,
		hub:        hub,
		static:     static,
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		dev:        cfg.Dev,
		trustProxy: cfg.TrustProxy,
		log:        log,
	}
	s.loadIndex(static)
	return s
}

func (s *Server) loadIndex(static fs.FS) {
	if static == nil {
		s.index = []byte("<!doctype html><title>Fana Karaoke</title><p>The frontend has not been built.</p>")
		return
	}
	if b, err := fs.ReadFile(static, "index.html"); err == nil {
		s.index = b
	}
	if len(s.index) == 0 {
		s.index = []byte("<!doctype html><title>Fana Karaoke</title><p>The frontend has not been built.</p>")
	}
}

// Handler mounts every route behind the security middleware.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	s.routes(mux)
	return s.middleware(mux)
}

func (s *Server) writeIndex(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(s.index)
}
