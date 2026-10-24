package domain

import (
	"strconv"
	"strings"
	"time"
)

// LibraryDumpHeader identifies the only SQL format the admin importer accepts.
const LibraryDumpHeader = "-- fana-karaoke-library v1"

// MaxImportSongs caps a single library import.
const MaxImportSongs = 2000

const dumpColumns = "video_id,title,channel,thumbnail_url,duration_seconds,search_text,first_seen_at,last_seen_at"

// LibrarySong is a stored song plus its persistence timestamps.
type LibrarySong struct {
	SearchResult
	SearchText  string
	FirstSeenAt time.Time
	LastSeenAt  time.Time
}

// Result returns the search-facing fields of a stored song.
func (s LibrarySong) Result() SearchResult { return s.SearchResult }

// NewLibrarySong builds a stored song from a search result.
func NewLibrarySong(result SearchResult, text string, now time.Time) LibrarySong {
	if text == "" {
		text = SearchText(result.Title, result.Channel)
	}
	return LibrarySong{
		SearchResult: result,
		SearchText:   text,
		FirstSeenAt:  now,
		LastSeenAt:   now,
	}
}

// ValidateLibrarySongs rejects invalid rows, caps the batch and deduplicates
// by video id. The last occurrence of a video id wins.
func ValidateLibrarySongs(songs []LibrarySong) ([]LibrarySong, error) {
	if len(songs) > MaxImportSongs {
		return nil, ErrLibraryTooLarge
	}
	out := make([]LibrarySong, 0, len(songs))
	seen := make(map[string]int, len(songs))
	for _, song := range songs {
		if err := validateLibrarySong(song); err != nil {
			return nil, err
		}
		song.Title = strings.TrimSpace(song.Title)
		if song.SearchText == "" {
			song.SearchText = SearchText(song.Title, song.Channel)
		}
		if index, ok := seen[song.VideoID]; ok {
			out[index] = song
			continue
		}
		seen[song.VideoID] = len(out)
		out = append(out, song)
	}
	return out, nil
}

func validateLibrarySong(song LibrarySong) error {
	if !ValidVideoID(song.VideoID) || strings.TrimSpace(song.Title) == "" {
		return ErrLibraryInvalid
	}
	if song.DurationSeconds < 0 || song.DurationSeconds > 36000 {
		return ErrLibraryInvalid
	}
	return nil
}

// EncodeLibraryDump renders the app's portable SQL format.
func EncodeLibraryDump(songs []LibrarySong) string {
	var b strings.Builder
	b.WriteString(LibraryDumpHeader)
	b.WriteString("\nBEGIN;\n")
	for _, song := range songs {
		writeDumpRow(&b, song)
	}
	b.WriteString("COMMIT;\n")
	return b.String()
}

func writeDumpRow(b *strings.Builder, song LibrarySong) {
	text := song.SearchText
	if text == "" {
		text = SearchText(song.Title, song.Channel)
	}
	b.WriteString("INSERT INTO songs (" + dumpColumns + ") VALUES (")
	b.WriteString(quoteDump(song.VideoID) + ", ")
	b.WriteString(quoteDump(song.Title) + ", ")
	b.WriteString(quoteDump(song.Channel) + ", ")
	b.WriteString(quoteDump(song.ThumbnailURL) + ", ")
	b.WriteString(strconv.Itoa(song.DurationSeconds) + ", ")
	b.WriteString(quoteDump(text) + ", ")
	b.WriteString(quoteDump(formatDumpTime(song.FirstSeenAt)) + ", ")
	b.WriteString(quoteDump(formatDumpTime(song.LastSeenAt)) + ");\n")
}

func quoteDump(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func formatDumpTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
