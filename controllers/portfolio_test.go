package controllers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/DryWaters/bitofbytes/models"
	"github.com/DryWaters/bitofbytes/templates"
	"github.com/DryWaters/bitofbytes/views"
)

func newTestPortfolio(t *testing.T) Portfolio {
	t.Helper()

	fsys := fstest.MapFS{
		"home/index.tmpl": {
			Data: []byte(`Home Daniel Waters{{ range .Keys }} {{ .Name }}{{ end }}`),
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
				Tagline: "Your permit pal: less yelling, more tracking.",
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

func TestPortfolioHomeSuppliesProfileAndKeyboardProjects(t *testing.T) {
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
		t.Fatalf("Home body = %q, want supplied keyboard project names", body)
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

func TestHomepageEscapesProjectContent(t *testing.T) {
	t.Parallel()
	page := views.Must(views.ParseFS(templates.FS, "home/index.gohtml", "base.gohtml"))
	portfolio := Portfolio{
		Projects: []models.Project{{
			Slug:            "example",
			Name:            `A "quoted" <b>project</b>`,
			Notes:           `Most recent work added <script>alert("unsafe")</script> & more.`,
			LastUpdate:      "September 6, 2026",
			FirstCommitDate: "2026-01-02",
		}},
		Templates: PortfolioTemplates{Home: page},
		Now:       fixedNow,
	}
	rr := httptest.NewRecorder()
	portfolio.Home(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("Home status = %d", rr.Code)
	}
	body := rr.Body.String()
	for _, want := range []string{
		`A &#34;quoted&#34; &lt;b&gt;project&lt;/b&gt;`,
		`Added &lt;script&gt;alert(&#34;unsafe&#34;)&lt;/script&gt; &amp; more.`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing safely escaped project content %q", want)
		}
	}
	if strings.Contains(body, `<script>alert(`) || strings.Contains(body, `<b>project`) {
		t.Error("project content was not escaped")
	}
}

func fixedNow() time.Time { return time.Date(2026, time.September, 24, 12, 0, 0, 0, time.UTC) }

func keyboardPortfolio(t *testing.T, projects []models.Project) Portfolio {
	t.Helper()
	fsys := fstest.MapFS{
		"home/index.tmpl": {Data: []byte(
			`{{ range .Keys }}[{{ .Slug }} {{ .Note }}{{ if .Recent }} lit{{ end }}]{{ end }}` +
				`|{{ range .BlackKeys }}{{ .Pos }}{{ end }}` +
				`|{{ range .Updates }}{{ .Slug }} {{ end }}` +
				`|{{ .KeyCount }} {{ .RecentCount }}`)},
		"projects/index.tmpl": {Data: []byte(`{{ .Sort }}:{{ range .Projects }} {{ .Slug }}{{ if .Recent }}*{{ end }}{{ end }}`)},
	}
	return Portfolio{
		Projects: projects,
		Templates: PortfolioTemplates{
			Home:          views.Must(views.ParseFS(fsys, "home/index.tmpl")),
			ProjectsIndex: views.Must(views.ParseFS(fsys, "projects/index.tmpl")),
		},
		Now: fixedNow,
	}
}

// Projects arrive newest-started first, as models.Projects returns them.
var sampleProjects = []models.Project{
	{Slug: "carma", FirstCommitDate: "2026-07-31", LastUpdate: "August 3, 2026"},
	{Slug: "noted", FirstCommitDate: "2026-07-13", LastUpdate: "July 26, 2026"},
	{Slug: "dined", FirstCommitDate: "2026-05-10", LastUpdate: "September 6, 2026"},
	{Slug: "site", FirstCommitDate: "2024-06-29", LastUpdate: "September 14, 2026"},
}

func TestHomeKeyboardOrdersByStartAndLightsRecentWork(t *testing.T) {
	t.Parallel()
	rr := httptest.NewRecorder()
	keyboardPortfolio(t, sampleProjects).Home(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	want := "[site C4 lit][dined D4 lit][noted E4][carma F4]" +
		"|12" +
		"|site dined carma noted " +
		"|Four 2"
	if got := rr.Body.String(); got != want {
		t.Fatalf("home data =\n%q\nwant\n%q", got, want)
	}
}

func TestHomeKeyboardHoldsOneOctaveOfTheMostRecentlyUpdated(t *testing.T) {
	t.Parallel()
	var projects []models.Project
	for i := range 10 {
		projects = append(projects, models.Project{
			Slug:            string(rune('a' + i)),
			FirstCommitDate: fmt.Sprintf("2025-01-%02d", 10-i),
			LastUpdate:      time.Date(2026, time.January, 1+i, 0, 0, 0, 0, time.UTC).Format("January 2, 2006"),
		})
	}
	rr := httptest.NewRecorder()
	keyboardPortfolio(t, projects).Home(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	keys, _, _ := strings.Cut(rr.Body.String(), "|")
	// a and b were updated least recently, so they drop off; the rest play in start order.
	if want := "[j C4][i D4][h E4][g F4][f G4][e A4][d B4][c C5]"; keys != want {
		t.Fatalf("keys = %q, want %q", keys, want)
	}
}

func TestProjectsIndexSortsByUpdateUnlessNewestRequested(t *testing.T) {
	t.Parallel()
	portfolio := keyboardPortfolio(t, sampleProjects)

	for _, tt := range []struct{ url, want string }{
		{"/projects", "updated: site* dined* carma noted"},
		{"/projects?sort=unknown", "updated: site* dined* carma noted"},
		{"/projects?sort=newest", "newest: carma noted dined* site*"},
	} {
		rr := httptest.NewRecorder()
		portfolio.ProjectsIndex(rr, httptest.NewRequest(http.MethodGet, tt.url, nil))
		if got := rr.Body.String(); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestProjectsIndexNewestFirstIgnoresDeclarationOrder(t *testing.T) {
	t.Parallel()
	shuffled := []models.Project{sampleProjects[2], sampleProjects[3], sampleProjects[0], sampleProjects[1]}
	portfolio := keyboardPortfolio(t, shuffled)

	rr := httptest.NewRecorder()
	portfolio.ProjectsIndex(rr, httptest.NewRequest(http.MethodGet, "/projects?sort=newest", nil))
	if got, want := rr.Body.String(), "newest: carma noted dined* site*"; got != want {
		t.Fatalf("newest = %q, want %q", got, want)
	}
	if shuffled[0].Slug != "dined" {
		t.Fatal("ProjectsIndex reordered the portfolio's projects")
	}
}

func TestProjectDetailWithoutScreenshotsUsesSoloHero(t *testing.T) {
	t.Parallel()
	page := views.Must(views.ParseFS(templates.FS, "projects/detail.gohtml", "base.gohtml"))
	portfolio := Portfolio{
		Projects: []models.Project{{
			Slug:            "bare",
			Name:            "Bare",
			RepoURL:         "https://github.com/bitofbytes-io/bare",
			Notes:           "Most recent work shipped the first version.",
			LastUpdate:      "September 6, 2026",
			FirstCommitDate: "2026-09-01",
		}},
		Templates: PortfolioTemplates{ProjectDetail: page},
		Now:       fixedNow,
	}
	req := httptest.NewRequest(http.MethodGet, "/projects/bare", nil)
	req.SetPathValue("slug", "bare")
	rr := httptest.NewRecorder()
	portfolio.ProjectDetail(rr, req)

	body := rr.Body.String()
	for _, want := range []string{
		`nc-detail-hero nc-detail-hero--solo`,
		`<a class="nc-btn nc-btn--primary" href="https://github.com/bitofbytes-io/bare">View on GitHub</a>`,
		`Shipped the first version.`,
		`Sep 1, 2026`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
	for _, unwanted := range []string{`nc-hero-shot`, `Screenshots`, `Open live site`} {
		if strings.Contains(body, unwanted) {
			t.Errorf("unexpected %q for a project without screenshots or live site", unwanted)
		}
	}
}
