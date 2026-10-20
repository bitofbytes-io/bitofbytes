package main

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestServeShutsDownWhenContextIsDone(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server := &http.Server{Handler: http.NotFoundHandler(), ReadHeaderTimeout: time.Second}
	done := make(chan error, 1)
	go func() { done <- serve(ctx, server, ln, slog.New(slog.NewTextHandler(io.Discard, nil))) }()

	// Wait until the server answers, then ask it to stop.
	for start := time.Now(); ; {
		if resp, err := http.Get("http://" + addr + "/"); err == nil {
			if err := resp.Body.Close(); err != nil {
				t.Fatal(err)
			}
			break
		}
		if time.Since(start) > 5*time.Second {
			t.Fatal("server never started")
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("serve returned %v, want nil after shutdown", err)
		}
	case <-time.After(shutdownTimeout + time.Second):
		t.Fatal("serve did not return after the context was done")
	}
}

func TestServeReturnsListenerError(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	if err := ln.Close(); err != nil {
		t.Fatal(err)
	}

	err = serve(context.Background(), &http.Server{ReadHeaderTimeout: time.Second}, ln, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err == nil {
		t.Fatal("serve returned nil for a listener that cannot accept")
	}
}
