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

	return newHandler(cfg, logger)
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
