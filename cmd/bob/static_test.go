package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DryWaters/bitofbytes/internal/metrics"
	"github.com/DryWaters/bitofbytes/internal/status"
	"github.com/DryWaters/bitofbytes/models"
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
	for _, want := range []string{"/static/daniel-waters-resume.pdf", "/static/styles.css?v=" + testAssetVersion, "/static/nocturne.js?v=" + testAssetVersion, "/static/screenshot-viewer.js?v=" + testAssetVersion} {
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

var (
	rootBlock  = regexp.MustCompile(`(?s)\n:root \{(.*?)\n\}`)
	lightBlock = regexp.MustCompile(`(?s)@media \(prefers-color-scheme: light\) \{\s*:root \{(.*?)\n  \}`)
	colorToken = regexp.MustCompile(`(?m)^\s*(--[\w-]+):\s*(#[0-9A-Fa-f]{3,8}|transparent)\s*;`)
)

// The light scheme overrides every colour token of the default (dark) one, so
// no role keeps a colour picked for the other background, in both the served
// stylesheet and the design-system copy of the tokens.
func TestLightSchemeOverridesEveryColourToken(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"../../static/styles.css", "../../docs/design/nocturne/tokens.css"} {
		css, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			t.Fatal(err)
		}
		root, light := rootBlock.FindSubmatch(css), lightBlock.FindSubmatch(css)
		if root == nil || light == nil {
			t.Fatalf("%s: missing :root block or prefers-color-scheme: light block", path)
		}
		overridden := map[string]bool{}
		for _, m := range colorToken.FindAllSubmatch(light[1], -1) {
			overridden[string(m[1])] = true
		}
		tokens := colorToken.FindAllSubmatch(root[1], -1)
		if len(tokens) == 0 {
			t.Fatalf("%s: no colour tokens in :root", path)
		}
		for _, m := range tokens {
			if !overridden[string(m[1])] {
				t.Errorf("%s: %s has no prefers-color-scheme: light override", path, m[1])
			}
		}
		if !strings.Contains(string(light[1]), "color-scheme: light;") {
			t.Errorf("%s: light block does not set color-scheme: light", path)
		}
	}
}

var customProp = regexp.MustCompile(`(?m)^\s*(--[\w-]+):\s*([^;]+?)\s*;`)

// normaliseToken makes token values comparable across files: one space between
// words and upper-case hex digits.
func normaliseToken(v string) string {
	return strings.ToUpper(strings.Join(strings.Fields(v), " "))
}

// lightTokens returns the custom properties of the light scheme in a stylesheet.
func lightTokens(t *testing.T, path string) map[string]string {
	t.Helper()
	css, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatal(err)
	}
	light := lightBlock.FindSubmatch(css)
	if light == nil {
		t.Fatalf("%s: missing prefers-color-scheme: light block", path)
	}
	tokens := map[string]string{}
	for _, m := range customProp.FindAllSubmatch(light[1], -1) {
		tokens[string(m[1])] = normaliseToken(string(m[2]))
	}
	if len(tokens) == 0 {
		t.Fatalf("%s: no light tokens found", path)
	}
	return tokens
}

type designToken struct {
	Name  string `json:"name"`
	Light string `json:"light"`
}

// lightTokensFromJSON returns the light value of every colour and shadow token
// in tokens.json that has one.
func lightTokensFromJSON(t *testing.T, path string) map[string]string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Color  struct{ Tokens []designToken } `json:"color"`
		Shadow struct{ Tokens []designToken } `json:"shadow"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	tokens := map[string]string{}
	for _, tok := range append(doc.Color.Tokens, doc.Shadow.Tokens...) {
		if tok.Light != "" {
			tokens["--"+tok.Name] = normaliseToken(tok.Light)
		}
	}
	if len(tokens) == 0 {
		t.Fatalf("%s: no light values found", path)
	}
	return tokens
}

// The Day values live in the served stylesheet, the design system's tokens.css
// and its tokens.json. All three must define the same light tokens with the
// same values, so editing one without the others fails.
func TestLightSchemeTokensMatchDesignSystem(t *testing.T) {
	t.Parallel()
	const designDir = "../../docs/design/nocturne/"
	sources := []struct {
		name   string
		tokens map[string]string
	}{
		{"static/styles.css", lightTokens(t, "../../static/styles.css")},
		{"tokens.css", lightTokens(t, designDir+"tokens.css")},
		{"tokens.json", lightTokensFromJSON(t, designDir+"tokens.json")},
	}

	ref := sources[0]
	for _, other := range sources[1:] {
		for name, want := range ref.tokens {
			got, ok := other.tokens[name]
			switch {
			case !ok:
				t.Errorf("light %s is in %s but missing from %s", name, ref.name, other.name)
			case got != want:
				t.Errorf("light %s is %s in %s but %s in %s", name, want, ref.name, got, other.name)
			}
		}
		for name := range other.tokens {
			if _, ok := ref.tokens[name]; !ok {
				t.Errorf("light %s is in %s but missing from %s", name, other.name, ref.name)
			}
		}
	}
}

func TestPagesDeclareThemeColorPerScheme(t *testing.T) {
	t.Parallel()
	rr := httptest.NewRecorder()
	newTestHandler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	body := rr.Body.String()
	for _, want := range []string{
		`<meta name="theme-color" content="#0A1224" media="(prefers-color-scheme: dark)">`,
		`<meta name="theme-color" content="#F6F8FC" media="(prefers-color-scheme: light)">`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("home page is missing %s", want)
		}
	}
}

// A development build has no asset version: URLs carry no ?v= and nothing is
// cached as immutable, so edits show up on reload.
func TestDevelopmentAssetsAreNotVersioned(t *testing.T) {
	t.Parallel()
	var cfg models.Config
	handler := newHandler(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), metrics.New(), "../../static", "", status.New())

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	body := rr.Body.String()
	if !strings.Contains(body, `href="/static/styles.css"`) || !strings.Contains(body, `src="/static/nocturne.js"`) {
		t.Error("development pages should link CSS and JS without ?v=")
	}

	for _, path := range []string{"/static/styles.css", "/static/styles.css?v=", "/static/nocturne.js"} {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if got := rr.Header().Get("Cache-Control"); got != "no-cache" {
			t.Errorf("%s Cache-Control = %q, want no-cache", path, got)
		}
	}
}
