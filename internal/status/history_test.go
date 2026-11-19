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
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	c.now = func() time.Time { return now }
	started := make(chan struct{})
	// The test context alone controls cancellation; a client timeout expiring
	// during a stall would make both channels ready and blame the transport.
	c.client.Timeout = 0
	c.client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		close(started)
		<-r.Context().Done()
		return nil, r.Context().Err()
	})
	// Enough healthy history to report a percentage, so a recorded failure would move it.
	for i := range minUptimeSamples {
		c.history[0].add(now.Add(time.Duration(i-minUptimeSamples)*time.Minute), true)
	}
	before := c.Snapshot()[0]
	if !before.HasUptime || before.Uptime24h != 1000 {
		t.Fatalf("seeded history: %+v", before)
	}

	// Cancel while the check is in flight, as on shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.poll(ctx)
	}()
	select {
	case <-started:
	case <-done:
		t.Fatal("poll returned without invoking the transport")
	}
	cancel()
	<-done

	after := c.Snapshot()[0]
	if !after.HasUptime || after.Uptime24h != before.Uptime24h {
		t.Fatalf("uptime %+v became %+v; a shutdown-canceled check must not add a sample", before, after)
	}
	if after.HasLastDown {
		t.Fatalf("a shutdown-canceled check was recorded as an outage: %+v", after)
	}
}

func TestHistoryKeepsEverySampleInWindowWithJitter(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	r := newRing(historyCapacity)
	// An early failure, then a full day of polls whose completions drift a few
	// seconds earlier than one per minute, so 1441 samples fall in the window.
	start := now.Add(-uptimeWindow + 4*time.Second)
	r.add(start, false)
	const polls = 1440
	for i := 1; i <= polls; i++ {
		r.add(start.Add(time.Duration(i)*(uptimeWindow-4*time.Second)/polls), true)
	}
	up, total := r.count(now, uptimeWindow)
	if total != polls+1 || up != polls {
		t.Fatalf("up=%d total=%d want %d of %d: the early failure was dropped", up, total, polls, polls+1)
	}
	if permille, ok := uptimePermille(up, total); !ok || permille != 999 {
		t.Fatalf("permille=%d ok=%v want 999", permille, ok)
	}
}
