package d1

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

// MaxCacheSongs is the largest library the warm cache accepts.
const MaxCacheSongs = 10_000

// Config wires the D1 library to its Worker proxy.
type Config struct {
	BaseURL           string
	Token             string
	HTTPClient        *http.Client
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	CacheWriteTimeout time.Duration
	WarmupTimeout     time.Duration
	Now               func() time.Time
	Log               *slog.Logger
}

// Library stores songs in D1 through the Worker and searches a warm cache.
type Library struct {
	client        *client
	cache         *cache
	now           func() time.Time
	log           *slog.Logger
	readTimeout   time.Duration
	writeTimeout  time.Duration
	cacheTimeout  time.Duration
	warmupTimeout time.Duration
	configErr     string
	bg            chan []domain.LibrarySong
	closed        chan struct{}
	closeOnce     sync.Once
	warming       atomic.Bool
}

// New builds a D1 library. Missing configuration yields a cache-only library
// whose writes are rejected, rather than a disabled library.
func New(cfg Config) *Library {
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	log := cfg.Log
	if log == nil {
		log = slog.Default()
	}
	baseURL, configErr := validateBaseURL(cfg.BaseURL)
	if strings.TrimSpace(cfg.Token) == "" {
		configErr = codeConfiguration
	}
	l := &Library{
		client:        newClient(baseURL, cfg.Token, cfg.HTTPClient, log),
		cache:         newCache(),
		now:           now,
		log:           log,
		readTimeout:   duration(cfg.ReadTimeout, 10*time.Second),
		writeTimeout:  duration(cfg.WriteTimeout, 5*time.Second),
		cacheTimeout:  duration(cfg.CacheWriteTimeout, 5*time.Second),
		warmupTimeout: duration(cfg.WarmupTimeout, 10*time.Second),
		configErr:     configErr,
		bg:            make(chan []domain.LibrarySong, 1),
		closed:        make(chan struct{}),
	}
	if configErr != "" {
		l.cache.markError(configErr, true)
	}
	go l.runWriter()
	return l
}

func validateBaseURL(raw string) (string, string) {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if raw == "" {
		return "", codeConfiguration
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "", codeConfiguration
	}
	if u.Scheme == "https" {
		return raw, ""
	}
	if u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1") {
		return raw, ""
	}
	return "", codeConfiguration
}

func duration(value, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}
	return value
}

// Search answers from the warm cache without touching D1.
func (l *Library) Search(_ context.Context, query string, limit int) ([]domain.SearchResult, error) {
	return l.cache.search(query, limit), nil
}

// Random samples the warm cache without touching D1.
func (l *Library) Random(_ context.Context, limit int) ([]domain.SearchResult, error) {
	return l.cache.random(limit), nil
}

// Lookup answers from the warm cache so queueing stays offline.
func (l *Library) Lookup(_ context.Context, videoID string) (domain.VideoMeta, bool) {
	return l.cache.lookup(videoID)
}

// Count returns the number of cached songs.
func (l *Library) Count(context.Context) (int, error) { return l.cache.count(), nil }

// Status pings the Worker and reports the last known sync state.
func (l *Library) Status(ctx context.Context) ports.LibraryStatus {
	if l.configErr != "" {
		return l.cache.status(false)
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	err := l.client.health(ctx)
	if err != nil {
		l.cache.markError(errorCode(err), l.cache.isCacheOnly())
		return l.cache.status(false)
	}
	l.cache.markConnected()
	return l.cache.status(true)
}

// Close stops the background writer. In-flight writes may be lost.
func (l *Library) Close() error {
	l.closeOnce.Do(func() { close(l.closed) })
	return nil
}
