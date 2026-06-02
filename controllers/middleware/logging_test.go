package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestLoggerEmitsDebugAccessLog(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	logs := buf.String()
	if !strings.Contains(logs, "http request") {
		t.Fatalf("log output = %q, want request log", logs)
	}
	if !strings.Contains(logs, "path=/healthz") {
		t.Fatalf("log output = %q, want path", logs)
	}
	if !strings.Contains(logs, "status=204") {
		t.Fatalf("log output = %q, want status", logs)
	}
	if !strings.Contains(logs, "level=DEBUG") {
		t.Fatalf("log output = %q, want debug level", logs)
	}
}

func TestRequestLoggerElevatesClientAndServerErrors(t *testing.T) {
	tests := []struct {
		name  string
		code  int
		level string
	}{
		{name: "client error", code: http.StatusNotFound, level: "level=WARN"},
		{name: "server error", code: http.StatusInternalServerError, level: "level=ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
			handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.code)
			}))

			req := httptest.NewRequest(http.MethodGet, "/missing", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if logs := buf.String(); !strings.Contains(logs, tt.level) {
				t.Fatalf("log output = %q, want %s", logs, tt.level)
			}
		})
	}
}
