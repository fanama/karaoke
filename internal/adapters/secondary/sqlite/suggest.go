package sqlite

import (
	"context"

	"fana-karaoke/internal/domain"
)

const randomSQL = `SELECT video_id, title, channel, thumbnail_url, duration_seconds
	FROM songs ORDER BY RANDOM() LIMIT ?`

const lookupSQL = `SELECT video_id, title, channel, thumbnail_url, duration_seconds
	FROM songs WHERE video_id = ?`

// Random returns up to limit songs picked at random by SQLite itself.
func (l *Library) Random(ctx context.Context, limit int) ([]domain.SearchResult, error) {
	if limit <= 0 {
		return []domain.SearchResult{}, nil
	}
	if limit > MaxResults {
		limit = MaxResults
	}
	rows, err := l.db.QueryContext(ctx, randomSQL, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.SearchResult, 0, limit)
	for rows.Next() {
		var song domain.SearchResult
		if err := rows.Scan(&song.VideoID, &song.Title, &song.Channel,
			&song.ThumbnailURL, &song.DurationSeconds); err != nil {
			return nil, err
		}
		out = append(out, song)
	}
	return out, rows.Err()
}

// Lookup returns the stored metadata of one song, if it is known.
func (l *Library) Lookup(ctx context.Context, videoID string) (domain.VideoMeta, bool) {
	if !domain.ValidVideoID(videoID) {
		return domain.VideoMeta{}, false
	}
	var meta domain.VideoMeta
	err := l.db.QueryRowContext(ctx, lookupSQL, videoID).Scan(&meta.VideoID,
		&meta.Title, &meta.Channel, &meta.ThumbnailURL, &meta.DurationSeconds)
	if err != nil {
		return domain.VideoMeta{}, false
	}
	return meta, true
}
