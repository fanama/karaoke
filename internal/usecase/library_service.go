package usecase

import (
	"context"
	"io"
	"strings"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

// maxSongFilterLength caps the filter text accepted from a caller.
const maxSongFilterLength = 120

// LibrarySongs returns one page of the stored library for the admin page,
// ordered by title, with the total number of songs matching the filter.
func (s *CatalogService) LibrarySongs(ctx context.Context, query ports.SongQuery) (ports.SongPage, error) {
	if s.Library == nil {
		return ports.SongPage{}, domain.ErrLibraryDisabled
	}
	page, err := s.Library.List(ctx, normaliseSongQuery(query))
	if err != nil {
		return ports.SongPage{}, err
	}
	if page.Songs == nil {
		page.Songs = []domain.SearchResult{}
	}
	return page, nil
}

// normaliseSongQuery clamps pagination and trims the filter.
func normaliseSongQuery(query ports.SongQuery) ports.SongQuery {
	if query.Offset < 0 {
		query.Offset = 0
	}
	if query.Limit <= 0 {
		query.Limit = ports.DefaultSongPageSize
	}
	if query.Limit > ports.MaxSongPageSize {
		query.Limit = ports.MaxSongPageSize
	}
	query.Query = strings.TrimSpace(query.Query)
	if len(query.Query) > maxSongFilterLength {
		query.Query = query.Query[:maxSongFilterLength]
	}
	return query
}

// LibraryStats returns how many songs the local library holds.
func (s *CatalogService) LibraryStats(ctx context.Context) (int, error) {
	if s.Library == nil {
		return 0, domain.ErrLibraryDisabled
	}
	return s.Library.Count(ctx)
}

// LibraryInfo returns the storage backend state for the admin page.
func (s *CatalogService) LibraryInfo(ctx context.Context) (ports.LibraryStatus, error) {
	if s.Library == nil {
		return ports.LibraryStatus{Backend: "none"}, domain.ErrLibraryDisabled
	}
	return s.Library.Status(ctx), nil
}

// ExportLibrary streams a portable copy of the local library.
func (s *CatalogService) ExportLibrary(ctx context.Context, format ports.LibraryFormat, w io.Writer) error {
	if s.Library == nil {
		return domain.ErrLibraryDisabled
	}
	if !validLibraryFormat(format) {
		return domain.ErrInvalidInput
	}
	return s.Library.Export(ctx, format, w)
}

// ImportLibrary merges an uploaded library into the local one.
func (s *CatalogService) ImportLibrary(ctx context.Context, format ports.LibraryFormat, r io.Reader) (int, error) {
	if s.Library == nil {
		return 0, domain.ErrLibraryDisabled
	}
	if !validLibraryFormat(format) {
		return 0, domain.ErrInvalidInput
	}
	return s.Library.Import(ctx, format, r)
}

func validLibraryFormat(format ports.LibraryFormat) bool {
	return format == ports.FormatSQLite || format == ports.FormatSQL
}
