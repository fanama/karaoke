package main

import "testing"

func TestBaseURLPrefersExplicitValue(t *testing.T) {
	t.Setenv("PUBLIC_BASE_URL", "https://karaoke.example/")
	t.Setenv("RENDER_EXTERNAL_URL", "https://fana.onrender.com")

	if got := baseURL(); got != "https://karaoke.example" {
		t.Fatalf("baseURL() = %q, want the trimmed explicit URL", got)
	}
}

func TestBaseURLFallsBackToRender(t *testing.T) {
	t.Setenv("PUBLIC_BASE_URL", "")
	t.Setenv("RENDER_EXTERNAL_URL", "https://fana.onrender.com/")

	if got := baseURL(); got != "https://fana.onrender.com" {
		t.Fatalf("baseURL() = %q, want the trimmed Render URL", got)
	}
}

func TestBaseURLUnset(t *testing.T) {
	t.Setenv("PUBLIC_BASE_URL", "")
	t.Setenv("RENDER_EXTERNAL_URL", "")

	if got := baseURL(); got != "" {
		t.Fatalf("baseURL() = %q, want an empty string", got)
	}
}
