package invidious

import (
	"sync"
	"time"

	"fana-karaoke/internal/domain"
)

type cachedQuery struct {
	results []domain.SearchResult
	at      time.Time
}

type cachedVideo struct {
	meta domain.VideoMeta
	at   time.Time
}

// Cache stores search results and video metadata with fresh/stale windows.
type Cache struct {
	mu       sync.Mutex
	queryTTL time.Duration
	staleTTL time.Duration
	videoTTL time.Duration
	queries  map[string]cachedQuery
	videos   map[string]cachedVideo
	blocked  map[string]time.Time
	now      func() time.Time
}

// NewCache builds an empty cache.
func NewCache(queryTTL, staleTTL, videoTTL time.Duration, now func() time.Time) *Cache {
	return &Cache{
		queryTTL: queryTTL,
		staleTTL: staleTTL,
		videoTTL: videoTTL,
		queries:  map[string]cachedQuery{},
		videos:   map[string]cachedVideo{},
		blocked:  map[string]time.Time{},
		now:      now,
	}
}

// Query returns a cached result set inside the given freshness window.
func (c *Cache) Query(key string, fresh bool) ([]domain.SearchResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.queries[key]
	if !ok {
		return nil, false
	}
	window := c.staleTTL
	if fresh {
		window = c.queryTTL
	}
	if c.now().Sub(entry.at) > window {
		return nil, false
	}
	out := make([]domain.SearchResult, len(entry.results))
	copy(out, entry.results)
	return out, true
}

// StoreQuery remembers a result set.
func (c *Cache) StoreQuery(key string, results []domain.SearchResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.queries) >= 1000 {
		for k := range c.queries {
			delete(c.queries, k)
			break
		}
	}
	out := make([]domain.SearchResult, len(results))
	copy(out, results)
	c.queries[key] = cachedQuery{results: out, at: c.now()}
}

// Video returns cached metadata inside the requested freshness window.
func (c *Cache) Video(id string, fresh bool) (domain.VideoMeta, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.videos[id]
	if !ok {
		return domain.VideoMeta{}, false
	}
	window := c.staleTTL
	if fresh {
		window = c.videoTTL
	}
	if c.now().Sub(entry.at) > window {
		return domain.VideoMeta{}, false
	}
	return entry.meta, true
}

// StoreVideo remembers metadata.
func (c *Cache) StoreVideo(id string, meta domain.VideoMeta) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.videos) >= 2000 {
		for k := range c.videos {
			delete(c.videos, k)
			break
		}
	}
	c.videos[id] = cachedVideo{meta: meta, at: c.now()}
}

// MarkBlocked remembers that a video cannot be embedded.
func (c *Cache) MarkBlocked(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.blocked) >= 2000 {
		for k := range c.blocked {
			delete(c.blocked, k)
			break
		}
	}
	c.blocked[id] = c.now().Add(c.videoTTL)
}

// Blocked reports whether a video is known to forbid embedding.
func (c *Cache) Blocked(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	until, ok := c.blocked[id]
	if !ok {
		return false
	}
	if !c.now().Before(until) {
		delete(c.blocked, id)
		return false
	}
	return true
}
