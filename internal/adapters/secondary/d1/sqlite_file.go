package d1

import (
	"context"
	"database/sql"
	"io"
	"os"
	"time"

	"fana-karaoke/internal/domain"

	_ "modernc.org/sqlite"
)

const sqliteSchema = `CREATE TABLE IF NOT EXISTS songs (
	video_id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	channel TEXT NOT NULL DEFAULT '',
	thumbnail_url TEXT NOT NULL DEFAULT '',
	duration_seconds INTEGER NOT NULL DEFAULT 0,
	search_text TEXT NOT NULL DEFAULT '',
	first_seen_at TEXT NOT NULL,
	last_seen_at TEXT NOT NULL
)`

const sqliteIndex = `CREATE INDEX IF NOT EXISTS songs_search_text ON songs (search_text)`

const sqliteInsert = `INSERT INTO songs (
	video_id, title, channel, thumbnail_url, duration_seconds,
	search_text, first_seen_at, last_seen_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

func writeSQLiteFile(w io.Writer, songs []domain.LibrarySong) error {
	name, cleanup, err := tempFile()
	if err != nil {
		return err
	}
	defer cleanup()
	if err := buildSQLiteFile(name, songs); err != nil {
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

func buildSQLiteFile(name string, songs []domain.LibrarySong) error {
	db, err := sql.Open("sqlite", name)
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err := db.Exec(sqliteSchema); err != nil {
		return err
	}
	if _, err := db.Exec(sqliteIndex); err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	stmt, err := tx.Prepare(sqliteInsert)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, song := range songs {
		if err := writeSQLiteSong(stmt, song); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func writeSQLiteSong(stmt *sql.Stmt, song domain.LibrarySong) error {
	text := song.SearchText
	if text == "" {
		text = domain.SearchText(song.Title, song.Channel)
	}
	_, err := stmt.Exec(song.VideoID, song.Title, song.Channel,
		song.ThumbnailURL, song.DurationSeconds, text,
		stamp(song.FirstSeenAt), stamp(song.LastSeenAt))
	return err
}

func readSQLiteSongs(ctx context.Context, name string) ([]domain.LibrarySong, error) {
	db, err := sql.Open("sqlite", "file:"+name+"?mode=ro")
	if err != nil {
		return nil, err
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	rows, err := db.QueryContext(ctx,
		`SELECT video_id, title, channel, thumbnail_url, duration_seconds,
			search_text, first_seen_at, last_seen_at FROM songs`)
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
		song.FirstSeenAt = parseFileTime(first)
		song.LastSeenAt = parseFileTime(last)
		songs = append(songs, song)
	}
	return songs, rows.Err()
}

func parseFileTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func tempFile() (string, func(), error) {
	file, err := os.CreateTemp("", "fana-d1-library-*.sqlite")
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
