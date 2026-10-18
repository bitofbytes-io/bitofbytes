package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DryWaters/bitofbytes/internal/metrics"
	"github.com/DryWaters/bitofbytes/models"
)

// The public handler never serves /metrics, and its requests are labelled by
// route pattern on the separate metrics handler.
func TestPublicRequestsAreRecordedByRoutePattern(t *testing.T) {
	t.Parallel()
	var cfg models.Config
	m := metrics.New()
	handler := newHandler(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), m, "../../static", testAssetVersion)

	for _, path := range []string{"/", "/projects/carma", "/projects/noted", "/static/styles.css", "/does-not-exist", "/health", "/ready"} {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("public GET /metrics = %d, want 404", rr.Code)
	}

	scrape := httptest.NewRecorder()
	m.Handler().ServeHTTP(scrape, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := scrape.Body.String()
	for _, want := range []string{
		`http_requests_total{method="GET",route="/{$}",status="200"} 1`,
		`http_requests_total{method="GET",route="/projects/{slug}",status="200"} 2`,
		`http_requests_total{method="GET",route="/static/",status="200"} 1`,
		`http_requests_total{method="GET",route="unmatched",status="404"} 1`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics output missing %q", want)
		}
	}
	for _, raw := range []string{"/projects/carma", "/static/styles.css", "/does-not-exist", `route="/health"`, `route="/ready"`} {
		if strings.Contains(body, raw) {
			t.Errorf("metrics output contains %q", raw)
		}
	}
}
