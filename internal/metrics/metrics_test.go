package metrics

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func scrape(t *testing.T, m *Metrics) string {
	t.Helper()
	rr := httptest.NewRecorder()
	m.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /metrics = %d, want 200", rr.Code)
	}
	return rr.Body.String()
}

func TestMiddlewareLabelsByRoutePattern(t *testing.T) {
	t.Parallel()
	m := New()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("GET /broken", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {})
	handler := m.Middleware(ServeMuxRoute(mux))(mux)

	for _, target := range []string{
		"/items/42?user=alice",
		"/items/43",
		"/broken",
		"/no-such-page",
		"/health",
		"/ready",
		"/metrics",
	} {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, target, nil))
	}
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/items/44", nil))

	body := scrape(t, m)
	for _, want := range []string{
		`http_requests_total{method="GET",route="/items/{id}",status="200"} 2`,
		`http_requests_total{method="GET",route="/broken",status="500"} 1`,
		`http_requests_total{method="GET",route="unmatched",status="404"} 1`,
		`http_requests_total{method="POST",route="unmatched",status="405"} 1`,
		`http_request_duration_seconds_count{method="GET",route="/items/{id}",status="200"} 2`,
		`http_request_duration_seconds_bucket{method="GET",route="/items/{id}",status="200",le="+Inf"} 2`,
		"go_goroutines ",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics output missing %q", want)
		}
	}
	for _, raw := range []string{"/items/42", "/items/43", "alice", "/no-such-page", `route="/health"`, `route="/ready"`, `route="/metrics"`} {
		if strings.Contains(body, raw) {
			t.Errorf("metrics output contains raw path or skipped route %q", raw)
		}
	}
}

func TestServeMuxRouteStripsMethod(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/any", func(w http.ResponseWriter, r *http.Request) {})
	route := ServeMuxRoute(mux)

	tests := map[string]string{
		"/items/7": "/items/{id}",
		"/any":     "/any",
		"/missing": "",
	}
	for target, want := range tests {
		if got := route(httptest.NewRequest(http.MethodGet, target, nil)); got != want {
			t.Errorf("route(%s) = %q, want %q", target, got, want)
		}
	}
}

func freeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()
	return addr
}

func TestServeExposesMetricsUntilContextIsDone(t *testing.T) {
	t.Parallel()
	m := New()
	addr := freeAddr(t)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		m.Serve(ctx, addr, discardLogger())
	}()

	var resp *http.Response
	for start := time.Now(); ; {
		var err error
		if resp, err = http.Get("http://" + addr + "/metrics"); err == nil {
			break
		}
		if time.Since(start) > 5*time.Second {
			t.Fatal("metrics listener never started")
		}
		time.Sleep(10 * time.Millisecond)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || !bytes.Contains(body, []byte("go_goroutines")) {
		t.Fatalf("GET /metrics = %d, body missing go_goroutines", resp.StatusCode)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(shutdownTimeout + time.Second):
		t.Fatal("Serve did not return after the context was done")
	}
}

func TestServeLogsBindFailureWithoutBlocking(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	var logs bytes.Buffer
	done := make(chan struct{})
	go func() {
		defer close(done)
		New().Serve(context.Background(), ln.Addr().String(), slog.New(slog.NewTextHandler(&logs, nil)))
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Serve blocked on an address already in use")
	}
	if !strings.Contains(logs.String(), "metrics listener failed") {
		t.Errorf("bind failure was not logged: %q", logs.String())
	}
}

func TestServeDisabledReturnsAtOnce(t *testing.T) {
	t.Parallel()
	New().Serve(context.Background(), "", discardLogger())
}
