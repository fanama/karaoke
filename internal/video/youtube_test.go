package video

import "testing"

func TestParseURL(t *testing.T) {
	valid := map[string]string{
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ":         "dQw4w9WgXcQ",
		"https://youtu.be/dQw4w9WgXcQ?t=12":                   "dQw4w9WgXcQ",
		"https://m.youtube.com/watch?v=dQw4w9WgXcQ&feature=x": "dQw4w9WgXcQ",
		"https://www.youtube.com/shorts/dQw4w9WgXcQ":          "dQw4w9WgXcQ",
		"https://www.youtube.com/embed/dQw4w9WgXcQ":           "dQw4w9WgXcQ",
	}
	for raw, want := range valid {
		got, err := ParseURL(raw)
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
		if _, err := ParseURL(raw); err == nil {
			t.Fatalf("%q should be rejected", raw)
		}
	}
}
