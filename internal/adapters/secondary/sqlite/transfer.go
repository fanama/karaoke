package sqlite

import (
	"context"
	"database/sql"
	"io"
	"os"
	"time"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

// MaxImportBytes is the largest library file the server accepts.
const MaxImportBytes = 128 << 20

// Export writes a portable copy of the library in the requested format.
func (l *Library) Export(ctx context.Context, format ports.LibraryFormat, w io.Writer) error {
	switch format {
	case ports.FormatSQLite:
		return l.exportSQLite(ctx, w)
	case ports.FormatSQL:
		return l.exportSQL(ctx, w)
	default:
		return domain.ErrInvalidInput
	}
}

func (l *Library) exportSQLite(ctx context.Context, w io.Writer) error {
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

// Import merges an uploaded library and returns the distinct songs read.
func (l *Library) Import(ctx context.Context, format ports.LibraryFormat, r io.Reader) (int, error) {
	switch format {
	case ports.FormatSQLite:
		return l.importSQLite(ctx, r)
	case ports.FormatSQL:
		return l.importSQL(ctx, r)
	default:
		return 0, domain.ErrInvalidInput
	}
}

func (l *Library) importSQLite(ctx context.Context, r io.Reader) (int, error) {
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
	songs, err := readLibrarySongs(ctx, name)
	if err != nil {
		return 0, domain.ErrLibraryInvalid
	}
	return l.storeImported(ctx, songs)
}

func (l *Library) storeImported(ctx context.Context, songs []domain.LibrarySong) (int, error) {
	songs, err := validateStored(songs)
	if err != nil {
		return 0, err
	}
	if err := l.saveSongs(ctx, songs); err != nil {
		return 0, err
	}
	return len(songs), nil
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

func readLibrarySongs(ctx context.Context, path string) ([]domain.LibrarySong, error) {
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		return nil, err
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	return queryLibrarySongs(ctx, db)
}

func parseStamp(value string) time.Time {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}
	return time.Time{}
}
