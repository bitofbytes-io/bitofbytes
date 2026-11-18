package status

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRingOverwritesOldestWhenFull(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	r := newRing(3)
	if up, total := r.count(now, time.Hour); up != 0 || total != 0 {
		t.Fatalf("empty ring: up=%d total=%d", up, total)
	}
	// Three failures fill the ring, then two successes replace the oldest two.
	for i := range 3 {
		r.add(now.Add(time.Duration(i-5)*time.Minute), false)
	}
	r.add(now.Add(-2*time.Minute), true)
	r.add(now.Add(-time.Minute), true)
	if up, total := r.count(now, time.Hour); up != 2 || total != 3 {
		t.Fatalf("full ring: up=%d total=%d want 2 of 3", up, total)
	}
	if got := len(r.samples); got != 3 {
		t.Fatalf("ring grew to %d", got)
	}
}

func TestRingCountWindow(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	r := newRing(10)
	r.add(now.Add(-25*time.Hour), false)            // outside the window
	r.add(now.Add(-uptimeWindow), false)            // exactly on the cutoff: outside
	r.add(now.Add(-uptimeWindow+time.Second), true) // just inside
	r.add(now.Add(-time.Hour), false)
	r.add(now, true)                  // taken now: inside
	r.add(now.Add(time.Second), true) // clock went backwards: ignored
	if up, total := r.count(now, uptimeWindow); up != 2 || total != 3 {
		t.Fatalf("up=%d total=%d want 2 of 3", up, total)
	}
}

func TestRingZeroCapacityIgnoresSamples(t *testing.T) {
	r := newRing(0)
	r.add(time.Now(), true)
	if up, total := r.count(time.Now(), time.Hour); up != 0 || total != 0 {
		t.Fatalf("up=%d total=%d", up, total)
	}
}

func TestUptimePermille(t *testing.T) {
	for _, tt := range []struct {
		name      string
		up, total int
		want      int
		ok        bool
	}{
		{"no data", 0, 0, 0, false},
		{"too few samples", 9, 9, 0, false},
		{"minimum all up", 10, 10, 1000, true},
		{"minimum all down", 0, 10, 0, true},
		{"rounds down", 1439, 1440, 999, true},
		{"one failure never reads 100", 9999, 10000, 999, true},
		{"two thirds", 20, 30, 666, true},
		{"half", 720, 1440, 500, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := uptimePermille(tt.up, tt.total)
			if got != tt.want || ok != tt.ok {
				t.Errorf("got %d,%v want %d,%v", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestEntryUptimeAndMetaText(t *testing.T) {
	for _, tt := range []struct {
		name  string
		entry Entry
		want  string
	}{
		{"collecting", Entry{}, "Uptime collecting"},
		{"perfect", Entry{HasUptime: true, Uptime24h: 1000}, "Uptime 100.0% (24 h)"},
		{"fraction", Entry{HasUptime: true, Uptime24h: 999}, "Uptime 99.9% (24 h)"},
		{"single digit", Entry{HasUptime: true, Uptime24h: 50}, "Uptime 5.0% (24 h)"},
		{"zero", Entry{HasUptime: true}, "Uptime 0.0% (24 h)"},
	} {
		if got := tt.entry.UptimeText(); got != tt.want {
			t.Errorf("%s: %q want %q", tt.name, got, tt.want)
		}
	}
	full := Entry{
		Up: true, Latency: 12 * time.Millisecond, HasUptime: true, Uptime24h: 995,
		HasLastDown: true, LastDownAgo: 5 * time.Minute,
	}
	if got, want := full.MetaText(), "12 ms · Uptime 99.5% (24 h) · Last down 5 min ago"; got != want {
		t.Errorf("meta %q want %q", got, want)
	}
	if got, want := (Entry{}).MetaText(), "Uptime collecting"; got != want {
		t.Errorf("down, no data: %q want %q", got, want)
	}
}

func TestSnapshotUptimeCollectsThenReports(t *testing.T) {
	c := newChecker([]target{{"Example", "https://example.test/health"}})
	now := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	c.now = func() time.Time { return now }
	code := 200
	c.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader("ok"))}, nil
	})

	// A fresh checker, as after a restart, has no history.
	if e := c.Snapshot()[0]; e.HasUptime || e.UptimeText() != "Uptime collecting" {
		t.Fatalf("no data: %+v", e)
	}
	for range minUptimeSamples - 1 {
		c.poll(context.Background())
		now = now.Add(time.Minute)
	}
	if e := c.Snapshot()[0]; e.HasUptime {
		t.Fatalf("still collecting after %d samples: %+v", minUptimeSamples-1, e)
	}

	// The tenth sample, a failure, unlocks the percentage: 9 up of 10.
	code = 503
	c.poll(context.Background())
	e := c.Snapshot()[0]
	if !e.HasUptime || e.Uptime24h != 900 || e.UptimeText() != "Uptime 90.0% (24 h)" {
		t.Fatalf("9 of 10: %+v", e)
	}

	// Samples age out of the 24 hour window; with too few left it collects again.
	now = now.Add(uptimeWindow - time.Minute)
	if e := c.Snapshot()[0]; e.HasUptime {
		t.Fatalf("aged-out samples still counted: %+v", e)
	}
}

func TestCanceledCheckIsNotRecordedAsSample(t *testing.T) {
	c := newChecker([]target{{"Example", "https://example.test/health"}})
	c.client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, r.Context().Err()
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c.history[0].add(c.now(), true) // keep the ring non-empty to see it unchanged
	c.poll(ctx)
	if _, total := c.history[0].count(c.now().Add(time.Second), uptimeWindow); total != 1 {
		t.Fatalf("total=%d; a shutdown-canceled check must not add a sample", total)
	}
}

func TestHistoryCapacityCoversWindow(t *testing.T) {
	if historyCapacity != 1440 {
		t.Fatalf("capacity=%d want one sample per minute for 24 h", historyCapacity)
	}
}
