package usecase

import (
	"context"
	"io"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

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
