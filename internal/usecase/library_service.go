package usecase

import (
	"context"
	"io"

	"fana-karaoke/internal/domain"
)

// LibraryStats returns how many songs the local library holds.
func (s *CatalogService) LibraryStats(ctx context.Context) (int, error) {
	if s.Library == nil {
		return 0, domain.ErrLibraryDisabled
	}
	return s.Library.Count(ctx)
}

// ExportLibrary streams a portable copy of the local library.
func (s *CatalogService) ExportLibrary(ctx context.Context, w io.Writer) error {
	if s.Library == nil {
		return domain.ErrLibraryDisabled
	}
	return s.Library.Export(ctx, w)
}

// ImportLibrary merges an uploaded library into the local one.
func (s *CatalogService) ImportLibrary(ctx context.Context, r io.Reader) (int, error) {
	if s.Library == nil {
		return 0, domain.ErrLibraryDisabled
	}
	return s.Library.Import(ctx, r)
}
