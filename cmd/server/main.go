package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kitti12911/lib-monitor/profiling"
	"github.com/kitti12911/lib-monitor/tracing"
	libconfig "github.com/kitti12911/lib-util/v3/config"
	"github.com/kitti12911/lib-util/v3/logger"

	"oas-sandbox/internal/config"
	"oas-sandbox/internal/server"
)

func main() {
	ctx := context.Background()

	// Load config
	cfg, err := libconfig.Load[config.Config]("config.yml")
	if err != nil {
		slog.ErrorContext(ctx, "failed to load config", "error", err)
		os.Exit(1)
	}

	if cfg.Service.ShutdownTimeout == 0 {
		cfg.Service.ShutdownTimeout = 10 * time.Second
	}

	// Init logger
	logger.NewFromConfig(cfg.Logging, cfg.Service.Name)

	// Init monitoring
	profiler, err := profiling.NewFromConfig(cfg.Service.Name, cfg.Profiling)
	if err != nil {
		slog.ErrorContext(ctx, "failed to init profiling", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := profiling.Shutdown(profiler); err != nil {
			slog.ErrorContext(ctx, "failed to stop profiling", "error", err)
		}
	}()

	// Init tracing
	tp, err := tracing.NewFromConfig(ctx, cfg.Service.Name, cfg.Tracing)
	if err != nil {
		slog.ErrorContext(ctx, "failed to init tracing", "error", err)
		os.Exit(1)
	}
	defer tracing.Shutdown(ctx, tp)

	// Start HTTP server
	srv := server.NewHTTPServer(cfg.Service.Port, cfg.Service.Name)

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			slog.ErrorContext(ctx, "HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	slog.InfoContext(ctx, "HTTP server started", "port", cfg.Service.Port)

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.InfoContext(ctx, "shutting down HTTP server")

	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Service.ShutdownTimeout)
	defer cancel()

	srv.Stop(shutdownCtx)

	slog.InfoContext(ctx, "server stopped")
}
