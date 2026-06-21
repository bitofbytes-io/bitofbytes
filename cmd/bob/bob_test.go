package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DryWaters/bitofbytes/models"
)

func newTestHandler() http.Handler {
	var cfg models.Config
	cfg.CSRF.Key = []byte("01234567890123456789012345678901")
	cfg.CSRF.Secure = false

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return newHandlerWithStaticDir(cfg, logger, "../../static")
}

func TestRoutesRenderCurrentSiteSurface(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()

	tests := []struct {
		path string
		want string
	}{
		{path: "/", want: "Daniel Waters"},
		{path: "/projects", want: "Selected personal projects"},
		{path: "/projects/permitpal", want: "PermitPal"},
		{path: "/projects/dined", want: "Dined"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("%s status code = %d, want %d", tt.path, rr.Code, http.StatusOK)
		}
		if body := rr.Body.String(); !strings.Contains(body, tt.want) {
			t.Fatalf("%s body missing %q", tt.path, tt.want)
		}
	}
}

func TestRemovedRoutesReturnNotFound(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()

	for _, path := range []string{
		"/blog",
		"/posts/1",
		"/utils",
		"/utils/base64/encode",
		"/utils/base64/decode",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("%s status code = %d, want %d", path, rr.Code, http.StatusNotFound)
		}
	}
}

func TestIconRoutesServeStaticFiles(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()

	tests := []struct {
		path            string
		wantContentType string
	}{
		{path: "/favicon.ico", wantContentType: "image/x-icon"},
		{path: "/apple-touch-icon.png", wantContentType: "image/png"},
		{path: "/apple-touch-icon-precomposed.png", wantContentType: "image/png"},
		{path: "/static/favicon.ico", wantContentType: "image/x-icon"},
		{path: "/static/favicon.svg", wantContentType: "image/svg+xml"},
		{path: "/static/favicon-32x32.png", wantContentType: "image/png"},
		{path: "/static/favicon-16x16.png", wantContentType: "image/png"},
		{path: "/static/apple-touch-icon.png", wantContentType: "image/png"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("%s status code = %d, want %d", tt.path, rr.Code, http.StatusOK)
		}
		if contentType := rr.Header().Get("Content-Type"); !strings.HasPrefix(contentType, tt.wantContentType) {
			t.Fatalf("%s Content-Type = %q, want prefix %q", tt.path, contentType, tt.wantContentType)
		}
		if rr.Body.Len() == 0 {
			t.Fatalf("%s served an empty body", tt.path)
		}
	}
}

func TestUnknownProjectReturnsNotFound(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/projects/not-real", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("unknown project status code = %d, want %d", rr.Code, http.StatusNotFound)
	}
}
