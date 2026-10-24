package d1

import (
	"sort"
	"sync"
	"time"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

type cache struct {
	mu        sync.RWMutex
	songs     map[string]domain.LibrarySong
	connected bool
	cacheOnly bool
	lastSync  time.Time
	lastErr   string
}

func newCache() *cache {
	return &cache{songs: make(map[string]domain.LibrarySong), cacheOnly: true}
}

func (c *cache) replace(songs []domain.LibrarySong) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.songs = make(map[string]domain.LibrarySong, len(songs))
	for _, song := range songs {
		c.songs[song.VideoID] = song
	}
}

func (c *cache) merge(songs []domain.LibrarySong) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, song := range songs {
		if old, ok := c.songs[song.VideoID]; ok {
			song.FirstSeenAt = old.FirstSeenAt
		}
		c.songs[song.VideoID] = song
	}
}

func (c *cache) search(query string, limit int) []domain.SearchResult {
	c.mu.RLock()
	songs := c.snapshotLocked()
	c.mu.RUnlock()
	results := make([]domain.SearchResult, 0, len(songs))
	for _, song := range songs {
		results = append(results, song.SearchResult)
	}
	return domain.RankSearch(results, query, limit)
}

func (c *cache) count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.songs)
}

func (c *cache) snapshot() []domain.LibrarySong {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.snapshotLocked()
}

func (c *cache) snapshotLocked() []domain.LibrarySong {
	out := make([]domain.LibrarySong, 0, len(c.songs))
	for _, song := range c.songs {
		out = append(out, song)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].VideoID < out[j].VideoID })
	return out
}

func (c *cache) markSuccess(now time.Time, cacheOnly bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connected = true
	c.cacheOnly = cacheOnly
	c.lastSync = now
	c.lastErr = ""
}

func (c *cache) markError(code string, cacheOnly bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connected = false
	c.cacheOnly = cacheOnly
	c.lastErr = code
}

func (c *cache) markConnected() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connected = true
	c.lastErr = ""
}

func (c *cache) isCacheOnly() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cacheOnly
}

func (c *cache) status(connected bool) ports.LibraryStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return ports.LibraryStatus{
		Backend:    "d1",
		Connected:  connected,
		CacheOnly:  c.cacheOnly,
		Songs:      len(c.songs),
		LastSyncAt: c.lastSync,
		LastError:  c.lastErr,
	}
}

func (c *cache) errorCode() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastErr
}
