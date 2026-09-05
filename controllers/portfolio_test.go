package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/DryWaters/bitofbytes/models"
	"github.com/DryWaters/bitofbytes/templates"
	"github.com/DryWaters/bitofbytes/views"
)

func newTestPortfolio(t *testing.T) Portfolio {
	t.Helper()

	fsys := fstest.MapFS{
		"home/index.tmpl": {
			Data: []byte(`Home Daniel Waters{{ range .Projects }} {{ .Name }}{{ end }}`),
		},
		"projects/index.tmpl": {
			Data: []byte(`Projects{{ range .Projects }} {{ .Name }}{{ end }}`),
		},
		"projects/detail.tmpl": {
			Data: []byte(`Detail: {{ .Project.Name }} {{ .Project.Tagline }}`),
		},
	}

	home, err := views.ParseFS(fsys, "home/index.tmpl")
	if err != nil {
		t.Fatalf("parse home template: %v", err)
	}
	index, err := views.ParseFS(fsys, "projects/index.tmpl")
	if err != nil {
		t.Fatalf("parse index template: %v", err)
	}
	detail, err := views.ParseFS(fsys, "projects/detail.tmpl")
	if err != nil {
		t.Fatalf("parse detail template: %v", err)
	}

	return Portfolio{
		Projects: []models.Project{
			{
				Slug:    "permitpal",
				Name:    "PermitPal",
				Tagline: "Your permit pal-less yelling, more tracking.",
			},
			{
				Slug:    "dejaview",
				Name:    "DejaView",
				Tagline: "A movie tracking application built with Go and server-rendered UI.",
			},
		},
		Templates: PortfolioTemplates{
			Home:          home,
			ProjectsIndex: index,
			ProjectDetail: detail,
		},
	}
}

func TestPortfolioHomeSuppliesProfileAndHandheldProjects(t *testing.T) {
	t.Parallel()

	portfolio := newTestPortfolio(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	portfolio.Home(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Home status code = %d, want %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Daniel Waters") {
		t.Fatalf("Home body = %q, want profile content", body)
	}
	if !strings.Contains(body, "PermitPal") || !strings.Contains(body, "DejaView") {
		t.Fatalf("Home body = %q, want supplied handheld project names", body)
	}
}

func TestPortfolioProjectsIndexRendersProjects(t *testing.T) {
	t.Parallel()

	portfolio := newTestPortfolio(t)
	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	rr := httptest.NewRecorder()

	portfolio.ProjectsIndex(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ProjectsIndex status code = %d, want %d", rr.Code, http.StatusOK)
	}
	if body := rr.Body.String(); !strings.Contains(body, "PermitPal") || !strings.Contains(body, "DejaView") {
		t.Fatalf("ProjectsIndex body = %q, want project names", body)
	}
}

func TestPortfolioProjectDetailRendersSelectedProject(t *testing.T) {
	t.Parallel()

	portfolio := newTestPortfolio(t)
	req := httptest.NewRequest(http.MethodGet, "/projects/permitpal", nil)
	req.SetPathValue("slug", "permitpal")
	rr := httptest.NewRecorder()

	portfolio.ProjectDetail(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ProjectDetail status code = %d, want %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Detail: PermitPal") {
		t.Fatalf("ProjectDetail body = %q, want selected project", body)
	}
	if strings.Contains(body, "DejaView") {
		t.Fatalf("ProjectDetail body = %q, should not render another project", body)
	}
}

func TestPortfolioProjectDetailReturnsNotFoundForUnknownSlug(t *testing.T) {
	t.Parallel()

	portfolio := newTestPortfolio(t)
	req := httptest.NewRequest(http.MethodGet, "/projects/missing", nil)
	req.SetPathValue("slug", "missing")
	rr := httptest.NewRecorder()

	portfolio.ProjectDetail(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("ProjectDetail status code = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestHomepageEscapesHandheldProjectAttributes(t *testing.T) {
	t.Parallel()
	page := views.Must(views.ParseFS(templates.FS, "home/index.gohtml", "base.gohtml"))
	portfolio := Portfolio{
		Projects: []models.Project{{
			Slug:    "example",
			Name:    `A "quoted" project`,
			Summary: `A description with <script>alert("unsafe")</script> and & characters.`,
		}},
		Templates: PortfolioTemplates{Home: page},
	}
	rr := httptest.NewRecorder()
	portfolio.Home(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("Home status = %d", rr.Code)
	}
	body := rr.Body.String()
	for _, want := range []string{
		`data-name="A &#34;quoted&#34; project"`,
		`data-summary="A description with &lt;script&gt;alert(&#34;unsafe&#34;)&lt;/script&gt; and &amp; characters."`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing safely escaped project data %q", want)
		}
	}
	if strings.Contains(body, `<script>alert(`) {
		t.Error("project content escaped its data attribute")
	}
}
