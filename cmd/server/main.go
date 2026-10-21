package main

import (
	"context"
	"crypto/rand"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"fana-karaoke/internal/app"
	"fana-karaoke/internal/httpapi"
	"fana-karaoke/internal/search"
	"fana-karaoke/web"
)

const version = "1.0.0"

func main() {
	logLevel := slog.LevelInfo
	switch strings.ToLower(env("LOG_LEVEL", "info")) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	dev := envBool("DEV", false)
	port := env("PORT", "8080")
	baseURL := strings.TrimRight(env("PUBLIC_BASE_URL", ""), "/")
	if baseURL == "" {
		log.Warn("PUBLIC_BASE_URL is not set; deriving it from each request (fine for local development, not for production)")
	}

	secret := []byte(env("SESSION_SECRET", ""))
	if len(secret) == 0 {
		secret = make([]byte, 32)
		_, _ = rand.Read(secret)
		log.Warn("SESSION_SECRET is not set; sessions will not survive a restart")
	}

	instances := splitList(env("INVIDIOUS_INSTANCES", "https://invidious.f5.si,https://yewtu.be,https://inv.nadeko.net,https://invidious.nerdvpn.de,https://iv.melmac.space"))
	catalog := search.New(search.Options{
		Instances: instances,
		Region:    env("YOUTUBE_REGION", "FR"),
		QueryTTL:  envDuration("SEARCH_CACHE_TTL", 24*time.Hour),
		StaleTTL:  envDuration("SEARCH_STALE_TTL", 7*24*time.Hour),
	})
	if len(instances) == 0 {
		log.Warn("no Invidious instances configured; search and new songs will fall back to URL paste/oEmbed")
	}

	mgr := app.NewManager(app.Config{
		MaxRooms:  envInt("MAX_ROOMS", 1000),
		MaxGuests: envInt("MAX_GUESTS", 25),
		MaxQueue:  envInt("MAX_QUEUE", 100),
		RoomTTL:   envDuration("ROOM_TTL", 2*time.Hour),
	}, catalog)

	srv, err := httpapi.New(mgr, catalog, web.Dist(), httpapi.Config{
		Dev:        dev,
		BaseURL:    baseURL,
		Secret:     secret,
		TrustProxy: envBool("TRUST_PROXY", false),
		Version:    version,
	}, log)
	if err != nil {
		log.Error("server setup failed", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go mgr.Start(ctx)

	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Info("Fana Karaoke listening", "addr", "http://localhost:"+port, "version", version, "dev", dev)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("listen failed", "error", err)
			cancel()
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("shutting down")
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Warn("shutdown", "error", err)
	}
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
