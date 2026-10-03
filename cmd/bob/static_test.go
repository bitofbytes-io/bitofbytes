package main

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

var staticRef = regexp.MustCompile(`(?:href|src)="(/static/[^"]+)"`)

// Every /static/ file a page links to, including the resume, screenshots,
// icons, CSS and JS, is served. CSS and JS carry this release's version.
func TestPagesLinkOnlyToServedStaticFiles(t *testing.T) {
	t.Parallel()
	handler := newTestHandler()

	refs := map[string]string{}
	for _, page := range pagePaths() {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, page, nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("%s status = %d", page, rr.Code)
		}
		for _, m := range staticRef.FindAllStringSubmatch(rr.Body.String(), -1) {
			refs[strings.ReplaceAll(m[1], "&amp;", "&")] = page
		}
	}
	for _, want := range []string{"/static/daniel-waters-resume.pdf", "/static/styles.css?v=" + testAssetVersion, "/static/nocturne.js?v=" + testAssetVersion} {
		if _, ok := refs[want]; !ok {
			t.Errorf("no page links to %s", want)
		}
	}

	for ref, page := range refs {
		if (strings.Contains(ref, ".css") || strings.Contains(ref, ".js")) && !strings.HasSuffix(ref, "?v="+testAssetVersion) {
			t.Errorf("%s links to %s without ?v=%s", page, ref, testAssetVersion)
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, ref, nil))
		if rr.Code != http.StatusOK || rr.Body.Len() == 0 {
			t.Errorf("%s (linked from %s) status = %d, %d bytes", ref, page, rr.Code, rr.Body.Len())
		}
	}
}

func TestStaticDirectoriesAreNotListed(t *testing.T) {
	t.Parallel()
	handler := newTestHandler()

	for _, path := range []string{"/static/", "/static/projects/", "/static/projects", "/static/projects/noted/"} {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != http.StatusNotFound {
			t.Errorf("%s status = %d, want %d", path, rr.Code, http.StatusNotFound)
		}
		if strings.Contains(rr.Body.String(), "<a href=") {
			t.Errorf("%s lists its contents", path)
		}
	}
}

func TestStaticCacheControl(t *testing.T) {
	t.Parallel()
	handler := newTestHandler()

	for _, tt := range []struct{ path, want string }{
		{"/static/styles.css?v=" + testAssetVersion, "public, max-age=31536000, immutable"},
		{"/static/nocturne.js?v=" + testAssetVersion, "public, max-age=31536000, immutable"},
		{"/static/styles.css", "no-cache"},
		{"/static/styles.css?v=otherrelease", "no-cache"},
		{"/static/projects/noted/local-library.webp", "no-cache"},
		{"/favicon.ico", "no-cache"},
	} {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if rr.Code != http.StatusOK {
			t.Errorf("%s status = %d", tt.path, rr.Code)
		}
		if got := rr.Header().Get("Cache-Control"); got != tt.want {
			t.Errorf("%s Cache-Control = %q, want %q", tt.path, got, tt.want)
		}
	}
}
