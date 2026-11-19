package controllers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DryWaters/bitofbytes/models"
	"github.com/DryWaters/bitofbytes/templates"
	"github.com/DryWaters/bitofbytes/views"
)

func pagerProject(slug, name string) models.Project {
	return models.Project{
		Slug:            slug,
		Name:            name,
		RepoURL:         "https://github.com/bitofbytes-io/" + slug,
		Notes:           "Most recent work shipped the first version.",
		LastUpdate:      "September 6, 2026",
		FirstCommitDate: "2026-09-01",
	}
}

func renderProjectDetail(t *testing.T, portfolio Portfolio, slug string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/projects/"+slug, nil)
	req.SetPathValue("slug", slug)
	rr := httptest.NewRecorder()
	portfolio.ProjectDetail(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /projects/%s status = %d, want %d", slug, rr.Code, http.StatusOK)
	}
	return rr.Body.String()
}

func TestProjectDetailLinksToNeighboursAndWraps(t *testing.T) {
	t.Parallel()
	page := views.Must(views.ParseFS("", templates.FS, "projects/detail.gohtml", "base.gohtml"))
	portfolio := Portfolio{
		Projects: []models.Project{
			pagerProject("alpha", "Alpha"),
			pagerProject("beta", "Beta"),
			pagerProject("gamma", "Gamma"),
		},
		Templates: PortfolioTemplates{ProjectDetail: page},
		Now:       fixedNow,
	}
	prevLink := func(slug, name string) string {
		return fmt.Sprintf(`<a class="nc-pager__link" href="/projects/%s" rel="prev"><span class="nc-pager__label">Previous project</span><span class="nc-pager__name">%s</span></a>`, slug, name)
	}
	nextLink := func(slug, name string) string {
		return fmt.Sprintf(`<a class="nc-pager__link nc-pager__link--next" href="/projects/%s" rel="next"><span class="nc-pager__label">Next project</span><span class="nc-pager__name">%s</span></a>`, slug, name)
	}

	for _, tt := range []struct {
		name string
		slug string
		prev models.Project
		next models.Project
	}{
		{"first wraps to the last", "alpha", pagerProject("gamma", "Gamma"), pagerProject("beta", "Beta")},
		{"middle", "beta", pagerProject("alpha", "Alpha"), pagerProject("gamma", "Gamma")},
		{"last wraps to the first", "gamma", pagerProject("beta", "Beta"), pagerProject("alpha", "Alpha")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			body := renderProjectDetail(t, portfolio, tt.slug)
			for _, want := range []string{
				`<nav class="nc-pager" aria-label="More projects">`,
				prevLink(tt.prev.Slug, tt.prev.Name),
				nextLink(tt.next.Slug, tt.next.Name),
			} {
				if !strings.Contains(body, want) {
					t.Errorf("missing %q", want)
				}
			}
		})
	}
}

func TestProjectDetailOmitsPagerForSingleProject(t *testing.T) {
	t.Parallel()
	page := views.Must(views.ParseFS("", templates.FS, "projects/detail.gohtml", "base.gohtml"))
	portfolio := Portfolio{
		Projects:  []models.Project{pagerProject("bare", "Bare")},
		Templates: PortfolioTemplates{ProjectDetail: page},
		Now:       fixedNow,
	}

	if body := renderProjectDetail(t, portfolio, "bare"); strings.Contains(body, "nc-pager") {
		t.Error("unexpected pager for a project with no neighbours")
	}
}

// Every catalogue page links to the projects either side of it in
// models.Projects() order.
func TestProjectDetailLinksToNeighboursForEveryProject(t *testing.T) {
	t.Parallel()
	page := views.Must(views.ParseFS("", templates.FS, "projects/detail.gohtml", "base.gohtml"))
	projects := models.Projects()
	portfolio := Portfolio{
		Projects:  projects,
		Templates: PortfolioTemplates{ProjectDetail: page},
		Now:       fixedNow,
	}

	for i, project := range projects {
		prev := projects[(i+len(projects)-1)%len(projects)]
		next := projects[(i+1)%len(projects)]
		body := renderProjectDetail(t, portfolio, project.Slug)
		for _, want := range []string{
			`href="/projects/` + prev.Slug + `" rel="prev"`,
			`href="/projects/` + next.Slug + `" rel="next"`,
		} {
			if !strings.Contains(body, want) {
				t.Errorf("/projects/%s missing %q", project.Slug, want)
			}
		}
	}
}
