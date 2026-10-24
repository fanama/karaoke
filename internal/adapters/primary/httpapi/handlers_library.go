package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"fana-karaoke/internal/domain"
)

const maxLibraryUpload = 64 << 20

func (s *Server) handleLibraryStats(w http.ResponseWriter, r *http.Request) {
	code := roomCode(r)
	if !s.requireHost(w, r, code) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	total, err := s.svc.LibraryStats(ctx)
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]int{"songs": total})
}

func (s *Server) handleLibraryExport(w http.ResponseWriter, r *http.Request) {
	code := roomCode(r)
	if !s.requireHost(w, r, code) {
		return
	}
	s.writeLibraryExport(w, r)
}

func (s *Server) handleLibraryImport(w http.ResponseWriter, r *http.Request) {
	code := roomCode(r)
	if !s.requireHost(w, r, code) {
		return
	}
	if ok, retry := s.limited("library:"+code, 5, time.Hour); !ok {
		s.writeError(w, rateLimited("Too many library imports.", retry))
		return
	}
	s.importLibrary(w, r)
}

// writeLibraryExport streams the SQLite library as a download.
func (s *Server) writeLibraryExport(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	name := fmt.Sprintf("fana-library-%s.sqlite", time.Now().UTC().Format("20060102-150405"))
	w.Header().Set("Content-Type", "application/vnd.sqlite3")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
	w.Header().Set("Cache-Control", "no-store")
	if err := s.svc.ExportLibrary(ctx, w); err != nil {
		s.log.Error("library export", "error", err)
	}
}

// importLibrary merges an uploaded SQLite file and reports the new total.
func (s *Server) importLibrary(w http.ResponseWriter, r *http.Request) {
	body := http.MaxBytesReader(w, r.Body, maxLibraryUpload)
	defer body.Close()
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	imported, err := s.svc.ImportLibrary(ctx, body)
	if err != nil {
		s.writeError(w, libraryError(err))
		return
	}
	total, err := s.svc.LibraryStats(ctx)
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]int{"imported": imported, "songs": total})
}

// libraryError turns an oversized upload into the matching public error.
func libraryError(err error) error {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return domain.ErrLibraryTooLarge
	}
	return err
}
