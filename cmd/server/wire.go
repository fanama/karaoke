package main

import (
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"fana-karaoke/internal/adapters/primary/httpapi"
	"fana-karaoke/internal/adapters/secondary/bus"
	"fana-karaoke/internal/adapters/secondary/clock"
	"fana-karaoke/internal/adapters/secondary/crypto"
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

// newLibrary opens the SQLite song store. A failure only disables the local
// cache: search keeps working through the upstream catalog.
func newLibrary(cfg settings, log *slog.Logger) ports.SongLibrary {
	if cfg.LibraryPath == "" {
		log.Warn("LIBRARY_PATH is empty; the local song library is disabled")
		return nil
	}
	library, err := songdb.Open(cfg.LibraryPath)
	if err != nil {
		log.Warn("song library unavailable; every search will go online", "path", cfg.LibraryPath, "error", err)
		return nil
	}
	log.Info("song library ready", "path", cfg.LibraryPath)
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
