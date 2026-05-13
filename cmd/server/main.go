package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/kitti12911/lib-monitor/profiling"
	"github.com/kitti12911/lib-monitor/tracing"
	libconfig "github.com/kitti12911/lib-util/v3/config"
	"github.com/kitti12911/lib-util/v3/logger"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"

	userv1 "oas-sandbox/gen/grpc/user/v1"
	workerv1 "oas-sandbox/gen/grpc/worker/v1"
	"oas-sandbox/internal/config"
	"oas-sandbox/internal/server"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx := context.Background()

	// Load config
	cfg, err := libconfig.Load[config.Config]("config.yml")
	if err != nil {
		slog.ErrorContext(ctx, "failed to load config", "error", err)
		return 1
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
		return 1
	}
	defer func() {
		if shutdownErr := profiling.Shutdown(profiler); shutdownErr != nil {
			slog.ErrorContext(ctx, "failed to stop profiling", "error", shutdownErr)
		}
	}()

	// Init tracing
	tp, err := tracing.NewFromConfig(ctx, cfg.Service.Name, cfg.Tracing)
	if err != nil {
		slog.ErrorContext(ctx, "failed to init tracing", "error", err)
		return 1
	}
	defer func() {
		if shutdownErr := tracing.Shutdown(ctx, tp); shutdownErr != nil {
			slog.ErrorContext(ctx, "failed to stop tracing", "error", shutdownErr)
		}
	}()

	// Init gRPC clients
	userAddr := "dns:///" + net.JoinHostPort(cfg.UserService.Host, strconv.Itoa(cfg.UserService.Port))
	userConn, err := grpc.NewClient(
		userAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig":[{"round_robin":{}}]}`),
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to connect to user service", "addr", userAddr, "error", err)
		return 1
	}
	defer func() {
		if closeErr := userConn.Close(); closeErr != nil {
			slog.ErrorContext(ctx, "failed to close user service connection", "error", closeErr)
		}
	}()

	slog.InfoContext(ctx, "connected to user service", "addr", userAddr)

	// Start HTTP server
	srv := server.NewHTTPServer(
		cfg.Service.Port,
		cfg.Service.Name,
		userv1.NewUserServiceClient(userConn),
		workerv1.NewWorkerServiceClient(userConn),
	)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.Start()
	}()

	slog.InfoContext(ctx, "HTTP server started", "port", cfg.Service.Port)

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-quit:
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(ctx, "HTTP server error", "error", err)
			return 1
		}
	}

	slog.InfoContext(ctx, "shutting down HTTP server")

	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Service.ShutdownTimeout)
	defer cancel()

	srv.Stop(shutdownCtx)

	slog.InfoContext(ctx, "server stopped")

	return 0
}
