package main

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"fana-karaoke/internal/adapters/primary/httpapi"
	"fana-karaoke/internal/adapters/secondary/bus"
	"fana-karaoke/internal/adapters/secondary/clock"
	"fana-karaoke/internal/adapters/secondary/crypto"
	songd1 "fana-karaoke/internal/adapters/secondary/d1"
	"fana-karaoke/internal/adapters/secondary/invidious"
	"fana-karaoke/internal/adapters/secondary/memory"
	"fana-karaoke/internal/adapters/secondary/qrgen"
	songdb "fana-karaoke/internal/adapters/secondary/sqlite"
	"fana-karaoke/internal/ports"
	"fana-karaoke/internal/usecase"
	"fana-karaoke/web"
)

// app is the wired object graph.
type app struct {
	facade *usecase.Facade
	server *httpapi.Server
}

func wire(cfg settings, log *slog.Logger) *app {
	repos := newRepositories()
	events := bus.New()
	facade := usecase.NewFacade(usecase.Deps{
		Rooms:   repos.rooms,
		Devices: repos.devices,
		Events:  events,
		Catalog: newCatalog(cfg),
		Library: newLibrary(cfg, log),
		Clock:   clock.System{},
		IDs:     crypto.NewIDs(),
		Cfg:     cfg.Limits,
	})
	hub := httpapi.NewHub(facade, log)
	events.OnRoomUpdated(hub.PublishRoomUpdated)
	events.OnRoomEnded(hub.PublishRoomEnded)
	server := newServer(facade, hub, cfg, log, web.Dist())
	return &app{facade: facade, server: server}
}

type repositories struct {
	rooms   *memory.RoomRepo
	devices *memory.DeviceRepo
}

func newRepositories() repositories {
	return repositories{rooms: memory.NewRoomRepo(), devices: memory.NewDeviceRepo()}
}

// newLibrary selects the durable song store. A D1 warm-up failure leaves the
// adapter alive so reads keep working while writes are explicitly rejected.
func newLibrary(cfg settings, log *slog.Logger) ports.SongLibrary {
	switch cfg.LibraryBackend {
	case "none":
		return nil
	case "d1":
		return newD1Library(cfg, log)
	case "file":
		return newFileLibrary(cfg, log)
	default:
		log.Warn("unknown LIBRARY_BACKEND; using file", "value", cfg.LibraryBackend)
		return newFileLibrary(cfg, log)
	}
}

func newFileLibrary(cfg settings, log *slog.Logger) ports.SongLibrary {
	if cfg.LibraryPath == "" {
		log.Warn("LIBRARY_PATH is empty; the local song library is disabled")
		return nil
	}
	library, err := songdb.Open(cfg.LibraryPath)
	if err != nil {
		log.Warn("song library unavailable; every search will go online", "path", cfg.LibraryPath, "error", err)
		return nil
	}
	log.Info("song library ready", "backend", "file", "path", cfg.LibraryPath)
	return library
}

func newD1Library(cfg settings, log *slog.Logger) ports.SongLibrary {
	library := songd1.New(songd1.Config{
		BaseURL:           cfg.LibraryProxyURL,
		Token:             cfg.LibraryProxyToken,
		ReadTimeout:       cfg.LibraryRead,
		WriteTimeout:      cfg.LibraryWrite,
		CacheWriteTimeout: cfg.LibraryCacheWrite,
		WarmupTimeout:     cfg.LibraryWarmup,
		Log:               log,
	})
	ctx, cancel := context.WithTimeout(context.Background(), cfg.LibraryWarmup)
	defer cancel()
	if err := library.Warm(ctx); err != nil {
		log.Warn("d1 song library is cache-only; new songs will be rejected until it recovers", "error", err)
	}
	return library
}

func newCatalog(cfg settings) ports.VideoCatalog {
	upstream := invidious.NewClient(invidious.Options{Instances: cfg.Instances, Region: cfg.Region})
	prober := invidious.NewProber(&http.Client{Timeout: 6 * time.Second}, "")
	cache := invidious.NewCache(cfg.QueryTTL, cfg.StaleTTL, cfg.VideoTTL, time.Now)
	return invidious.NewCatalog(upstream, prober, cache)
}

func newServer(svc ports.KaraokeService, hub *httpapi.Hub, cfg settings, log *slog.Logger, static fs.FS) *httpapi.Server {
	return httpapi.New(svc, hub, crypto.NewSigner(cfg.Secret), memory.NewRateLimiter(), qrgen.New(), static, httpapi.Config{
		Dev:        cfg.Dev,
		BaseURL:    cfg.BaseURL,
		TrustProxy: cfg.TrustProxy,
		AdminToken: cfg.AdminToken,
		Version:    version,
	}, log)
}

func newHTTP(cfg settings, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}
