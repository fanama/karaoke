package domain

import (
	"strings"
	"testing"
	"time"
)

func TestSearchFoldsAccentsAndRequiresEveryWord(t *testing.T) {
	songs := []SearchResult{
		{VideoID: "aaaaaaaaaaa", Title: "Été – Toto (karaoké version)", Channel: "Studio"},
		{VideoID: "bbbbbbbbbbb", Title: "Bohemian Rhapsody Karaoke", Channel: "Queen"},
	}
	cases := map[string]int{
		"ete": 1, "ETE KARAOKE": 1, "ete queen": 0, "été": 1,
		"bohemian karaok": 1, "toto": 1, "unknown": 0,
	}
	for query, want := range cases {
		if got := RankSearch(songs, query, 20); len(got) != want {
			t.Fatalf("%q: got %d, want %d", query, len(got), want)
		}
	}
}

func TestSearchTreatsWildcardsLiterally(t *testing.T) {
	songs := []SearchResult{
		{VideoID: "aaaaaaaaaaa", Title: "100% Karaoke", Channel: "Studio"},
		{VideoID: "bbbbbbbbbbb", Title: "Hello Karaoke", Channel: "Studio"},
	}
	found := RankSearch(songs, "100%", 20)
	if len(found) != 1 || found[0].VideoID != "aaaaaaaaaaa" {
		t.Fatalf("wildcard search = %+v", found)
	}
}

func TestSearchRanksExactPhraseBeforeShorterTitle(t *testing.T) {
	songs := []SearchResult{
		{VideoID: "aaaaaaaaaaa", Title: "Karaoke", Channel: "Studio"},
		{VideoID: "bbbbbbbbbbb", Title: "My Karaoke Song", Channel: "Studio"},
	}
	found := RankSearch(songs, "karaoke", 20)
	if len(found) != 2 || found[0].VideoID != "aaaaaaaaaaa" {
		t.Fatalf("ranking = %+v", found)
	}
}

func TestLibraryDumpRoundTrip(t *testing.T) {
	now := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	songs := []LibrarySong{
		NewLibrarySong(SearchResult{VideoID: "aaaaaaaaaaa", Title: "L'été\nKaraoke", Channel: "Studio"}, "", now),
		NewLibrarySong(SearchResult{VideoID: "bbbbbbbbbbb", Title: "Second", Channel: "Queen"}, "", now.Add(time.Hour)),
	}
	raw := EncodeLibraryDump(songs)
	parsed, err := DecodeLibraryDump(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 2 || parsed[0].Title != "L'été\nKaraoke" || parsed[1].LastSeenAt.IsZero() {
		t.Fatalf("round trip = %+v", parsed)
	}
}

func TestLibraryDumpRejectsForeignSQL(t *testing.T) {
	cases := []string{
		"DROP TABLE songs;",
		LibraryDumpHeader + "\nBEGIN;\nDELETE FROM songs;\nCOMMIT;\n",
		LibraryDumpHeader + "\nBEGIN;\nUPDATE songs SET title = 'x';\nCOMMIT;\n",
		LibraryDumpHeader + "\nBEGIN;\nINSERT INTO songs (video_id, title) VALUES ('a', 'b');\nCOMMIT;\n",
	}
	for _, raw := range cases {
		if _, err := DecodeLibraryDump(raw); err == nil {
			t.Fatalf("accepted foreign SQL: %q", raw)
		}
	}
}

func TestValidateLibrarySongsRejectsInvalidRows(t *testing.T) {
	now := time.Now()
	valid := NewLibrarySong(SearchResult{VideoID: "aaaaaaaaaaa", Title: "Song"}, "", now)
	if _, err := ValidateLibrarySongs([]LibrarySong{valid, valid}); err != nil {
		t.Fatal(err)
	}
	invalid := valid
	invalid.VideoID = "short"
	if _, err := ValidateLibrarySongs([]LibrarySong{invalid}); err == nil {
		t.Fatal("accepted an invalid video id")
	}
	if _, err := ValidateLibrarySongs([]LibrarySong{valid, {SearchResult: SearchResult{VideoID: "bbbbbbbbbbb"}}}); err == nil {
		t.Fatal("accepted an empty title")
	}
	if !strings.Contains(EncodeLibraryDump([]LibrarySong{valid}), LibraryDumpHeader) {
		t.Fatal("dump header missing")
	}
}
