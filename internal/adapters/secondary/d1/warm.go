package d1

import (
	"context"
	"errors"
	"time"

	"fana-karaoke/internal/domain"
)

const warmPageSize = 500

// Warm loads every song from D1. Any failure leaves the cache empty and marks
// the library cache-only; a partial page set is never exposed.
func (l *Library) Warm(ctx context.Context) error {
	if l.configErr != "" {
		l.cache.markError(l.configErr, true)
		return errors.New("d1: library configuration is incomplete")
	}
	ctx, cancel := context.WithTimeout(ctx, l.warmupTimeout)
	defer cancel()
	songs, err := l.loadAll(ctx)
	if err != nil {
		l.cache.markError(errorCode(err), true)
		l.log.Warn("d1 library warm failed", "error", err)
		return err
	}
	l.cache.replace(songs)
	l.cache.markSuccess(l.now().UTC(), false)
	l.log.Info("d1 library warm loaded", "songs", len(songs))
	return nil
}

func (l *Library) loadAll(ctx context.Context) ([]domain.LibrarySong, error) {
	offset := 0
	all := make([]domain.LibrarySong, 0, warmPageSize)
	for pages := 0; pages < 40; pages++ {
		page, err := l.client.songs(ctx, offset, warmPageSize)
		if err != nil {
			return nil, err
		}
		converted, err := storedSongs(page.Songs, l.now().UTC())
		if err != nil {
			return nil, err
		}
		all = append(all, converted...)
		if len(all) > MaxCacheSongs {
			return nil, &apiError{code: codeOversize, status: 0, message: "too many songs"}
		}
		if page.NextOffset == nil {
			return all, nil
		}
		if *page.NextOffset <= offset {
			return nil, &apiError{code: codeInvalidResponse, status: 0, message: "pagination did not advance"}
		}
		offset = *page.NextOffset
	}
	return nil, &apiError{code: codeOversize, status: 0, message: "too many pages"}
}

func storedSongs(songs []apiSong, now time.Time) ([]domain.LibrarySong, error) {
	out := make([]domain.LibrarySong, 0, len(songs))
	for _, song := range songs {
		stored, err := song.stored(now)
		if err != nil {
			return nil, err
		}
		out = append(out, stored)
	}
	return out, nil
}

func (s apiSong) stored(now time.Time) (domain.LibrarySong, error) {
	if !domain.ValidVideoID(s.VideoID) || s.Title == "" {
		return domain.LibrarySong{}, invalidSong()
	}
	if s.DurationSeconds < 0 || s.DurationSeconds > 36000 {
		return domain.LibrarySong{}, invalidSong()
	}
	first, err := parseAPITime(s.FirstSeenAt, now)
	if err != nil {
		return domain.LibrarySong{}, err
	}
	last, err := parseAPITime(s.LastSeenAt, now)
	if err != nil {
		return domain.LibrarySong{}, err
	}
	text := s.SearchText
	if text == "" {
		text = domain.SearchText(s.Title, s.Channel)
	}
	return domain.LibrarySong{
		SearchResult: domain.SearchResult{
			VideoID: s.VideoID, Title: s.Title, Channel: s.Channel,
			ThumbnailURL: s.ThumbnailURL, DurationSeconds: s.DurationSeconds,
		},
		SearchText: text, FirstSeenAt: first, LastSeenAt: last,
	}, nil
}

func parseAPITime(value string, fallback time.Time) (time.Time, error) {
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, invalidSong()
	}
	return parsed, nil
}

func invalidSong() error {
	return &apiError{code: codeInvalidResponse, status: 0, message: "invalid song in response"}
}
