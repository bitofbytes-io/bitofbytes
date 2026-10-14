package main

import (
	"html"
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
	cfg.CSRF.Secure = false

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return newHandler(cfg, logger, "../../static", testAssetVersion)
}

const testAssetVersion = "testrev"

func TestRoutesRenderCurrentSiteSurface(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()

	tests := []struct {
		path string
		want string
	}{
		{path: "/", want: "Daniel Waters"},
		{path: "/projects", want: "Recently updated"},
		{path: "/projects?sort=newest", want: "Newest first"},
		{path: "/projects/carma", want: "Carma"},
		{path: "/projects/noted", want: "Noted"},
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

func TestHandlerRejectsCrossSitePost(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("cross-site POST status code = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestProjectDetailTimelinesShowUpdatesWithoutFirstCommit(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()
	for _, project := range models.Projects() {
		req := httptest.NewRequest(http.MethodGet, "/projects/"+project.Slug, nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("%s status = %d", project.Slug, rr.Code)
			continue
		}
		body := rr.Body.String()
		if got, want := strings.Count(body, `class="nc-timeline__item`), len(project.TimelineUpdates()); got != want {
			t.Errorf("%s timeline items = %d, want %d", project.Slug, got, want)
		}
		if strings.Contains(body, "First commit.") {
			t.Errorf("%s still shows First commit", project.Slug)
		}
	}
}

// Every screenshot in the catalog must exist and be served as a WebP.
func TestProjectScreenshotsServeAsWebP(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()
	count := 0
	for _, project := range models.Projects() {
		for _, shot := range project.Screenshots {
			if shot.Path == "" {
				continue
			}
			count++
			t.Run(shot.Path, func(t *testing.T) {
				if !strings.HasSuffix(shot.Path, ".webp") {
					t.Fatalf("path %q is not a .webp", shot.Path)
				}
				rr := httptest.NewRecorder()
				handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, shot.Path, nil))
				if rr.Code != http.StatusOK {
					t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
				}
				if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "image/webp") {
					t.Fatalf("Content-Type = %q, want image/webp", got)
				}
			})
		}
	}
	if count == 0 {
		t.Fatal("no project screenshots in the catalog")
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
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("status code = %d, want %d", rr.Code, http.StatusOK)
			}
			contentType := rr.Header().Get("Content-Type")
			contentTypeMatches := strings.HasPrefix(contentType, tt.wantContentType)
			if tt.wantContentType == "image/x-icon" {
				contentTypeMatches = contentTypeMatches || strings.HasPrefix(contentType, "image/vnd.microsoft.icon")
			}
			if !contentTypeMatches {
				t.Errorf("Content-Type = %q, want prefix %q", contentType, tt.wantContentType)
			}
			if rr.Body.Len() == 0 {
				t.Errorf("served an empty body")
			}
		})
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

// Every page must carry a unique, per-page title plus search and social meta
// with absolute canonical and og: URLs (scrapers reject relative ones).
func TestPagesHaveUniqueTitleAndMeta(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()

	homeWant := map[string]string{
		`<title>BitOfBytes | Daniel Waters</title>`:                                                           "home title",
		`<meta name="description" content="The personal landing page and project portfolio for Daniel Waters`: "home description",
		`<link rel="canonical" href="https://bitofbytes.io/">`:                                                "home canonical",
		`<meta property="og:title" content="BitOfBytes | Daniel Waters">`:                                     "home og:title",
		`<meta property="og:url" content="https://bitofbytes.io/">`:                                           "home og:url",
		`<meta name="twitter:card" content="summary_large_image">`:                                            "twitter card",
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("home status = %d", rr.Code)
	}
	for want, what := range homeWant {
		if !strings.Contains(rr.Body.String(), want) {
			t.Errorf("home missing %s: %q", what, want)
		}
	}

	// Every project detail page: "<Name> | BitOfBytes" title, Tagline
	// description, canonical + og:image pointing at the first screenshot.
	titles := map[string]bool{}
	for _, project := range models.Projects() {
		t.Run(project.Slug, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/projects/"+project.Slug, nil))
			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d", rr.Code)
			}
			// Unescape like the other content tests: html/template escapes the
			// apostrophe in "Learn'd" as &#39; in the raw body.
			body := html.UnescapeString(rr.Body.String())
			wantTitle := "<title>" + project.Name + " | BitOfBytes</title>"
			if !strings.Contains(body, wantTitle) {
				t.Errorf("missing %q", wantTitle)
			}
			if !strings.Contains(body, `<meta name="description" content="`+project.Tagline+`">`) {
				t.Errorf("missing tagline meta description")
			}
			if !strings.Contains(body, `<link rel="canonical" href="https://bitofbytes.io/projects/`+project.Slug+`">`) {
				t.Errorf("missing per-page canonical")
			}
			if thumb := project.Thumbnail(); thumb.Path != "" {
				abs := "https://bitofbytes.io" + thumb.Path
				if !strings.Contains(body, `<meta property="og:image" content="`+abs+`">`) {
					t.Errorf("missing absolute og:image %q", abs)
				}
			}
			if !strings.Contains(body, `"@type": "SoftwareApplication"`) {
				t.Errorf("missing SoftwareApplication JSON-LD data block")
			}
		})
		titles[project.Name+" | BitOfBytes"] = true
	}
	if len(titles) != len(models.Projects()) {
		t.Errorf("detail titles not unique: got %d for %d projects", len(titles), len(models.Projects()))
	}
}

// The sitemap must be valid XML and list the home page, the projects index,
// and every project slug.
func TestSitemapListsAllProjectPages(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("sitemap status = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/xml") {
		t.Errorf("Content-Type = %q, want application/xml", got)
	}

	body := rr.Body.String()
	for _, want := range []string{"https://bitofbytes.io/", "https://bitofbytes.io/projects"} {
		if !strings.Contains(body, want) {
			t.Errorf("sitemap missing %q", want)
		}
	}
	for _, project := range models.Projects() {
		if !strings.Contains(body, "<loc>https://bitofbytes.io/projects/"+project.Slug+"</loc>") {
			t.Errorf("sitemap missing project %q", project.Slug)
		}
	}
}

// robots.txt must allow crawling and point at the sitemap.
func TestRobotsTXT(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("robots.txt status = %d", rr.Code)
	}
	body := rr.Body.String()
	for _, want := range []string{"User-agent: *", "Allow: /", "Sitemap: https://bitofbytes.io/sitemap.xml"} {
		if !strings.Contains(body, want) {
			t.Errorf("robots.txt missing %q", want)
		}
	}
}

func TestPortfolioPreservesProjectContentAndScreenshots(t *testing.T) {
	t.Parallel()
	handler := newTestHandler()
	for _, project := range models.Projects() {
		t.Run(project.Slug, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/projects/"+project.Slug, nil))
			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d", rr.Code)
			}
			body := html.UnescapeString(rr.Body.String())
			want := []string{project.Name, project.Tagline, project.RepoURL, project.LiveURL, project.UpdatedLong(), project.StartedLong(), project.LatestNote()}
			want = append(want, project.Paragraphs...)
			want = append(want, project.Tech...)
			want = append(want, project.Highlights...)
			for _, screenshot := range project.Screenshots {
				want = append(want, screenshot.Title, screenshot.Alt, screenshot.Note)
				if screenshot.Path != "" {
					want = append(want, `href="`+screenshot.Path+`"`, `src="`+screenshot.Path+`"`)
				}
			}
			for _, text := range want {
				if !strings.Contains(body, text) {
					t.Errorf("missing content %q", text)
				}
			}
			if strings.Contains(body, "nocturne.js") {
				t.Error("project detail loads the homepage-only keyboard script")
			}
		})
	}
}

func TestHomepageKeyboardAndContact(t *testing.T) {
	t.Parallel()
	handler := newTestHandler()
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("home status = %d, want %d", rr.Code, http.StatusOK)
	}
	body := html.UnescapeString(rr.Body.String())
	for _, want := range []string{
		`src="/static/nocturne.js?v=` + testAssetVersion + `"`,
		`data-nc-octave`,
		`data-nc-sound aria-pressed="true"`,
		`id="contact"`,
		`Say hello.`,
		`href="mailto:daniel@bitofbytes.io"`,
		`href="https://www.linkedin.com/in/daniel-waters/"`,
		`href="https://github.com/bitofbytes-io"`,
		`href="/projects/noted"`,
		`href="/projects/anthology"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing homepage element %q", want)
		}
	}
	onKeys := models.SortByLastUpdate(models.Projects())
	onKeys = onKeys[:min(len(onKeys), 8)]
	if got, want := strings.Count(body, `<a class="nc-key`), len(onKeys); got != want {
		t.Errorf("piano keys = %d, want %d", got, want)
	}
	for _, project := range onKeys {
		if want := `href="/projects/` + project.Slug + `" data-note=`; !strings.Contains(body, want) {
			t.Errorf("missing piano key for %q", project.Slug)
		}
	}
	for _, removed := range []string{"atomic-handheld.js", "three-r160", "data-scene"} {
		if strings.Contains(body, removed) {
			t.Errorf("homepage still references the retired handheld: %q", removed)
		}
	}

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/projects", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("project index status = %d, want %d", rr.Code, http.StatusOK)
	}
	if strings.Contains(rr.Body.String(), "nocturne.js") {
		t.Error("project index loads the homepage-only keyboard script")
	}
}

func TestStaticAssetsServeKeyboardScriptAndRetireHandheld(t *testing.T) {
	t.Parallel()
	handler := newTestHandler()

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/static/nocturne.js", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("nocturne.js status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "window.Nocturne") {
		t.Error("nocturne.js does not define window.Nocturne")
	}

	for _, path := range []string{"/static/atomic-handheld.js", "/static/vendor/three-r160.module.js"} {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != http.StatusNotFound {
			t.Errorf("%s status = %d, want %d", path, rr.Code, http.StatusNotFound)
		}
	}
}
