// Package server builds the site's HTTP handler: routes, static file serving,
// and the middleware chain.
package server

import (
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/DryWaters/bitofbytes/controllers"
	"github.com/DryWaters/bitofbytes/controllers/middleware"
	"github.com/DryWaters/bitofbytes/internal/metrics"
	"github.com/DryWaters/bitofbytes/internal/status"
	"github.com/DryWaters/bitofbytes/models"
	"github.com/DryWaters/bitofbytes/templates"
	"github.com/DryWaters/bitofbytes/views"
)

// New builds the site's HTTP handler: the routes, static files, and the
// middleware chain. staticDir is the directory served under /static/ and
// assetVersion the ?v= value of this release's CSS and JS URLs (empty in a
// development build).
func New(cfg models.Config, logger *slog.Logger, m *metrics.Metrics, staticDir string, assetVersion string, projectStatus *status.Checker) http.Handler {
	portfolio := controllers.Portfolio{
		Projects:   models.Projects(),
		Activities: models.CurrentActivities(),
		Updates:    models.AllUpdates(),
		Templates: controllers.PortfolioTemplates{
			Home:          views.Must(views.ParseFS(assetVersion, templates.FS, "home/index.gohtml", "base.gohtml")),
			ProjectsIndex: views.Must(views.ParseFS(assetVersion, templates.FS, "projects/index.gohtml", "base.gohtml")),
			ProjectDetail: views.Must(views.ParseFS(assetVersion, templates.FS, "projects/detail.gohtml", "base.gohtml")),
			UpdatesIndex:  views.Must(views.ParseFS(assetVersion, templates.FS, "updates/index.gohtml", "base.gohtml")),
			NotFound:      views.Must(views.ParseFS(assetVersion, templates.FS, "errors/404.gohtml", "base.gohtml")),
		},
	}

	r := http.NewServeMux()
	r.HandleFunc("GET /{$}", portfolio.Home)
	r.HandleFunc("GET /projects", portfolio.ProjectsIndex)
	r.HandleFunc("GET /projects/{slug}", portfolio.ProjectDetail)
	r.HandleFunc("GET /projects/{slug}/feed.xml", portfolio.ProjectFeed)
	r.HandleFunc("GET /updates", portfolio.UpdatesIndex)
	r.Handle("GET /status", controllers.Status{Snapshot: projectStatus.Snapshot, Page: views.Must(views.ParseFS(assetVersion, templates.FS, "status/index.gohtml", "base.gohtml"))})
	r.HandleFunc("GET /feed.xml", portfolio.Feed)
	// /health means the process is up and /ready means its dependencies are.
	// bob has no database or mount, so /ready has nothing to check and always
	// reports ok. /healthz stays for the image's HEALTHCHECK.
	r.HandleFunc("GET /health", health)
	r.HandleFunc("GET /healthz", health)
	r.HandleFunc("GET /ready", ready)
	// Support browser default icon discovery paths in addition to the template's
	// explicit /static/... icon links.
	r.Handle("GET /favicon.ico", cacheStatic(assetVersion, serveStaticFile(staticDir, "favicon.ico")))
	r.Handle("GET /robots.txt", cacheStatic(assetVersion, serveStaticFile(staticDir, "robots.txt")))
	r.HandleFunc("GET /sitemap.xml", portfolio.Sitemap)
	r.Handle("GET /apple-touch-icon.png", cacheStatic(assetVersion, serveStaticFile(staticDir, "apple-touch-icon.png")))
	r.Handle("GET /apple-touch-icon-precomposed.png", cacheStatic(assetVersion, serveStaticFile(staticDir, "apple-touch-icon.png")))

	staticHandler := http.FileServer(noDirFS{http.Dir(staticDir)})
	r.Handle("GET /static/", cacheStatic(assetVersion, http.StripPrefix("/static/", staticHandler)))
	// Catch-all for paths no route above matches: the branded 404 page.
	r.HandleFunc("GET /", portfolio.NotFound)

	var handler http.Handler = r
	handler = middleware.CSRF()(handler)
	handler = middleware.SecureHeaders(cfg.CSRF.Secure)(handler)
	handler = middleware.RequestLogger(logger)(handler)
	// Outermost, so latency covers every middleware and CSRF rejections count.
	// /metrics is not routed here; it is served only on the metrics listener.
	handler = m.Middleware(metricsRoute(r))(handler)

	return handler
}

// metricsRoute labels requests by matched route. The "GET /" 404 catch-all
// counts as unmatched, as it did before the branded page, so unknown paths
// stay one label.
func metricsRoute(mux *http.ServeMux) func(*http.Request) string {
	route := metrics.ServeMuxRoute(mux)
	return func(r *http.Request) string {
		if p := route(r); p != "/" {
			return p
		}
		return ""
	}
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func ready(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func serveStaticFile(staticDir string, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticDir, name))
	}
}
