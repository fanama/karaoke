package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const version = "1.0.0"

func main() {
	applied, envErr := loadDotEnv(".env")
	log := buildLogger()
	logDotEnv(log, applied, envErr)
	cfg := loadSettings(log)
	warnConfig(log, cfg)
	application := wire(cfg, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go application.facade.Lifecycle.Start(ctx)

	server := newHTTP(cfg, application.server.Handler())
	go serve(server, cfg, log, cancel)
	waitForShutdown(log, cancel, server)
}

func logDotEnv(log *slog.Logger, applied int, err error) {
	if err != nil {
		log.Warn(".env could not be loaded", "error", err)
		return
	}
	if applied > 0 {
		log.Info(".env loaded", "variables", applied)
	}
}

func buildLogger() *slog.Logger {
	level := logLevel()
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}

func warnConfig(log *slog.Logger, cfg settings) {
	if cfg.BaseURL == "" {
		log.Warn("PUBLIC_BASE_URL is not set; deriving it from each request (fine for local development, not for production)")
	}
	if len(cfg.Instances) == 0 {
		log.Warn("no Invidious instances configured; search and new songs will fall back to URL paste/oEmbed")
	}
	if cfg.AdminToken == "" {
		log.Info("ADMIN_TOKEN is not set; the /admin page is disabled")
	} else if len(cfg.AdminToken) < 16 {
		log.Warn("ADMIN_TOKEN is short; use at least 16 random characters")
	}
}

func serve(server *http.Server, cfg settings, log *slog.Logger, cancel context.CancelFunc) {
	log.Info("Fana Karaoke listening", "addr", "http://localhost:"+cfg.Port, "version", version, "dev", cfg.Dev)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("listen failed", "error", err)
		cancel()
		os.Exit(1)
	}
}

func waitForShutdown(log *slog.Logger, cancel context.CancelFunc, server *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("shutting down")
	cancel()
	ctx, shutdown := context.WithTimeout(context.Background(), 8*time.Second)
	defer shutdown()
	if err := server.Shutdown(ctx); err != nil {
		log.Warn("shutdown", "error", err)
	}
}
