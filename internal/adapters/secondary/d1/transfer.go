package d1

import (
	"context"
	"io"
	"os"
	"time"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

const (
	maxImportBytes = 64 << 20
	importTimeout  = 30 * time.Second
)

// Export writes the warm cache without reading D1 again.
func (l *Library) Export(_ context.Context, format ports.LibraryFormat, w io.Writer) error {
	songs := l.cache.snapshot()
	switch format {
	case ports.FormatSQL:
		_, err := io.WriteString(w, domain.EncodeLibraryDump(songs))
		return err
	case ports.FormatSQLite:
		return writeSQLiteFile(w, songs)
	default:
		return domain.ErrInvalidInput
	}
}

// Import validates the whole file, sends one atomic batch to D1, and only then
// updates the cache.
func (l *Library) Import(ctx context.Context, format ports.LibraryFormat, r io.Reader) (int, error) {
	switch format {
	case ports.FormatSQL:
		return l.importSQL(ctx, r)
	case ports.FormatSQLite:
		return l.importSQLite(ctx, r)
	default:
		return 0, domain.ErrInvalidInput
	}
}

func (l *Library) importSQL(ctx context.Context, r io.Reader) (int, error) {
	raw, err := readLimited(r)
	if err != nil {
		return 0, err
	}
	songs, err := domain.DecodeLibraryDump(string(raw))
	if err != nil {
		return 0, err
	}
	return l.storeImported(ctx, songs)
}

func (l *Library) importSQLite(ctx context.Context, r io.Reader) (int, error) {
	name, cleanup, err := tempFile()
	if err != nil {
		return 0, err
	}
	defer cleanup()
	if err := writeLimited(name, r); err != nil {
		return 0, err
	}
	songs, err := readSQLiteSongs(ctx, name)
	if err != nil {
		return 0, domain.ErrLibraryInvalid
	}
	return l.storeImported(ctx, songs)
}

func (l *Library) storeImported(ctx context.Context, songs []domain.LibrarySong) (int, error) {
	songs, err := domain.ValidateLibrarySongs(songs)
	if err != nil {
		return 0, err
	}
	if l.configErr != "" {
		l.cache.markError(l.configErr, true)
		return 0, domain.ErrLibraryUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, importTimeout)
	defer cancel()
	if err := l.client.importSongs(ctx, songs); err != nil {
		l.cache.markError(errorCode(err), true)
		l.log.Warn("d1 library import failed", "error", err)
		return 0, domain.ErrLibraryUnavailable
	}
	l.cache.merge(songs)
	l.cache.markSuccess(l.now().UTC(), l.cache.isCacheOnly())
	l.scheduleRewarm()
	return len(songs), nil
}

func readLimited(r io.Reader) ([]byte, error) {
	raw, err := io.ReadAll(io.LimitReader(r, maxImportBytes+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > maxImportBytes {
		return nil, domain.ErrLibraryTooLarge
	}
	return raw, nil
}

func writeLimited(name string, r io.Reader) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()
	written, err := io.Copy(file, io.LimitReader(r, maxImportBytes+1))
	if err != nil {
		return err
	}
	if written > maxImportBytes {
		return domain.ErrLibraryTooLarge
	}
	return nil
}
