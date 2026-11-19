package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/DryWaters/bitofbytes/internal/metrics"
	"github.com/DryWaters/bitofbytes/internal/server"
	"github.com/DryWaters/bitofbytes/internal/status"
	"github.com/DryWaters/bitofbytes/models"
)

var (
	version  = "dev"
	revision = "unknown"
)

func main() {
	// load config
	cfg, err := models.LoadEnvConfig()
	if err != nil {
		fatal("load configuration", err)
	}

	logger, err := models.NewLogger(cfg.Logging)
	if err != nil {
		fatal("initialize logger", err)
	}

	slog.SetDefault(logger)

	if err := run(cfg, logger); err != nil {
		logger.Error("server exited", "error", err)
		os.Exit(1)
	}
}

// fatal reports a startup failure before the configured logger exists.
func fatal(msg string, err error) {
	slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})).Error(msg, "error", err)
	os.Exit(1)
}

// shutdownTimeout bounds how long in-flight requests get to finish on
// SIGTERM; it stays under Docker's default 10s stop grace period.
const shutdownTimeout = 5 * time.Second

func run(cfg models.Config, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	m := metrics.New()
	projectStatus := status.New()
	srv := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           server.New(cfg, logger, m, "static", assetVersion(), projectStatus),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	// Bind the public listener before the metrics one, so an address clash
	// always fails the optional metrics listener rather than the site.
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}

	var background sync.WaitGroup
	background.Go(func() { projectStatus.Run(ctx) })
	// The metrics listener is internal only and never fatal: it follows the
	// public server's lifecycle and is shut down when serve returns.
	background.Go(func() { m.Serve(ctx, cfg.Metrics.Address, logger) })

	logger.Info("Starting the server", "address", cfg.Server.Address, "version", version, "revision", revision)

	err = serve(ctx, srv, ln, logger)
	stop()
	background.Wait()
	return err
}

// serve runs server on ln until it fails or ctx is done, then shuts it down
// gracefully.
func serve(ctx context.Context, server *http.Server, ln net.Listener, logger *slog.Logger) error {
	errs := make(chan error, 1)
	go func() { errs <- server.Serve(ln) }()

	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
	}

	logger.Info("Shutting down the server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}

// assetVersion is the ?v= value on CSS and JS URLs: the git revision in a
// release build. A development build has none, so its assets are never cached
// as immutable and edits show up on reload.
func assetVersion() string {
	if revision == "unknown" {
		return ""
	}
	return revision
}
