package main

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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

func run(cfg models.Config, logger *slog.Logger) error {
	server := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           newHandler(cfg, logger),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	logger.Info("Starting the server", "address", cfg.Server.Address, "version", version, "revision", revision)

	return server.ListenAndServe()
}

func newHandler(cfg models.Config, logger *slog.Logger) http.Handler {
	return newHandlerWithStaticDir(cfg, logger, "static")
}

func newHandlerWithStaticDir(cfg models.Config, logger *slog.Logger, staticDir string) http.Handler {
	portfolio := controllers.Portfolio{
		Projects: models.Projects(),
		Templates: controllers.PortfolioTemplates{
			Home:          views.Must(views.ParseFS(templates.FS, "home/index.gohtml", "base.gohtml")),
			ProjectsIndex: views.Must(views.ParseFS(templates.FS, "projects/index.gohtml", "base.gohtml")),
			ProjectDetail: views.Must(views.ParseFS(templates.FS, "projects/detail.gohtml", "base.gohtml")),
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
	r.HandleFunc("GET /favicon.ico", serveStaticFile(staticDir, "favicon.ico"))
	r.HandleFunc("GET /apple-touch-icon.png", serveStaticFile(staticDir, "apple-touch-icon.png"))
	r.HandleFunc("GET /apple-touch-icon-precomposed.png", serveStaticFile(staticDir, "apple-touch-icon.png"))

	staticHandler := http.FileServer(http.Dir(staticDir))
	r.Handle("GET /static/", http.StripPrefix("/static/", staticHandler))

	var handler http.Handler = r
	handler = middleware.CSRF(cfg.CSRF.Key, cfg.CSRF.Secure)(handler)
	handler = middleware.SecureHeaders(cfg.CSRF.Secure)(handler)
	handler = middleware.RequestLogger(logger)(handler)

	return handler
}

func serveStaticFile(staticDir string, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticDir, name))
	}
}
