package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DryWaters/bitofbytes/models"
)

// pagePaths are every page the site renders.
func pagePaths() []string {
	paths := []string{"/", "/projects", "/projects?sort=newest"}
	for _, project := range models.Projects() {
		paths = append(paths, "/projects/"+project.Slug)
	}
	return paths
}

// The CSP has no 'unsafe-inline' for styles, so pages must not use inline styles.
func TestPagesHaveNoInlineStyles(t *testing.T) {
	t.Parallel()
	handler := newTestHandler()

	for _, page := range pagePaths() {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, page, nil))
		if body := rr.Body.String(); strings.Contains(body, "style=") || strings.Contains(body, "<style") {
			t.Errorf("%s has an inline style", page)
		}
	}
}
