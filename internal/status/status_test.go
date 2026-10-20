package status

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DryWaters/bitofbytes/models"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type trackedBody struct {
	reader   io.Reader
	read     int
	closed   bool
	closeErr error
}

func (b *trackedBody) Read(p []byte) (int, error) {
	n, err := b.reader.Read(p)
	b.read += n
	return n, err
}
func (b *trackedBody) Close() error { b.closed = true; return b.closeErr }

type failedReader struct{}

func (failedReader) Read([]byte) (int, error) { return 0, errors.New("internal diagnostic") }

func TestTargetsCoverEveryProject(t *testing.T) {
	targets := projectTargets()
	projects := models.Projects()
	if len(targets) != 8 || len(targets) != len(projects) {
		t.Fatalf("targets=%d projects=%d", len(targets), len(projects))
	}
	seen := map[string]bool{}
	for i, target := range targets {
		if target.name != projects[i].Name || seen[target.name] {
			t.Errorf("missing or duplicate project: %q", target.name)
		}
		seen[target.name] = true
		u, err := url.Parse(target.url)
		if err != nil {
			t.Fatal(err)
		}
		host := projects[i].Slug + ".bitofbytes.io"
		if projects[i].Slug == "bitofbytes" {
			host = "bitofbytes.io"
		}
		path := "/health"
		if projects[i].Slug == "noted" || projects[i].Slug == "anthology" {
			path = "/api/health"
		}
		if u.Scheme != "https" || u.Host != host || u.Path != path || u.User != nil || u.RawQuery != "" {
			t.Errorf("unexpected endpoint for %s: %s", target.name, target.url)
		}
	}
}

func TestCheckResponseAndBodyBounds(t *testing.T) {
	for _, tt := range []struct {
		name     string
		code     int
		reader   io.Reader
		closeErr error
		want     bool
	}{
		{"healthy", 200, strings.NewReader("ok"), nil, true},
		{"empty success", 204, strings.NewReader(""), nil, true},
		{"failure", 503, strings.NewReader("internal diagnostic"), nil, false},
		{"unauthorized", 401, strings.NewReader("login"), nil, false},
		{"oversized", 200, strings.NewReader(strings.Repeat("x", maxBody*2)), nil, false},
		{"read error", 200, failedReader{}, nil, false},
		{"close error", 200, strings.NewReader("ok"), errors.New("close"), false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			body := &trackedBody{reader: tt.reader, closeErr: tt.closeErr}
			c := newChecker(nil)
			c.client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tt.code, Body: body, Header: make(http.Header)}, nil
			})
			if got := c.check(context.Background(), "https://example.test/health"); got != tt.want {
				t.Errorf("up=%v want=%v", got, tt.want)
			}
			if !body.closed || body.read > maxBody+1 {
				t.Errorf("body closed=%v read=%d", body.closed, body.read)
			}
		})
	}
}

func TestCheckRejectsRedirectsAndTimeouts(t *testing.T) {
	var loginCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			http.Redirect(w, r, "/login", http.StatusFound)
		case "/login":
			loginCalls.Add(1)
		case "/slow":
			<-r.Context().Done()
		}
	}))
	defer server.Close()
	c := newChecker(nil)
	defer c.client.CloseIdleConnections()
	if c.check(context.Background(), server.URL+"/health") || loginCalls.Load() != 0 {
		t.Fatal("redirect was followed or considered up")
	}
	c.client.Timeout = 20 * time.Millisecond
	if c.check(context.Background(), server.URL+"/slow") {
		t.Fatal("timeout considered up")
	}
	c.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New("connection failure") })
	if c.check(context.Background(), "https://example.test/health") {
		t.Fatal("network failure considered up")
	}
}

func TestSnapshotUnknownStaleAndRecovery(t *testing.T) {
	c := newChecker([]target{{"Example", "https://example.test/health"}})
	now := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	c.now = func() time.Time { return now }
	code := 200
	c.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader("ok"))}, nil
	})
	if c.Snapshot()[0].Up {
		t.Fatal("unknown is up")
	}
	c.poll(context.Background())
	snapshot := c.Snapshot()
	if !snapshot[0].Up {
		t.Fatal("success is down")
	}
	snapshot[0].Name = "changed"
	if c.Snapshot()[0].Name != "Example" {
		t.Fatal("snapshot aliases state")
	}
	now = now.Add(staleAfter)
	if c.Snapshot()[0].Up {
		t.Fatal("stale success is up")
	}
	code = 503
	c.poll(context.Background())
	if c.Snapshot()[0].Up {
		t.Fatal("failed check is up")
	}
	code = 200
	c.poll(context.Background())
	if !c.Snapshot()[0].Up {
		t.Fatal("did not recover")
	}
}

func TestSnapshotDoesNotMarkConcurrentSuccessDown(t *testing.T) {
	c := newChecker([]target{{"Example", "https://example.test/health"}})
	now := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	c.results[0] = result{up: true, checked: now.Add(-time.Second)}
	c.now = func() time.Time {
		// Model a check finishing just after the snapshot samples its clock.
		// It may publish only if the snapshot has not locked the results yet.
		// TryLock makes this interleaving deterministic without timed sleeps.
		if c.mu.TryLock() {
			c.results[0] = result{up: true, checked: now.Add(time.Nanosecond)}
			c.mu.Unlock()
		}
		return now
	}
	if !c.Snapshot()[0].Up {
		t.Fatal("a concurrent successful check was incorrectly reported down")
	}
}

func TestRunImmediateBoundedNonBlockingAndCanceled(t *testing.T) {
	c := newChecker(projectTargets())
	c.interval = time.Millisecond
	var active, maximum, calls atomic.Int32
	started := make(chan struct{}, concurrency)
	c.client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		n := active.Add(1)
		defer active.Add(-1)
		for old := maximum.Load(); n > old; old = maximum.Load() {
			if maximum.CompareAndSwap(old, n) {
				break
			}
		}
		started <- struct{}{}
		<-r.Context().Done()
		return nil, r.Context().Err()
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { defer close(done); c.Run(ctx) }()
	for range concurrency {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("initial checks did not start")
		}
	}
	snapshots := make(chan struct{})
	go func() {
		defer close(snapshots)
		for range 100 {
			c.Snapshot()
		}
	}()
	select {
	case <-snapshots:
	case <-time.After(time.Second):
		t.Fatal("network I/O holds snapshot lock")
	}
	time.Sleep(10 * time.Millisecond)
	if calls.Load() != concurrency || maximum.Load() > concurrency {
		t.Fatal("polls overlapped or concurrency exceeded")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("poller did not stop")
	}
	if active.Load() != 0 {
		t.Fatal("checks still running")
	}
	for _, entry := range c.Snapshot() {
		if entry.Up {
			t.Fatal("canceled check is up")
		}
	}
}

func TestRepeatedPollingAndConcurrentSnapshots(t *testing.T) {
	c := newChecker(projectTargets())
	c.interval = time.Millisecond
	var calls atomic.Int32
	c.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("ok"))}, nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { defer close(done); c.Run(ctx) }()
	var readers sync.WaitGroup
	for range 4 {
		readers.Go(func() {
			for range 1000 {
				c.Snapshot()
			}
		})
	}
	readers.Wait()
	deadline := time.Now().Add(time.Second)
	for calls.Load() < 16 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	cancel()
	<-done
	if calls.Load() < 16 {
		t.Fatal("poller did not repeat")
	}
}
