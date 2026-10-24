package d1

import (
	"context"
	"strings"
	"time"

	"fana-karaoke/internal/domain"
)

// Save waits for D1 before updating the cache. A failed save is rejected by
// the queue use case, so a song is never queued on a false durability promise.
func (l *Library) Save(ctx context.Context, songs []domain.SearchResult) error {
	stored := prepareSongs(songs, l.now().UTC())
	if len(stored) == 0 {
		if len(songs) == 0 {
			return nil
		}
		return domain.ErrLibraryInvalid
	}
	if l.configErr != "" {
		l.cache.markError(l.configErr, true)
		return domain.ErrLibraryUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, l.writeTimeout)
	defer cancel()
	if err := l.client.upsert(ctx, stored); err != nil {
		l.cache.markError(errorCode(err), true)
		l.log.Warn("d1 library save failed", "error", err)
		return domain.ErrLibraryUnavailable
	}
	now := l.now().UTC()
	l.cache.merge(stored)
	l.cache.markSuccess(now, l.cache.isCacheOnly())
	l.scheduleRewarm()
	return nil
}

// Cache updates memory immediately and hands the durable write to a
// single-flight worker. If the worker is busy, the batch is dropped.
func (l *Library) Cache(_ context.Context, songs []domain.SearchResult) error {
	stored := prepareSongs(songs, l.now().UTC())
	if len(stored) == 0 {
		return nil
	}
	l.cache.merge(stored)
	if l.configErr != "" {
		return nil
	}
	select {
	case l.bg <- stored:
	default:
		l.log.Debug("d1 library background write dropped", "songs", len(stored))
	}
	return nil
}

func prepareSongs(songs []domain.SearchResult, now time.Time) []domain.LibrarySong {
	out := make([]domain.LibrarySong, 0, len(songs))
	for _, song := range songs {
		if !domain.ValidVideoID(song.VideoID) || strings.TrimSpace(song.Title) == "" {
			continue
		}
		if song.DurationSeconds < 0 || song.DurationSeconds > 36000 {
			continue
		}
		out = append(out, domain.NewLibrarySong(song, "", now))
	}
	return out
}

func (l *Library) runWriter() {
	for {
		select {
		case batch := <-l.bg:
			l.writeBatch(batch)
		case <-l.closed:
			return
		}
	}
}

func (l *Library) writeBatch(batch []domain.LibrarySong) {
	ctx, cancel := context.WithTimeout(context.Background(), l.cacheTimeout)
	defer cancel()
	if err := l.client.upsert(ctx, batch); err != nil {
		l.cache.markError(errorCode(err), l.cache.isCacheOnly())
		l.log.Warn("d1 library background write failed", "error", err)
		return
	}
	l.cache.markSuccess(l.now().UTC(), l.cache.isCacheOnly())
	l.scheduleRewarm()
}

func (l *Library) scheduleRewarm() {
	if !l.cache.isCacheOnly() || !l.warming.CompareAndSwap(false, true) {
		return
	}
	go func() {
		defer l.warming.Store(false)
		_ = l.Warm(context.Background())
	}()
}
