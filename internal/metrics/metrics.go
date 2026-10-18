// Package metrics exports Prometheus metrics on an internal listener that is
// separate from the public one, so the public reverse proxy routes can never
// reach /metrics.
package metrics

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// DefaultAddr is the metrics listen address when METRICS_ADDR is unset.
const DefaultAddr = ":9464"

// UnmatchedRoute is the route label for requests that matched no route, so
// unknown paths never become label values.
const UnmatchedRoute = "unmatched"

// shutdownTimeout bounds how long an in-flight scrape gets on shutdown.
const shutdownTimeout = 5 * time.Second

// skipPaths are probe and scrape endpoints left out of the RED metrics.
var skipPaths = map[string]bool{
	"/health":  true,
	"/healthz": true,
	"/ready":   true,
	"/metrics": true,
}

// Metrics owns a private registry (not the global default) with the Go
// runtime and process collectors and the HTTP RED series.
type Metrics struct {
	registry *prometheus.Registry
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

// New returns a Metrics with its collectors registered.
func New() *Metrics {
	m := &Metrics{
		registry: prometheus.NewRegistry(),
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP requests by method, route pattern, and status code.",
		}, []string{"method", "route", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency by method, route pattern, and status code.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route", "status"}),
	}
	m.registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		m.requests,
		m.duration,
	)
	return m
}

// Registry is where app-specific collectors (DB pool, domain counters) go.
func (m *Metrics) Registry() *prometheus.Registry {
	return m.registry
}

// Handler serves GET /metrics from the private registry. Mount it only on the
// internal listener.
func (m *Metrics) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{Registry: m.registry}))
	return mux
}

// Middleware records http_requests_total and http_request_duration_seconds.
// route is called after the handler returns and must give the matched route
// pattern, or "" when nothing matched; it must never return the raw path.
func (m *Metrics) Middleware(route func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skipPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w}
			next.ServeHTTP(rec, r)

			pattern := route(r)
			if pattern == "" {
				pattern = UnmatchedRoute
			}
			labels := prometheus.Labels{
				"method": r.Method,
				"route":  pattern,
				"status": strconv.Itoa(rec.Status()),
			}
			m.requests.With(labels).Inc()
			m.duration.With(labels).Observe(time.Since(start).Seconds())
		})
	}
}

// ServeMuxRoute resolves the route label from an http.ServeMux: the
// registered pattern without its method ("GET /projects/{slug}" becomes
// "/projects/{slug}"), or "" for 404s and 405s.
func ServeMuxRoute(mux *http.ServeMux) func(*http.Request) string {
	return func(r *http.Request) string {
		_, pattern := mux.Handler(r)
		if _, path, ok := strings.Cut(pattern, " "); ok {
			return path
		}
		return pattern
	}
}

// Serve runs the metrics listener on addr until ctx is done, then shuts it
// down gracefully. A listener that cannot bind is logged and Serve returns at
// once, leaving the app running without metrics. An empty addr disables it.
func (m *Metrics) Serve(ctx context.Context, addr string, logger *slog.Logger) {
	if addr == "" {
		logger.Info("Metrics listener disabled")
		return
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           m.Handler(),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	errs := make(chan error, 1)
	go func() { errs <- server.ListenAndServe() }()
	logger.Info("Starting the metrics listener", "address", addr)

	select {
	case err := <-errs:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("metrics listener failed", "address", addr, "error", err)
		}
		return
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shut down metrics listener", "error", err)
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) Status() int {
	if s.status == 0 {
		return http.StatusOK
	}
	return s.status
}

func (s *statusRecorder) WriteHeader(statusCode int) {
	// 1xx responses are informational; the final status comes later.
	if s.status == 0 && statusCode >= http.StatusOK {
		s.status = statusCode
	}
	s.ResponseWriter.WriteHeader(statusCode)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.ResponseWriter.Write(b)
}

func (s *statusRecorder) Flush() {
	if flusher, ok := s.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (s *statusRecorder) Unwrap() http.ResponseWriter {
	return s.ResponseWriter
}
