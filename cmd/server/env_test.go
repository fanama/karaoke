package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDotEnvLine(t *testing.T) {
	cases := []struct {
		line  string
		key   string
		value string
		ok    bool
	}{
		{"ADMIN_TOKEN=secret", "ADMIN_TOKEN", "secret", true},
		{"  export LOG_LEVEL=debug  ", "LOG_LEVEL", "debug", true},
		{`QUOTED="two words"`, "QUOTED", "two words", true},
		{"SINGLE='two words'", "SINGLE", "two words", true},
		{`ESCAPED="line\nbreak"`, "ESCAPED", "line\nbreak", true},
		{"EMPTY=", "EMPTY", "", true},
		{"# comment", "", "", false},
		{"NO_EQUALS", "", "", false},
		{"1BAD=x", "", "", false},
		{"=x", "", "", false},
	}
	for _, tc := range cases {
		key, value, ok := parseDotEnvLine(tc.line)
		if ok != tc.ok || key != tc.key || value != tc.value {
			t.Errorf("%q: got (%q, %q, %v), want (%q, %q, %v)",
				tc.line, key, value, ok, tc.key, tc.value, tc.ok)
		}
	}
}

func TestLoadDotEnvAppliesMissingVariablesOnly(t *testing.T) {
	keys := []string{
		"FANA_TEST_DOTENV_A", "FANA_TEST_DOTENV_B",
		"FANA_TEST_DOTENV_C", "FANA_TEST_DOTENV_EXISTING",
	}
	for _, key := range keys {
		_ = os.Unsetenv(key)
	}
	t.Cleanup(func() {
		for _, key := range keys {
			_ = os.Unsetenv(key)
		}
	})
	path := filepath.Join(t.TempDir(), ".env")
	content := "# comment\n" +
		"FANA_TEST_DOTENV_A=from-file\n" +
		"export FANA_TEST_DOTENV_B=\"two words\"\n" +
		"FANA_TEST_DOTENV_C='single # hash'\n" +
		"FANA_TEST_DOTENV_EXISTING=file\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FANA_TEST_DOTENV_EXISTING", "process")
	applied, err := loadDotEnv(path)
	if err != nil {
		t.Fatal(err)
	}
	if applied != 3 {
		t.Fatalf("applied = %d, want 3", applied)
	}
	assertEnv(t, "FANA_TEST_DOTENV_A", "from-file")
	assertEnv(t, "FANA_TEST_DOTENV_B", "two words")
	assertEnv(t, "FANA_TEST_DOTENV_C", "single # hash")
	assertEnv(t, "FANA_TEST_DOTENV_EXISTING", "process")
}

func assertEnv(t *testing.T, key, want string) {
	t.Helper()
	if got := os.Getenv(key); got != want {
		t.Fatalf("%s = %q, want %q", key, got, want)
	}
}

func TestLoadDotEnvMissingFile(t *testing.T) {
	applied, err := loadDotEnv(filepath.Join(t.TempDir(), ".env"))
	if applied != 0 || err != nil {
		t.Fatalf("missing .env: applied=%d err=%v", applied, err)
	}
}
