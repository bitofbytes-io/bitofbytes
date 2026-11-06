package controllers

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/DryWaters/bitofbytes/models"
	"github.com/DryWaters/bitofbytes/templates"
	"github.com/DryWaters/bitofbytes/views"
)

var techProjects = []models.Project{
	{Slug: "carma", Name: "Carma", FirstCommitDate: "2026-07-31", LastUpdate: "August 3, 2026", Tech: []string{"Go", "HTMX", "Google OIDC"}},
	{Slug: "noted", Name: "Noted", FirstCommitDate: "2026-07-13", LastUpdate: "July 26, 2026", Tech: []string{"Go", "PostgreSQL"}},
	{Slug: "dined", Name: "Dined", FirstCommitDate: "2026-05-10", LastUpdate: "September 6, 2026", Tech: []string{"Go", "htmx"}},
	{Slug: "site", Name: "Site", FirstCommitDate: "2024-06-29", LastUpdate: "September 14, 2026", Tech: []string{"Go", "Templ"}},
}

func techPortfolio() Portfolio {
	page := views.Must(views.ParseFS("", templates.FS, "projects/index.gohtml", "base.gohtml"))
	return Portfolio{
		Projects:  techProjects,
		Templates: PortfolioTemplates{ProjectsIndex: page},
		Now:       fixedNow,
	}
}

func getProjects(t *testing.T, url string) string {
	t.Helper()
	rr := httptest.NewRecorder()
	techPortfolio().ProjectsIndex(rr, httptest.NewRequest(http.MethodGet, url, nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("%s status = %d, want %d", url, rr.Code, http.StatusOK)
	}
	return rr.Body.String()
}

// rowNames is the project names in the order the page lists them.
func rowNames(body string) []string {
	var names []string
	for _, line := range strings.Split(body, "\n") {
		_, rest, ok := strings.Cut(line, `<h2 class="nc-row__title"><a href="/projects/`)
		if !ok {
			continue
		}
		_, name, _ := strings.Cut(rest, `">`)
		name, _, _ = strings.Cut(name, "</a>")
		names = append(names, name)
	}
	return names
}

func TestProjectsIndexFiltersByTechInChosenSortOrder(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		url  string
		want []string
	}{
		{"/projects?tech=HTMX", []string{"Dined", "Carma"}},
		{"/projects?tech=htmx&sort=newest", []string{"Carma", "Dined"}},
		{"/projects?sort=newest&tech=%20HtMx%20", []string{"Carma", "Dined"}},
		{"/projects?tech=Go", []string{"Site", "Dined", "Carma", "Noted"}},
		{"/projects?tech=", []string{"Site", "Dined", "Carma", "Noted"}},
	} {
		if got := rowNames(getProjects(t, tt.url)); !slices.Equal(got, tt.want) {
			t.Errorf("%s lists %v, want %v", tt.url, got, tt.want)
		}
	}
	if techProjects[0].Slug != "carma" {
		t.Fatal("ProjectsIndex reordered the portfolio's projects")
	}
}

func TestProjectsIndexTechFilterKeepsCanonicalAndLinks(t *testing.T) {
	t.Parallel()
	body := getProjects(t, "/projects?tech=htmx&sort=newest")

	for _, want := range []string{
		`<link rel="canonical" href="https://bitofbytes.io/projects">`,
		`<meta property="og:url" content="https://bitofbytes.io/projects">`,
		`Projects · 2 of 4`,
		// The pill spells the technology as the first project that lists it does.
		`<p class="nc-filter__text">Technology <b>HTMX</b></p>`,
		`<a class="nc-filter__clear" href="/projects?sort=newest"`,
		// Sorting keeps the filter and filtering keeps the sort.
		`<a href="/projects?tech=HTMX">Recently updated</a>`,
		`<a href="/projects?sort=newest&amp;tech=HTMX" aria-current="true">Newest first</a>`,
		`<a class="nc-chip" href="/projects?sort=newest&amp;tech=Google&#43;OIDC">Google OIDC</a>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(body, "nc-empty") {
		t.Error("a filter with matches rendered the empty state")
	}
}

func TestProjectsIndexWithoutFilterHasNoPill(t *testing.T) {
	t.Parallel()
	body := getProjects(t, "/projects")

	for _, want := range []string{
		`<link rel="canonical" href="https://bitofbytes.io/projects">`,
		`Projects · 4<`,
		`<a href="/projects" aria-current="true">Recently updated</a>`,
		`<a href="/projects?sort=newest">Newest first</a>`,
		`<a class="nc-chip" href="/projects?tech=Go">Go</a>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(body, "nc-filter") {
		t.Error("rendered a filter pill without a filter")
	}
}

func TestProjectsIndexUnknownTechShowsEmptyState(t *testing.T) {
	t.Parallel()
	body := getProjects(t, `/projects?tech=<b>Rust</b>&sort=newest`)

	if got := rowNames(body); len(got) != 0 {
		t.Errorf("unknown technology lists %v, want none", got)
	}
	for _, want := range []string{
		`<link rel="canonical" href="https://bitofbytes.io/projects">`,
		`Projects · 0 of 4`,
		`No projects use &lt;b&gt;Rust&lt;/b&gt; yet.`,
		`<a href="/projects?sort=newest">See all projects</a>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(body, "<b>Rust</b>") {
		t.Error("technology from the query string was not escaped")
	}
}

func TestSitemapOmitsTechFilteredURLs(t *testing.T) {
	t.Parallel()
	rr := httptest.NewRecorder()
	Portfolio{Projects: techProjects}.Sitemap(rr, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))

	body := rr.Body.String()
	if !strings.Contains(body, "<loc>https://bitofbytes.io/projects</loc>") {
		t.Error("sitemap is missing /projects")
	}
	if strings.Contains(body, "tech=") {
		t.Errorf("sitemap lists a filtered URL:\n%s", body)
	}
}
