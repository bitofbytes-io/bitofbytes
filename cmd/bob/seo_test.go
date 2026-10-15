package main

import (
	"encoding/json"
	"encoding/xml"
	"html"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DryWaters/bitofbytes/models"
)

var (
	titlePattern  = regexp.MustCompile(`<title>([^<]*)</title>`)
	jsonLDPattern = regexp.MustCompile(`(?s)<script type="application/ld\+json">(.*?)</script>`)
)

func get(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	rr := httptest.NewRecorder()
	newTestHandler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("%s status = %d, want %d", path, rr.Code, http.StatusOK)
	}
	return rr
}

func metaContent(t *testing.T, body, attr, name string) string {
	t.Helper()
	m := regexp.MustCompile(`<meta ` + attr + `="` + regexp.QuoteMeta(name) + `" content="([^"]*)">`).FindStringSubmatch(body)
	if m == nil {
		t.Fatalf("missing <meta %s=%q>", attr, name)
	}
	return html.UnescapeString(m[1])
}

func TestEveryPageHasUniqueTitleAndMetadata(t *testing.T) {
	t.Parallel()

	titles := map[string]string{}
	for _, page := range pagePaths() {
		body := get(t, page).Body.String()

		m := titlePattern.FindStringSubmatch(body)
		if m == nil {
			t.Fatalf("%s has no <title>", page)
		}
		title := html.UnescapeString(m[1])
		if title == "" {
			t.Errorf("%s has an empty title", page)
		}
		if page != "/projects?sort=newest" {
			if other, ok := titles[title]; ok {
				t.Errorf("%s and %s share the title %q", page, other, title)
			}
			titles[title] = page
		}

		description := metaContent(t, body, "name", "description")
		if description == "" {
			t.Errorf("%s has an empty description", page)
		}
		if got := metaContent(t, body, "property", "og:title"); got != title {
			t.Errorf("%s og:title = %q, want %q", page, got, title)
		}
		if got := metaContent(t, body, "property", "og:description"); got != description {
			t.Errorf("%s og:description = %q, want %q", page, got, description)
		}
		if got := metaContent(t, body, "name", "twitter:title"); got != title {
			t.Errorf("%s twitter:title = %q, want %q", page, got, title)
		}
		for _, image := range []string{metaContent(t, body, "property", "og:image"), metaContent(t, body, "name", "twitter:image")} {
			if !strings.HasPrefix(image, "https://bitofbytes.io/static/") {
				t.Errorf("%s image %q is not an absolute site URL", page, image)
			}
		}
		if card := metaContent(t, body, "name", "twitter:card"); card != "summary" && card != "summary_large_image" {
			t.Errorf("%s twitter:card = %q", page, card)
		}
	}
}

func TestCanonicalURLsUseApexHostWithoutQuery(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"/":                     "https://bitofbytes.io/",
		"/projects":             "https://bitofbytes.io/projects",
		"/projects?sort=newest": "https://bitofbytes.io/projects",
		"/projects/carma":       "https://bitofbytes.io/projects/carma",
	}
	for page, want := range tests {
		body := get(t, page).Body.String()
		if tag := `<link rel="canonical" href="` + want + `">`; !strings.Contains(body, tag) {
			t.Errorf("%s missing %s", page, tag)
		}
		if got := metaContent(t, body, "property", "og:url"); got != want {
			t.Errorf("%s og:url = %q, want %q", page, got, want)
		}
	}
}

func TestProjectDetailMetadataComesFromProject(t *testing.T) {
	t.Parallel()

	for _, project := range models.Projects() {
		body := get(t, "/projects/"+project.Slug).Body.String()

		if m := titlePattern.FindStringSubmatch(body); m == nil || html.UnescapeString(m[1]) != project.Name+" | BitOfBytes" {
			t.Errorf("%s title = %v, want %q", project.Slug, m, project.Name+" | BitOfBytes")
		}
		if got := metaContent(t, body, "name", "description"); got != project.Tagline {
			t.Errorf("%s description = %q, want %q", project.Slug, got, project.Tagline)
		}
		if thumb := project.Thumbnail().Path; thumb != "" {
			if got, want := metaContent(t, body, "property", "og:image"), "https://bitofbytes.io"+thumb; got != want {
				t.Errorf("%s og:image = %q, want %q", project.Slug, got, want)
			}
			if got := metaContent(t, body, "name", "twitter:card"); got != "summary_large_image" {
				t.Errorf("%s twitter:card = %q, want summary_large_image", project.Slug, got)
			}
		}
	}
}

// JSON-LD is a data block, not script, so script-src 'self' does not block it:
// the type must be a non-executable one and the CSP must not need
// 'unsafe-inline' for scripts.
func TestJSONLDIsAnInertDataBlock(t *testing.T) {
	t.Parallel()

	wantTypes := map[string]string{"/": "Person", "/projects/carma": "SoftwareApplication"}
	for _, page := range pagePaths() {
		rr := get(t, page)
		body := rr.Body.String()
		blocks := jsonLDPattern.FindAllStringSubmatch(body, -1)

		wantType, hasBlock := wantTypes[page]
		if !hasBlock && strings.HasPrefix(page, "/projects/") {
			wantType, hasBlock = "SoftwareApplication", true
		}
		if !hasBlock {
			if len(blocks) != 0 {
				t.Errorf("%s has %d JSON-LD blocks, want none", page, len(blocks))
			}
			continue
		}
		if len(blocks) != 1 {
			t.Fatalf("%s has %d JSON-LD blocks, want 1", page, len(blocks))
		}
		var data map[string]any
		if err := json.Unmarshal([]byte(blocks[0][1]), &data); err != nil {
			t.Fatalf("%s JSON-LD is not valid JSON: %v\n%s", page, err, blocks[0][1])
		}
		if data["@context"] != "https://schema.org" || data["@type"] != wantType {
			t.Errorf("%s JSON-LD @context/@type = %v/%v, want https://schema.org/%s", page, data["@context"], data["@type"], wantType)
		}

		csp := rr.Header().Get("Content-Security-Policy")
		for _, directive := range strings.Split(csp, ";") {
			directive = strings.TrimSpace(directive)
			if strings.HasPrefix(directive, "script-src") && strings.Contains(directive, "'unsafe-inline'") {
				t.Errorf("%s CSP script-src allows 'unsafe-inline': %q", page, directive)
			}
		}
	}
}

func TestRobotsTxtPointsAtSitemap(t *testing.T) {
	t.Parallel()

	rr := get(t, "/robots.txt")
	if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/plain") {
		t.Errorf("Content-Type = %q, want text/plain", got)
	}
	body := rr.Body.String()
	for _, want := range []string{"User-agent: *", "Sitemap: https://bitofbytes.io/sitemap.xml"} {
		if !strings.Contains(body, want) {
			t.Errorf("robots.txt missing %q", want)
		}
	}
	if strings.Contains(body, "Disallow: /\n") {
		t.Error("robots.txt disallows the whole site")
	}
}

func TestSitemapListsEveryPage(t *testing.T) {
	t.Parallel()

	rr := get(t, "/sitemap.xml")
	if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/xml") {
		t.Errorf("Content-Type = %q, want application/xml", got)
	}

	var set struct {
		XMLName xml.Name `xml:"urlset"`
		XMLNS   string   `xml:"xmlns,attr"`
		URLs    []struct {
			Loc     string `xml:"loc"`
			LastMod string `xml:"lastmod"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(rr.Body.Bytes(), &set); err != nil {
		t.Fatalf("sitemap is not valid XML: %v", err)
	}
	if set.XMLNS != "http://www.sitemaps.org/schemas/sitemap/0.9" {
		t.Errorf("sitemap namespace = %q", set.XMLNS)
	}

	lastMods := map[string]string{}
	for _, u := range set.URLs {
		if _, dup := lastMods[u.Loc]; dup {
			t.Errorf("sitemap lists %s twice", u.Loc)
		}
		lastMods[u.Loc] = u.LastMod
		if _, err := time.Parse("2006-01-02", u.LastMod); err != nil {
			t.Errorf("%s lastmod %q is not a date", u.Loc, u.LastMod)
		}
	}

	projects := models.Projects()
	if len(projects) == 0 {
		t.Fatal("no projects in the catalog")
	}
	if want := len(projects) + 2; len(set.URLs) != want {
		t.Errorf("sitemap has %d URLs, want %d", len(set.URLs), want)
	}
	for _, loc := range []string{"https://bitofbytes.io/", "https://bitofbytes.io/projects"} {
		if _, ok := lastMods[loc]; !ok {
			t.Errorf("sitemap missing %s", loc)
		}
	}
	for _, project := range projects {
		loc := "https://bitofbytes.io/projects/" + project.Slug
		got, ok := lastMods[loc]
		if !ok {
			t.Errorf("sitemap missing %s", loc)
			continue
		}
		if want := project.UpdatedOn().Format("2006-01-02"); got != want {
			t.Errorf("%s lastmod = %q, want %q", loc, got, want)
		}
	}
}
