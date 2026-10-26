package d1

import (
	"context"
	"sort"
	"strings"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

// List returns one page of cached songs ordered by title. The warm cache mirrors
// D1 (warmup fails rather than exposing a partial library), so paging it keeps
// the admin list authoritative without a Worker round trip per page.
func (l *Library) List(_ context.Context, query ports.SongQuery) (ports.SongPage, error) {
	return l.cache.list(query), nil
}

// list filters the cache, sorts the matches by title and cuts the requested page.
func (c *cache) list(query ports.SongQuery) ports.SongPage {
	c.mu.RLock()
	songs := c.snapshotLocked()
	c.mu.RUnlock()

	matching := filterSongs(songs, query.Query)
	total := len(matching)
	start := min(query.Offset, total)
	end := min(start+max(query.Limit, 0), total)
	page := make([]domain.SearchResult, 0, end-start)
	for _, song := range matching[start:end] {
		page = append(page, song.SearchResult)
	}
	return ports.SongPage{
		Songs: page, Total: total, Offset: query.Offset, Limit: query.Limit, Query: query.Query,
	}
}

// filterSongs keeps the songs whose stored text contains the filter, ordered by
// title. An empty filter keeps every song.
func filterSongs(songs []domain.LibrarySong, filter string) []domain.LibrarySong {
	needle := strings.ToLower(strings.TrimSpace(filter))
	out := make([]domain.LibrarySong, 0, len(songs))
	for _, song := range songs {
		if needle == "" || strings.Contains(songText(song), needle) {
			out = append(out, song)
		}
	}
	sortByTitle(out)
	return out
}

// songText falls back to the title and channel when no search text was stored.
func songText(song domain.LibrarySong) string {
	if song.SearchText != "" {
		return song.SearchText
	}
	return domain.SearchText(song.Title, song.Channel)
}

// sortByTitle orders songs alphabetically, breaking ties on the video id.
func sortByTitle(songs []domain.LibrarySong) {
	sort.SliceStable(songs, func(i, j int) bool {
		left, right := strings.ToLower(songs[i].Title), strings.ToLower(songs[j].Title)
		if left != right {
			return left < right
		}
		return songs[i].VideoID < songs[j].VideoID
	})
}
