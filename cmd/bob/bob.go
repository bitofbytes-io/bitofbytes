package main

import (
	"log/slog"
	"net/http"
	"os"
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
	// TODO:  Setup database

	// setup services
	postService := models.PostService{
		// Add DB when needed
	}
	base64Service := models.Base64Service{
		// Add DB when needed
	}

	// setup CSRF protection
	csrfMw := middleware.CSRF(cfg.CSRF.Key, cfg.CSRF.Secure)
	cspMw := middleware.SecureHeaders
	loggingMw := middleware.RequestLogger(logger)

	// setup controllers
	blogController := controllers.Blog{
		PostService: postService,
		Templates: controllers.BlogTemplates{
			Index: views.Must(views.ParseFS(templates.FS, "blog/index.gohtml", "base.gohtml")),
			Post:  views.Must(views.ParseFS(templates.FS, "blog/post.gohtml", "base.gohtml")),
		},
	}

	utilsController := controllers.Utils{
		Base64Service: base64Service,
		Templates: controllers.UtilsTemplates{
			Index: views.Must(views.ParseFS(templates.FS, "utils/index.gohtml", "base.gohtml")),
			Base64: controllers.Base64Templates{
				Base64Response: views.Must(views.ParseFS(templates.FS, "utils/base64/base64_response.gohtml")),
			},
		},
	}

	// Setup our router and routes
	r := http.NewServeMux()
	var handler http.Handler = r
	handler = csrfMw(handler)
	handler = cspMw(handler)
	handler = loggingMw(handler)
	r.HandleFunc("GET /", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "home/index.gohtml", "home/infocard.gohtml", "base.gohtml"))))

	// Healthcheck
	r.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Utils
	r.HandleFunc("GET /utils", utilsController.Index)

	// Base64 Utils
	r.HandleFunc("GET /utils/base64/encode", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "utils/base64/encode.gohtml", "base.gohtml"))))
	r.HandleFunc("POST /utils/base64/encode", utilsController.Encode)
	r.HandleFunc("GET /utils/base64/decode", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "utils/base64/decode.gohtml", "base.gohtml"))))
	r.HandleFunc("POST /utils/base64/decode", utilsController.Decode)

	// Blog
	r.HandleFunc("GET /blog", blogController.Index)
	r.HandleFunc("GET /posts/{slug}", blogController.Blog)

	staticHandler := http.FileServer(http.Dir("static"))
	r.Handle("GET /static/", http.StripPrefix("/static/", staticHandler))

	// Start the server
	server := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           handler,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	logger.Info("Starting the server", "address", cfg.Server.Address, "version", version, "revision", revision)

	return server.ListenAndServe()
}
