package sqlite

import (
	"context"
	"database/sql"
	"io"
	"os"

	"fana-karaoke/internal/domain"
)

// MaxImportBytes is the largest library file the server accepts.
const MaxImportBytes = 128 << 20

// Export writes a compact, self-contained copy of the database.
func (l *Library) Export(ctx context.Context, w io.Writer) error {
	name, cleanup, err := tempName("fana-library-*.sqlite")
	if err != nil {
		return err
	}
	defer cleanup()
	if _, err := l.db.ExecContext(ctx, "VACUUM INTO ?", name); err != nil {
		return err
	}
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(w, file)
	return err
}

// Import merges the songs of an uploaded library and returns how many it read.
func (l *Library) Import(ctx context.Context, r io.Reader) (int, error) {
	name, cleanup, err := tempName("fana-import-*.sqlite")
	if err != nil {
		return 0, err
	}
	defer cleanup()
	written, err := writeTemp(name, r)
	if err != nil {
		return 0, err
	}
	if written > MaxImportBytes {
		return 0, domain.ErrLibraryTooLarge
	}
	songs, err := readLibrary(ctx, name)
	if err != nil {
		return 0, domain.ErrLibraryInvalid
	}
	if err := l.Save(ctx, songs); err != nil {
		return 0, err
	}
	return len(validSongs(songs)), nil
}

func writeTemp(name string, r io.Reader) (int64, error) {
	file, err := os.Create(name)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return io.Copy(file, io.LimitReader(r, MaxImportBytes+1))
}

// tempName reserves an unused path; VACUUM INTO refuses to overwrite a file.
func tempName(pattern string) (string, func(), error) {
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", func() {}, err
	}
	name := file.Name()
	_ = file.Close()
	if err := os.Remove(name); err != nil {
		return "", func() {}, err
	}
	return name, func() { _ = os.Remove(name) }, nil
}

func readLibrary(ctx context.Context, path string) ([]domain.SearchResult, error) {
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		return nil, err
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	rows, err := db.QueryContext(ctx,
		`SELECT video_id, title, channel, thumbnail_url, duration_seconds FROM songs`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var songs []domain.SearchResult
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
