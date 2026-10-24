package domain

import (
	"strconv"
	"strings"
	"time"
)

// DecodeLibraryDump parses only the format produced by EncodeLibraryDump.
func DecodeLibraryDump(raw string) ([]LibrarySong, error) {
	if !strings.HasPrefix(strings.TrimSpace(raw), LibraryDumpHeader) {
		return nil, ErrLibraryInvalid
	}
	statements, err := splitDumpStatements(raw)
	if err != nil {
		return nil, err
	}
	songs := make([]LibrarySong, 0, len(statements))
	for _, statement := range statements {
		song, ok, err := parseDumpStatement(statement)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		songs = append(songs, song)
		if len(songs) > MaxImportSongs {
			return nil, ErrLibraryTooLarge
		}
	}
	if len(songs) == 0 {
		return nil, ErrLibraryInvalid
	}
	return songs, nil
}

func parseDumpStatement(statement string) (LibrarySong, bool, error) {
	statement = strings.TrimSpace(statement)
	if statement == "" || strings.EqualFold(statement, "BEGIN") || strings.EqualFold(statement, "COMMIT") {
		return LibrarySong{}, false, nil
	}
	if !strings.HasPrefix(statement, "INSERT INTO songs (") {
		return LibrarySong{}, false, ErrLibraryInvalid
	}
	rest := statement[len("INSERT INTO songs ("):]
	closeIdx := strings.Index(rest, ") VALUES (")
	if closeIdx < 0 {
		return LibrarySong{}, false, ErrLibraryInvalid
	}
	columns := strings.Join(strings.Fields(rest[:closeIdx]), ",")
	if columns != dumpColumns {
		return LibrarySong{}, false, ErrLibraryInvalid
	}
	values := strings.TrimSpace(rest[closeIdx+len(") VALUES ("):])
	if !strings.HasSuffix(values, ")") {
		return LibrarySong{}, false, ErrLibraryInvalid
	}
	parts, err := splitDumpValues(strings.TrimSuffix(values, ")"))
	if err != nil {
		return LibrarySong{}, false, err
	}
	song, err := songFromDumpValues(parts)
	return song, true, err
}

func songFromDumpValues(parts []string) (LibrarySong, error) {
	fields, err := parseDumpFields(parts)
	if err != nil {
		return LibrarySong{}, err
	}
	return LibrarySong{
		SearchResult: SearchResult{
			VideoID: fields.videoID, Title: fields.title, Channel: fields.channel,
			ThumbnailURL: fields.thumbnail, DurationSeconds: fields.duration,
		},
		SearchText: fields.text, FirstSeenAt: fields.first, LastSeenAt: fields.last,
	}, nil
}

type dumpFields struct {
	videoID, title, channel, thumbnail, text string
	duration                                 int
	first, last                              time.Time
}

func parseDumpFields(parts []string) (dumpFields, error) {
	if len(parts) != 8 {
		return dumpFields{}, ErrLibraryInvalid
	}
	videoID, err := parseDumpString(parts[0])
	if err != nil || !ValidVideoID(videoID) {
		return dumpFields{}, ErrLibraryInvalid
	}
	title, err := parseDumpString(parts[1])
	if err != nil || strings.TrimSpace(title) == "" {
		return dumpFields{}, ErrLibraryInvalid
	}
	channel, err := parseDumpString(parts[2])
	if err != nil {
		return dumpFields{}, ErrLibraryInvalid
	}
	thumbnail, err := parseDumpString(parts[3])
	if err != nil {
		return dumpFields{}, ErrLibraryInvalid
	}
	duration, err := strconv.Atoi(strings.TrimSpace(parts[4]))
	if err != nil || duration < 0 || duration > 36000 {
		return dumpFields{}, ErrLibraryInvalid
	}
	text, err := parseDumpString(parts[5])
	if err != nil {
		return dumpFields{}, ErrLibraryInvalid
	}
	first, err := parseDumpTime(parts[6])
	if err != nil {
		return dumpFields{}, ErrLibraryInvalid
	}
	last, err := parseDumpTime(parts[7])
	if err != nil {
		return dumpFields{}, ErrLibraryInvalid
	}
	return dumpFields{videoID, strings.TrimSpace(title), channel, thumbnail, text, duration, first, last}, nil
}

func parseDumpString(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) < 2 || raw[0] != '\'' || raw[len(raw)-1] != '\'' {
		return "", ErrLibraryInvalid
	}
	body := raw[1 : len(raw)-1]
	var b strings.Builder
	for i := 0; i < len(body); i++ {
		if body[i] != '\'' {
			b.WriteByte(body[i])
			continue
		}
		if i+1 >= len(body) || body[i+1] != '\'' {
			return "", ErrLibraryInvalid
		}
		b.WriteByte('\'')
		i++
	}
	return b.String(), nil
}

func parseDumpTime(raw string) (time.Time, error) {
	value, err := parseDumpString(raw)
	if err != nil {
		return time.Time{}, err
	}
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, value)
}

func splitDumpValues(raw string) ([]string, error) {
	parts := make([]string, 0, 8)
	start := 0
	inQuote := false
	for i := 0; i < len(raw); i++ {
		switch raw[i] {
		case '\'':
			if inQuote && i+1 < len(raw) && raw[i+1] == '\'' {
				i++
				continue
			}
			inQuote = !inQuote
		case ',':
			if !inQuote {
				parts = append(parts, strings.TrimSpace(raw[start:i]))
				start = i + 1
			}
		}
	}
	if inQuote {
		return nil, ErrLibraryInvalid
	}
	parts = append(parts, strings.TrimSpace(raw[start:]))
	return parts, nil
}

func splitDumpStatements(raw string) ([]string, error) {
	var statements []string
	var b strings.Builder
	inQuote := false
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if !inQuote && c == '-' && i+1 < len(raw) && raw[i+1] == '-' {
			for i < len(raw) && raw[i] != '\n' {
				i++
			}
			b.WriteByte('\n')
			continue
		}
		if c == '\'' {
			b.WriteByte(c)
			if inQuote && i+1 < len(raw) && raw[i+1] == '\'' {
				b.WriteByte(raw[i+1])
				i++
				continue
			}
			inQuote = !inQuote
			continue
		}
		if c == ';' && !inQuote {
			statements = append(statements, b.String())
			b.Reset()
			continue
		}
		b.WriteByte(c)
	}
	if inQuote {
		return nil, ErrLibraryInvalid
	}
	if tail := strings.TrimSpace(b.String()); tail != "" {
		return nil, ErrLibraryInvalid
	}
	return statements, nil
}
