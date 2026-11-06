// Package status polls the fixed public project health endpoints. It exposes
// only names, availability, response time and the age of the last failure to
// the web layer.
package status

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/DryWaters/bitofbytes/models"
)

const (
	pollInterval = time.Minute
	checkTimeout = 5 * time.Second
	staleAfter   = 2 * time.Minute
	maxBody      = 4096
	concurrency  = 4
)

// Entry is the complete public status view. Diagnostic details never enter it.
// Latency is the round trip of the latest check and is set only while the
// target is up. HasLastDown reports whether any failed check has been seen
// since start, and LastDownAgo is the age of the most recent one.
type Entry struct {
	Name        string
	Up          bool
	Latency     time.Duration
	HasLastDown bool
	LastDownAgo time.Duration
}

// LatencyText is the response time in milliseconds, or "" when the target is down.
func (e Entry) LatencyText() string {
	if !e.Up {
		return ""
	}
	if ms := e.Latency.Milliseconds(); ms > 0 {
		return strconv.FormatInt(ms, 10) + " ms"
	}
	return "<1 ms"
}

// LastDownText is the age of the most recent failed check, or "" if none was seen.
func (e Entry) LastDownText() string {
	if !e.HasLastDown {
		return ""
	}
	switch ago := e.LastDownAgo; {
	case ago < time.Minute:
		return "Last down just now"
	case ago < time.Hour:
		return "Last down " + strconv.Itoa(int(ago/time.Minute)) + " min ago"
	case ago < 24*time.Hour:
		return "Last down " + strconv.Itoa(int(ago/time.Hour)) + " hr ago"
	default:
		return "Last down " + strconv.Itoa(int(ago/(24*time.Hour))) + " d ago"
	}
}

type target struct{ name, url string }
type result struct {
	up       bool
	checked  time.Time
	latency  time.Duration
	lastDown time.Time // most recent failed check; zero until one is seen
}

// Checker keeps a small in-memory snapshot; constructing it performs no I/O.
type Checker struct {
	targets  []target
	client   *http.Client
	interval time.Duration
	now      func() time.Time
	mu       sync.RWMutex
	results  []result
}

func projectTargets() []target {
	endpoints := map[string]string{
		"carma":      "https://carma.bitofbytes.io/health",
		"noted":      "https://noted.bitofbytes.io/api/health",
		"dined":      "https://dined.bitofbytes.io/health",
		"permitpal":  "https://permitpal.bitofbytes.io/health",
		"learnd":     "https://learnd.bitofbytes.io/health",
		"dejaview":   "https://dejaview.bitofbytes.io/health",
		"bitofbytes": "https://bitofbytes.io/health",
		"anthology":  "https://anthology.bitofbytes.io/api/health",
	}
	targets := make([]target, 0, len(endpoints))
	for _, project := range models.Projects() {
		targets = append(targets, target{name: project.Name, url: endpoints[project.Slug]})
	}
	return targets
}

func New() *Checker { return newChecker(projectTargets()) }

func newChecker(targets []target) *Checker {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	return &Checker{
		targets:  targets,
		results:  make([]result, len(targets)),
		interval: pollInterval,
		now:      time.Now,
		client: &http.Client{
			Transport: transport,
			Timeout:   checkTimeout,
			// A redirect to a login page is not a successful health check.
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
}

// Run is called once by the server. Polls are immediate and then periodic;
// each completes before the next begins, including when a check is slow.
func (c *Checker) Run(ctx context.Context) {
	defer c.client.CloseIdleConnections()
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		c.poll(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (c *Checker) poll(ctx context.Context) {
	var workers sync.WaitGroup
	jobs := make(chan int)
	for range min(concurrency, len(c.targets)) {
		workers.Go(func() {
			for i := range jobs {
				started := c.now()
				up := c.check(ctx, c.targets[i].url)
				checked := c.now()
				c.mu.Lock()
				r := result{up: up, checked: checked, latency: checked.Sub(started), lastDown: c.results[i].lastDown}
				// A check cut short by shutdown is not an outage.
				if !up && ctx.Err() == nil {
					r.lastDown = checked
				}
				c.results[i] = r
				c.mu.Unlock()
			}
		})
	}
	for i := range c.targets {
		select {
		case jobs <- i:
		case <-ctx.Done():
			close(jobs)
			workers.Wait()
			return
		}
	}
	close(jobs)
	workers.Wait()
}

func (c *Checker) check(ctx context.Context, endpoint string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return false
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	// Read at most a small health response, including on failure, and always
	// close it. A truncated, oversized, or unreadable response counts as down.
	n, readErr := io.Copy(io.Discard, io.LimitReader(resp.Body, maxBody+1))
	closeErr := resp.Body.Close()
	return ctx.Err() == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 && n <= maxBody && readErr == nil && closeErr == nil
}

// Snapshot never performs network I/O. Unknown and stale results are down and
// report no latency, since it would describe an old check.
func (c *Checker) Snapshot() []Entry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := c.now()
	entries := make([]Entry, len(c.targets))
	for i, target := range c.targets {
		result := c.results[i]
		age := now.Sub(result.checked)
		up := result.up && !result.checked.IsZero() && age >= 0 && age < staleAfter
		entry := Entry{Name: target.name, Up: up}
		if up {
			entry.Latency = result.latency
		}
		if !result.lastDown.IsZero() {
			entry.HasLastDown = true
			entry.LastDownAgo = max(now.Sub(result.lastDown), 0)
		}
		entries[i] = entry
	}
	return entries
}
