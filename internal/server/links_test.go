package server

import (
	"html"
	"mime"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

var (
	internalRef = regexp.MustCompile(`\s(?:href|src)\s*=\s*(?:"(/[^"]*)"|'(/[^']*)')`)
	// An unterminated comment runs to the end of the document, as in a browser.
	htmlComment = regexp.MustCompile(`(?s)<!--.*?(?:-->|$)`)
)

// internalLinks returns the site-internal targets a page's href and src
// attributes point at. Links inside HTML comments are ignored, as are
// protocol-relative and external URLs; fragments are dropped, as a browser
// never sends them.
func internalLinks(body string) []string {
	var links []string
	body = htmlComment.ReplaceAllString(body, "")
	for _, m := range internalRef.FindAllStringSubmatch(body, -1) {
		ref := html.UnescapeString(m[1] + m[2])
		if strings.HasPrefix(ref, "//") {
			continue
		}
		ref, _, _ = strings.Cut(ref, "#")
		links = append(links, ref)
	}
	return links
}

// Every internal link and asset reachable from the site's pages resolves to a
// 200, so a renamed screenshot or a dropped route fails here rather than as a
// broken image in production. Pages it finds along the way are crawled too.
func TestEveryInternalLinkAndAssetResolves(t *testing.T) {
	t.Parallel()
	handler := newTestHandler()

	queue := pagePaths()
	seen := map[string]bool{}
	for _, page := range queue {
		seen[page] = true
	}
	from := map[string]string{}
	linked := map[string]bool{}
	checked := 0

	for len(queue) > 0 {
		target := queue[0]
		queue = queue[1:]

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, target, nil))
		checked++
		if rr.Code != http.StatusOK {
			t.Errorf("%s (linked from %q) status = %d, want %d", target, from[target], rr.Code, http.StatusOK)
			continue
		}
		if mediaType, _, _ := mime.ParseMediaType(rr.Header().Get("Content-Type")); mediaType != "text/html" {
			continue
		}
		for _, link := range internalLinks(rr.Body.String()) {
			linked[link] = true
			if !seen[link] {
				seen[link] = true
				from[link] = target
				queue = append(queue, link)
			}
		}
	}

	for _, want := range []string{"/static/daniel-waters-resume.pdf", "/feed.xml", "/projects/carma", "/projects/carma/feed.xml"} {
		if !linked[want] {
			t.Errorf("no page links to %s; the crawl is not reaching it", want)
		}
	}
	t.Logf("checked %d internal targets", checked)
}

func TestInternalLinksIgnoresHTMLComments(t *testing.T) {
	t.Parallel()

	body := `<a href="/a">a</a><!-- <a href="/gone">g</a> -->` +
		`<!--
		<img src="/also-gone.webp">
		--><a href="/b">b</a><!-- unterminated <a href="/never">n</a>`
	got := strings.Join(internalLinks(body), " ")
	if want := "/a /b"; got != want {
		t.Errorf("internalLinks = %q, want %q", got, want)
	}
}

func TestInternalLinksIgnoresExternalAndFragmentOnlyTargets(t *testing.T) {
	t.Parallel()

	body := `<a href="/projects?a=1&amp;b=2">x</a><a href="/#contact">c</a>` +
		`<a href="#top">t</a><a href="//cdn.example/x.js">p</a>` +
		`<a href="https://example.com/x">e</a><a href="mailto:a@b.c">m</a>` +
		`<img src="/static/projects/a.webp"><a href='/projects'>q</a>` +
		`<img data-src="/static/lazy.webp"><a data-href="/nope">d</a>`
	got := strings.Join(internalLinks(body), " ")
	if want := "/projects?a=1&b=2 / /static/projects/a.webp /projects"; got != want {
		t.Errorf("internalLinks = %q, want %q", got, want)
	}
}
