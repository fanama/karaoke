package main

import (
	"crypto/rand"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"fana-karaoke/internal/ports"
)

// settings is everything the wiring layer needs from the environment.
type settings struct {
	LogLevel   slog.Level
	Dev        bool
	Port       string
	BaseURL    string
	Secret     []byte
	Instances  []string
	Region     string
	QueryTTL   time.Duration
	StaleTTL   time.Duration
	VideoTTL   time.Duration
	TrustProxy bool
	Limits     ports.Config
}

func loadSettings(log *slog.Logger) settings {
	return settings{
		LogLevel:   logLevel(),
		Dev:        envBool("DEV", false),
		Port:       env("PORT", "8080"),
		BaseURL:    strings.TrimRight(env("PUBLIC_BASE_URL", ""), "/"),
		Secret:     loadSecret(log),
		Instances:  splitList(env("INVIDIOUS_INSTANCES", defaultInstances)),
		Region:     env("YOUTUBE_REGION", "FR"),
		QueryTTL:   envDuration("SEARCH_CACHE_TTL", 24*time.Hour),
		StaleTTL:   envDuration("SEARCH_STALE_TTL", 7*24*time.Hour),
		VideoTTL:   envDuration("SEARCH_VIDEO_TTL", 7*24*time.Hour),
		TrustProxy: envBool("TRUST_PROXY", false),
		Limits: ports.Config{
			MaxRooms:  envInt("MAX_ROOMS", 1000),
			MaxGuests: envInt("MAX_GUESTS", 25),
			MaxQueue:  envInt("MAX_QUEUE", 100),
			RoomTTL:   envDuration("ROOM_TTL", 2*time.Hour),
		},
	}
}

const defaultInstances = "https://invidious.f5.si,https://yewtu.be,https://inv.nadeko.net,https://invidious.nerdvpn.de,https://iv.melmac.space"

func loadSecret(log *slog.Logger) []byte {
	secret := []byte(env("SESSION_SECRET", ""))
	if len(secret) > 0 {
		return secret
	}
	secret = make([]byte, 32)
	_, _ = rand.Read(secret)
	log.Warn("SESSION_SECRET is not set; sessions will not survive a restart")
	return secret
}

func logLevel() slog.Level {
	switch strings.ToLower(env("LOG_LEVEL", "info")) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	}
	return slog.LevelInfo
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v, err := strconv.Atoi(env(key, ""))
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func envBool(key string, fallback bool) bool {
	v, err := strconv.ParseBool(env(key, ""))
	if err != nil {
		return fallback
	}
	return v
}

func envDuration(key string, fallback time.Duration) time.Duration {
	v, err := time.ParseDuration(env(key, ""))
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func splitList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
