package d1

import (
	"math/rand/v2"

	"fana-karaoke/internal/domain"
)

// random samples the warm cache without reading a single D1 row.
func (c *cache) random(limit int) []domain.SearchResult {
	c.mu.RLock()
	songs := c.snapshotLocked()
	c.mu.RUnlock()
	if limit <= 0 || len(songs) == 0 {
		return []domain.SearchResult{}
	}
	if limit > len(songs) {
		limit = len(songs)
	}
	for i := 0; i < limit; i++ {
		j := i + rand.IntN(len(songs)-i)
		songs[i], songs[j] = songs[j], songs[i]
	}
	out := make([]domain.SearchResult, limit)
	for i := 0; i < limit; i++ {
		out[i] = songs[i].SearchResult
	}
	return out
}

// lookup answers from the warm cache so queueing stays offline.
func (c *cache) lookup(videoID string) (domain.VideoMeta, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	song, ok := c.songs[videoID]
	if !ok {
		return domain.VideoMeta{}, false
	}
	return domain.VideoMeta{
		VideoID:         song.VideoID,
		Title:           song.Title,
		Channel:         song.Channel,
		ThumbnailURL:    song.ThumbnailURL,
		DurationSeconds: song.DurationSeconds,
	}, true
}
