package sqlite

import (
	"context"
	"database/sql"
	"io"

	"fana-karaoke/internal/domain"
)

func (l *Library) exportSQL(ctx context.Context, w io.Writer) error {
	songs, err := queryLibrarySongs(ctx, l.db)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, domain.EncodeLibraryDump(songs))
	return err
}

func (l *Library) importSQL(ctx context.Context, r io.Reader) (int, error) {
	raw, err := io.ReadAll(io.LimitReader(r, MaxImportBytes+1))
	if err != nil {
		return 0, err
	}
	if len(raw) > MaxImportBytes {
		return 0, domain.ErrLibraryTooLarge
	}
	songs, err := domain.DecodeLibraryDump(string(raw))
	if err != nil {
		return 0, err
	}
	return l.storeImported(ctx, songs)
}

func queryLibrarySongs(ctx context.Context, db *sql.DB) ([]domain.LibrarySong, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT video_id, title, channel, thumbnail_url, duration_seconds,
			search_text, first_seen_at, last_seen_at FROM songs
			ORDER BY video_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var songs []domain.LibrarySong
	for rows.Next() {
		var song domain.LibrarySong
		var first, last string
		if err := rows.Scan(&song.VideoID, &song.Title, &song.Channel,
			&song.ThumbnailURL, &song.DurationSeconds, &song.SearchText,
			&first, &last); err != nil {
			return nil, err
		}
		song.FirstSeenAt = parseStamp(first)
		song.LastSeenAt = parseStamp(last)
		songs = append(songs, song)
	}
	return songs, rows.Err()
}

func validateStored(songs []domain.LibrarySong) ([]domain.LibrarySong, error) {
	return domain.ValidateLibrarySongs(songs)
}
