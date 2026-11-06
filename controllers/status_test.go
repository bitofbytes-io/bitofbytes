package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DryWaters/bitofbytes/internal/status"
	"github.com/DryWaters/bitofbytes/templates"
	"github.com/DryWaters/bitofbytes/views"
)

func TestStatusRendersOnlyPublicSnapshot(t *testing.T) {
	calls := 0
	handler := Status{
		Snapshot: func() []status.Entry {
			calls++
			return []status.Entry{{Name: "Example", Up: true}, {Name: "Other", Up: false}}
		},
		Page: views.Must(views.ParseFS("testrev", templates.FS, "status/index.gohtml", "base.gohtml")),
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/status?target=http://internal:9000/health", nil))
	if rr.Code != 200 || calls != 1 || rr.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("status=%d calls=%d cache=%s", rr.Code, calls, rr.Header().Get("Cache-Control"))
	}
	body := rr.Body.String()
	for _, want := range []string{"<dt>Example</dt>", "<dd>Up</dd>", "<dt>Other</dt>", "<dd>Down</dd>", `href="/status" aria-current="page"`} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %s", want)
		}
	}
	for _, private := range []string{"internal:9000", "/health", "degraded", "checked", "latency", "endpoint", "error"} {
		if strings.Contains(body, private) {
			t.Errorf("leaked %s", private)
		}
	}
}

func TestStatusRendersResponseTimeAndLastDown(t *testing.T) {
	handler := Status{
		Snapshot: func() []status.Entry {
			return []status.Entry{
				{Name: "Fast", Up: true, Latency: 142 * time.Millisecond},
				{Name: "Flaky", Up: true, Latency: 9 * time.Millisecond, HasLastDown: true, LastDownAgo: 5 * time.Minute},
				{Name: "Broken", Up: false, HasLastDown: true, LastDownAgo: 30 * time.Second},
			}
		},
		Page: views.Must(views.ParseFS("testrev", templates.FS, "status/index.gohtml", "base.gohtml")),
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/status", nil))
	body := rr.Body.String()
	for _, want := range []string{
		`<dd class="nc-status__meta">142 ms</dd>`,
		`<dd class="nc-status__meta">9 ms · Last down 5 min ago</dd>`,
		`<dd class="nc-status__meta">Last down just now</dd>`,
	} {
		if strings.Count(body, want) != 1 {
			t.Errorf("missing %s", want)
		}
	}
	if strings.Count(body, "nc-status__meta") != 3 {
		t.Errorf("meta rows=%d want 3", strings.Count(body, "nc-status__meta"))
	}
}
