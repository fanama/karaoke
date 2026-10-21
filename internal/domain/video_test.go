package domain

import "testing"

func TestParseYouTubeURL(t *testing.T) {
	valid := map[string]string{
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ":         "dQw4w9WgXcQ",
		"https://youtu.be/dQw4w9WgXcQ?t=12":                   "dQw4w9WgXcQ",
		"https://m.youtube.com/watch?v=dQw4w9WgXcQ&feature=x": "dQw4w9WgXcQ",
		"https://www.youtube.com/shorts/dQw4w9WgXcQ":          "dQw4w9WgXcQ",
		"https://www.youtube.com/embed/dQw4w9WgXcQ":           "dQw4w9WgXcQ",
	}
	for raw, want := range valid {
		got, err := ParseYouTubeURL(raw)
		if err != nil || got != want {
			t.Fatalf("%s: got %q, %v", raw, got, err)
		}
	}
	invalid := []string{
		"https://evil.example/watch?v=dQw4w9WgXcQ",
		"javascript:alert(1)",
		"https://youtu.be/nope",
		"https://www.youtube.com/watch?v=too-long-id-here",
		"",
	}
	for _, raw := range invalid {
		if _, err := ParseYouTubeURL(raw); err == nil {
			t.Fatalf("%q should be rejected", raw)
		}
	}
}

func TestValidateVideoMeta(t *testing.T) {
	ok := VideoMeta{VideoID: "dQw4w9WgXcQ", Title: "Song", DurationSeconds: 200}
	if err := ValidateVideoMeta(ok); err != nil {
		t.Fatalf("valid meta rejected: %v", err)
	}
	cases := []struct {
		meta VideoMeta
		want error
	}{
		{VideoMeta{Title: "Live", Live: true}, ErrVideoLive},
		{VideoMeta{Title: "Long", DurationSeconds: 601}, ErrVideoTooLong},
		{VideoMeta{Title: "Short", DurationSeconds: 60}, ErrVideoShort},
		{VideoMeta{DurationSeconds: 200}, ErrVideoMissing},
	}
	for _, c := range cases {
		if err := ValidateVideoMeta(c.meta); err != c.want {
			t.Fatalf("%+v: got %v want %v", c.meta, err, c.want)
		}
	}
}

func TestCleanNickname(t *testing.T) {
	if got, err := CleanNickname("  Alex  "); err != nil || got != "Alex" {
		t.Fatalf("got %q, %v", got, err)
	}
	if _, err := CleanNickname("   "); err == nil {
		t.Fatal("blank nickname should be rejected")
	}
}
