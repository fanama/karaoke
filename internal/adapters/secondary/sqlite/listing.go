package sqlite

import (
	"context"
	"strings"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

// List returns one page of stored songs ordered by title, with the total number
// of songs matching the optional filter.
func (l *Library) List(ctx context.Context, query ports.SongQuery) (ports.SongPage, error) {
	where, args := listWhere(query.Query)
	total, err := l.countMatching(ctx, where, args)
	if err != nil {
		return ports.SongPage{}, err
	}
	songs, err := l.listPage(ctx, query, where, args)
	if err != nil {
		return ports.SongPage{}, err
	}
	return ports.SongPage{
		Songs: songs, Total: total, Offset: query.Offset, Limit: query.Limit, Query: query.Query,
	}, nil
}

// listWhere builds the filter clause; search_text is already lowercased.
func listWhere(query string) (string, []any) {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return "", nil
	}
	return ` WHERE search_text LIKE ? ESCAPE '\'`, []any{"%" + escapeLike(query) + "%"}
}

func (l *Library) countMatching(ctx context.Context, where string, args []any) (int, error) {
	var total int
	err := l.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM songs`+where, args...).Scan(&total)
	return total, err
}

func (l *Library) listPage(ctx context.Context, query ports.SongQuery, where string, args []any) ([]domain.SearchResult, error) {
	statement := `SELECT video_id, title, channel, thumbnail_url, duration_seconds FROM songs` + where +
		` ORDER BY title COLLATE NOCASE, video_id LIMIT ? OFFSET ?`
	rows, err := l.db.QueryContext(ctx, statement, append(args, query.Limit, query.Offset)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	songs := make([]domain.SearchResult, 0, query.Limit)
	for rows.Next() {
		var song domain.SearchResult
		if err := rows.Scan(&song.VideoID, &song.Title, &song.Channel,
			&song.ThumbnailURL, &song.DurationSeconds); err != nil {
			return nil, err
		}
		songs = append(songs, song)
	}
	return songs, rows.Err()
}
