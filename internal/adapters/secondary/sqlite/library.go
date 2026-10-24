// Package sqlite persists karaoke songs discovered online or played in a room,
// so repeat searches never hit the upstream quota.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"

	_ "modernc.org/sqlite"
)

// MaxResults caps how many library matches a search returns.
const MaxResults = 20

var schema = []string{
	`CREATE TABLE IF NOT EXISTS songs (
		video_id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		channel TEXT NOT NULL DEFAULT '',
		thumbnail_url TEXT NOT NULL DEFAULT '',
		duration_seconds INTEGER NOT NULL DEFAULT 0,
		search_text TEXT NOT NULL DEFAULT '',
		first_seen_at TEXT NOT NULL,
		last_seen_at TEXT NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS songs_search_text ON songs (search_text)`,
}

const upsertSQL = `INSERT INTO songs (
		video_id, title, channel, thumbnail_url, duration_seconds,
		search_text, first_seen_at, last_seen_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT (video_id) DO UPDATE SET
		title = CASE WHEN excluded.title <> '' THEN excluded.title ELSE songs.title END,
		channel = CASE WHEN excluded.channel <> '' THEN excluded.channel ELSE songs.channel END,
		thumbnail_url = CASE WHEN excluded.thumbnail_url <> '' THEN excluded.thumbnail_url ELSE songs.thumbnail_url END,
		duration_seconds = CASE WHEN excluded.duration_seconds > 0 THEN excluded.duration_seconds ELSE songs.duration_seconds END,
		search_text = CASE WHEN excluded.search_text <> '' THEN excluded.search_text ELSE songs.search_text END,
		last_seen_at = excluded.last_seen_at`

// Library is a SQLite-backed song catalog.
type Library struct {
	db *sql.DB
}

// Open creates the database file and its schema when they are missing.
func Open(path string) (*Library, error) {
	if err := ensureDir(path); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := prepare(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Library{db: db}, nil
}

func ensureDir(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("sqlite: empty database path")
	}
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func prepare(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA synchronous=NORMAL",
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("sqlite: %s: %w", pragma, err)
		}
	}
	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("sqlite: schema: %w", err)
		}
	}
	return nil
}

// Close releases the database.
func (l *Library) Close() error { return l.db.Close() }

// Count returns how many songs the library holds.
func (l *Library) Count(ctx context.Context) (int, error) {
	var total int
	err := l.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM songs`).Scan(&total)
	return total, err
}

// Status reports the local backend state for the admin page.
func (l *Library) Status(ctx context.Context) ports.LibraryStatus {
	status := ports.LibraryStatus{Backend: "sqlite", Connected: l.db.PingContext(ctx) == nil}
	if total, err := l.Count(ctx); err == nil {
		status.Songs = total
	}
	return status
}

// Search returns songs whose title or channel contains every query word.
func (l *Library) Search(ctx context.Context, query string, limit int) ([]domain.SearchResult, error) {
	tokens := domain.SearchTokens(query)
	if len(tokens) == 0 {
		return nil, nil
	}
	if limit <= 0 || limit > MaxResults {
		limit = MaxResults
	}
	statement, args := searchStatement(tokens, limit)
	rows, err := l.db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.SearchResult, 0, limit)
	for rows.Next() {
		var song domain.SearchResult
		if err := rows.Scan(&song.VideoID, &song.Title, &song.Channel, &song.ThumbnailURL, &song.DurationSeconds); err != nil {
			return nil, err
		}
		out = append(out, song)
	}
	return out, rows.Err()
}

func searchStatement(tokens []string, limit int) (string, []any) {
	clauses := make([]string, 0, len(tokens))
	args := make([]any, 0, len(tokens)+2)
	for _, token := range tokens {
		clauses = append(clauses, `search_text LIKE ? ESCAPE '\'`)
		args = append(args, "%"+escapeLike(token)+"%")
	}
	args = append(args, "%"+escapeLike(strings.Join(tokens, " "))+"%", limit)
	statement := `SELECT video_id, title, channel, thumbnail_url, duration_seconds
		FROM songs WHERE ` + strings.Join(clauses, " AND ") + `
		ORDER BY CASE WHEN search_text LIKE ? ESCAPE '\' THEN 0 ELSE 1 END,
			length(title), title
		LIMIT ?`
	return statement, args
}

// Save inserts songs and refreshes known ones without creating duplicates.
func (l *Library) Save(ctx context.Context, songs []domain.SearchResult) error {
	now := time.Now().UTC()
	stored := make([]domain.LibrarySong, 0, len(songs))
	for _, song := range validSongs(songs) {
		stored = append(stored, domain.NewLibrarySong(song, "", now))
	}
	return l.saveSongs(ctx, stored)
}

// Cache is the best-effort counterpart of Save; local disk writes are fast.
func (l *Library) Cache(ctx context.Context, songs []domain.SearchResult) error {
	return l.Save(ctx, songs)
}

func (l *Library) saveSongs(ctx context.Context, songs []domain.LibrarySong) error {
	if len(songs) == 0 {
		return nil
	}
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	stmt, err := tx.PrepareContext(ctx, upsertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, song := range songs {
		if err := insertSong(ctx, stmt, song); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func insertSong(ctx context.Context, stmt *sql.Stmt, song domain.LibrarySong) error {
	first := stamp(song.FirstSeenAt)
	last := stamp(song.LastSeenAt)
	text := song.SearchText
	if text == "" {
		text = domain.SearchText(song.Title, song.Channel)
	}
	_, err := stmt.ExecContext(ctx, song.VideoID, song.Title, song.Channel,
		song.ThumbnailURL, song.DurationSeconds, text, first, last)
	return err
}

func stamp(t time.Time) string {
	if t.IsZero() {
		return time.Now().UTC().Format(time.RFC3339)
	}
	return t.UTC().Format(time.RFC3339)
}

func validSongs(songs []domain.SearchResult) []domain.SearchResult {
	out := make([]domain.SearchResult, 0, len(songs))
	for _, song := range songs {
		if !domain.ValidVideoID(song.VideoID) || strings.TrimSpace(song.Title) == "" {
			continue
		}
		out = append(out, song)
	}
	return out
}

func escapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "%", `\%`)
	return strings.ReplaceAll(value, "_", `\_`)
}
