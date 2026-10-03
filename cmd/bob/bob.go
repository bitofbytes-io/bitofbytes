package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/DryWaters/bitofbytes/controllers"
	"github.com/DryWaters/bitofbytes/controllers/middleware"
	"github.com/DryWaters/bitofbytes/models"
	"github.com/DryWaters/bitofbytes/templates"
	"github.com/DryWaters/bitofbytes/views"
)

var (
	version  = "dev"
	revision = "unknown"
)

func main() {
	// load config
	cfg, err := models.LoadEnvConfig()
	if err != nil {
		fallback := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
		fallback.Error("load configuration", "error", err)
		os.Exit(1)
	}

	logger, err := models.NewLogger(cfg.Logging)
	if err != nil {
		fallback := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
		fallback.Error("initialize logger", "error", err)
		os.Exit(1)
	}

	slog.SetDefault(logger)

	if err := run(cfg, logger); err != nil {
		logger.Error("server exited", "error", err)
		os.Exit(1)
	}
}

// shutdownTimeout bounds how long in-flight requests get to finish on
// SIGTERM; it stays under Docker's default 10s stop grace period.
const shutdownTimeout = 5 * time.Second

func run(cfg models.Config, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           newHandler(cfg, logger, "static", assetVersion()),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	logger.Info("Starting the server", "address", cfg.Server.Address, "version", version, "revision", revision)

	return serve(ctx, server, logger)
}

// serve runs server until it fails or ctx is done, then shuts it down
// gracefully. A listener that cannot bind returns its error at once.
func serve(ctx context.Context, server *http.Server, logger *slog.Logger) error {
	errs := make(chan error, 1)
	go func() { errs <- server.ListenAndServe() }()

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
// release build, or the start time in development, where air restarts the
// server on every change.
func assetVersion() string {
	if revision != "unknown" {
		return revision
	}
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func newHandler(cfg models.Config, logger *slog.Logger, staticDir string, assetVersion string) http.Handler {
	portfolio := controllers.Portfolio{
		Projects:   models.Projects(),
		Activities: models.CurrentActivities(),
		Templates: controllers.PortfolioTemplates{
			Home:          views.Must(views.ParseFS(assetVersion, templates.FS, "home/index.gohtml", "base.gohtml")),
			ProjectsIndex: views.Must(views.ParseFS(assetVersion, templates.FS, "projects/index.gohtml", "base.gohtml")),
			ProjectDetail: views.Must(views.ParseFS(assetVersion, templates.FS, "projects/detail.gohtml", "base.gohtml")),
		},
	}

	r := http.NewServeMux()
	r.HandleFunc("GET /{$}", portfolio.Home)
	r.HandleFunc("GET /projects", portfolio.ProjectsIndex)
	r.HandleFunc("GET /projects/{slug}", portfolio.ProjectDetail)
	r.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	// Support browser default icon discovery paths in addition to the template's
	// explicit /static/... icon links.
	r.Handle("GET /favicon.ico", cacheStatic(assetVersion, serveStaticFile(staticDir, "favicon.ico")))
	r.Handle("GET /apple-touch-icon.png", cacheStatic(assetVersion, serveStaticFile(staticDir, "apple-touch-icon.png")))
	r.Handle("GET /apple-touch-icon-precomposed.png", cacheStatic(assetVersion, serveStaticFile(staticDir, "apple-touch-icon.png")))

	staticHandler := http.FileServer(noDirFS{http.Dir(staticDir)})
	r.Handle("GET /static/", cacheStatic(assetVersion, http.StripPrefix("/static/", staticHandler)))

	var handler http.Handler = r
	handler = middleware.CSRF()(handler)
	handler = middleware.SecureHeaders(cfg.CSRF.Secure)(handler)
	handler = middleware.RequestLogger(logger)(handler)

	return handler
}

func serveStaticFile(staticDir string, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticDir, name))
	}
}
