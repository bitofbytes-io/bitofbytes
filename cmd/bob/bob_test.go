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
	cfg.CSRF.Key = []byte("01234567890123456789012345678901")
	cfg.CSRF.Secure = false

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return newHandlerWithStaticDir(cfg, logger, "../../static")
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
			want := []string{project.Name, project.Tagline, project.RepoURL, project.LiveURL, project.LastUpdate, project.Notes}
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
			if strings.Contains(body, "atomic-handheld.js") {
				t.Error("project detail loads homepage-only renderer")
			}
		})
	}
}

func TestHomepageHandheldAndContact(t *testing.T) {
	t.Parallel()
	handler := newTestHandler()
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	body := html.UnescapeString(rr.Body.String())
	for _, want := range []string{`src="/static/atomic-handheld.js"`, `type="module"`, `id="contact"`, `href="mailto:daniel@bitofbytes.io"`, `href="https://www.linkedin.com/in/daniel-waters/"`, `href="https://github.com/bitofbytes-io"`, `class="screen-prev" disabled`, `class="screen-next" disabled`} {
		if !strings.Contains(body, want) {
			t.Errorf("missing homepage element %q", want)
		}
	}
	for _, project := range models.Projects() {
		for _, want := range []string{`data-slug="` + project.Slug + `"`, `data-name="` + project.Name + `"`, `data-summary="` + project.Summary + `"`} {
			if !strings.Contains(body, want) {
				t.Errorf("missing handheld data %q", want)
			}
		}
	}
	if strings.Contains(body, "Say hello") || strings.Contains(body, "<footer") {
		t.Error("homepage restored removed contact/footer content")
	}
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/projects", nil))
	if strings.Contains(rr.Body.String(), "atomic-handheld.js") {
		t.Error("project index loads homepage-only renderer")
	}
}
