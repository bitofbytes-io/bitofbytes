package main

import (
	"html"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DryWaters/bitofbytes/models"
)

func TestStatusIncludesEveryProjectWithoutDiagnostics(t *testing.T) {
	handler := newTestHandler()
	for range 3 {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/status", nil))
		body := html.UnescapeString(rr.Body.String())
		if rr.Code != 200 || strings.Count(body, "<dd>Down</dd>") != 8 || strings.Count(body, "<dt>") != 8 {
			t.Fatalf("unexpected initial status page: %d", rr.Code)
		}
		for _, project := range models.Projects() {
			if strings.Count(body, "<dt>"+project.Name+"</dt>") != 1 {
				t.Errorf("missing or duplicate %s", project.Name)
			}
		}
		for _, private := range []string{"/health", "/ready", "127.0.0.1", "localhost", "tail209", "degraded", "latency", "data-status", "status-dot"} {
			if strings.Contains(body, private) {
				t.Errorf("status leaks %q", private)
			}
		}
	}
	// Project availability cannot affect the site's own liveness/readiness.
	for _, route := range []string{"/health", "/healthz", "/ready"} {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, route, nil))
		if rr.Code != 200 {
			t.Errorf("%s=%d", route, rr.Code)
		}
	}
}

func TestStatusDoesNotAddIndicatorsToHomeOrProjectPages(t *testing.T) {
	for _, route := range []string{"/", "/projects/bitofbytes"} {
		body := get(t, route).Body.String()
		if !strings.Contains(body, `href="/status"`) {
			t.Errorf("%s has no status link", route)
		}
		for _, indicator := range []string{"nc-status__row", "status-dot", "data-status", "<dd>Up</dd>", "<dd>Down</dd>"} {
			if strings.Contains(body, indicator) {
				t.Errorf("%s has indicator %s", route, indicator)
			}
		}
	}
}
