package main

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"strings"
)

// loadDotEnv reads KEY=VALUE pairs from path and sets those that are not
// already in the process environment. A missing file is not an error.
func loadDotEnv(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	defer file.Close()
	applied := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, ok := parseDotEnvLine(scanner.Text())
		if !ok {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			continue
		}
		applied++
	}
	return applied, scanner.Err()
}

// parseDotEnvLine accepts optional "export", comments and quoted values.
func parseDotEnvLine(raw string) (string, string, bool) {
	line := strings.TrimSpace(raw)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}
	line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
	key, value, ok := strings.Cut(line, "=")
	if !ok {
		return "", "", false
	}
	key = strings.TrimSpace(key)
	if !validEnvKey(key) {
		return "", "", false
	}
	return key, dotEnvValue(strings.TrimSpace(value)), true
}

func validEnvKey(key string) bool {
	if key == "" {
		return false
	}
	for i, r := range key {
		if i == 0 && r >= '0' && r <= '9' {
			return false
		}
		if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' {
			return false
		}
	}
	return true
}

func dotEnvValue(value string) string {
	if len(value) < 2 {
		return value
	}
	if value[0] == '\'' && value[len(value)-1] == '\'' {
		return value[1 : len(value)-1]
	}
	if value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
		value = strings.ReplaceAll(value, `\n`, "\n")
		return strings.ReplaceAll(value, `\"`, `"`)
	}
	return value
}
