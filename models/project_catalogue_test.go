package models

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

// allowedHobbies are the Hobby values a project may use. Empty means the
// project is not tied to a home-page hobby. The names are the ones documented
// on Project.Hobby.
var allowedHobbies = map[string]struct{}{
	"":           {},
	"piano":      {},
	"videogames": {},
}

// TestProjectCatalogue fails when hand-maintained project data would break a
// page: slugs, dates, https URLs, hobbies, or screenshot files.
func TestProjectCatalogue(t *testing.T) {
	t.Parallel()

	projects := Projects()
	if len(projects) == 0 {
		t.Fatal("Projects() returned no projects")
	}

	counts := make(map[string]int, len(projects))
	for _, project := range projects {
		counts[project.Slug]++
	}

	files := catalogueStaticFiles(t)
	for i, project := range projects {
		t.Run(catalogueCaseName(project.Slug, i), func(t *testing.T) {
			t.Parallel()
			for _, problem := range projectCatalogueProblems(project, counts[project.Slug], files) {
				t.Error(problem)
			}
		})
	}
}

func TestProjectCatalogueRejectsViolations(t *testing.T) {
	t.Parallel()

	files := fstest.MapFS{
		"projects/ok/home.webp": &fstest.MapFile{Data: []byte("webp")},
		"projects/ok/dir":       &fstest.MapFile{Mode: fs.ModeDir},
	}
	if problems := projectCatalogueProblems(validCatalogueProject(), 1, files); len(problems) != 0 {
		t.Fatalf("valid project problems = %v", problems)
	}
	unset := validCatalogueProject()
	unset.Hobby = ""
	if problems := projectCatalogueProblems(unset, 1, files); len(problems) != 0 {
		t.Fatalf("project with no hobby problems = %v", problems)
	}

	tests := []struct {
		name  string
		count int
		edit  func(*Project)
		want  string
	}{
		{name: "empty slug", count: 1, edit: func(p *Project) { p.Slug = " " }, want: "slug is empty"},
		{name: "duplicate slug", count: 2, edit: func(*Project) {}, want: `slug "ok" appears 2 times`},
		{name: "unparseable last update", count: 1, edit: func(p *Project) { p.LastUpdate = "soon" }, want: "LastUpdate"},
		{name: "impossible last update", count: 1, edit: func(p *Project) { p.LastUpdate = "February 31, 2026" }, want: "LastUpdate"},
		{name: "unparseable first commit", count: 1, edit: func(p *Project) { p.FirstCommitDate = "07-31-2026" }, want: "FirstCommitDate"},
		{name: "unparseable update date", count: 1, edit: func(p *Project) { p.PreviousUpdates[0].Date = "Oct 15, 2026" }, want: "PreviousUpdates[0].Date"},
		{name: "http repo url", count: 1, edit: func(p *Project) { p.RepoURL = "http://github.com/bitofbytes-io/ok" }, want: "RepoURL"},
		{name: "relative live url", count: 1, edit: func(p *Project) { p.LiveURL = "ok.bitofbytes.io" }, want: "LiveURL"},
		{name: "unknown hobby", count: 1, edit: func(p *Project) { p.Hobby = "chess" }, want: "Hobby"},
		{name: "missing screenshot", count: 1, edit: func(p *Project) { p.Screenshots[0].Path = "/static/projects/ok/missing.webp" }, want: "Screenshots[0].Path"},
		{name: "screenshot outside static", count: 1, edit: func(p *Project) { p.Screenshots[0].Path = "/static/../secret.webp" }, want: "not a file under /static/"},
		{name: "screenshot directory", count: 1, edit: func(p *Project) { p.Screenshots[0].Path = "/static/projects/ok/dir" }, want: "is not a file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			project := validCatalogueProject()
			tt.edit(&project)
			problems := projectCatalogueProblems(project, tt.count, files)
			if len(problems) != 1 || !strings.Contains(problems[0], tt.want) {
				t.Fatalf("problems = %v, want one problem containing %q", problems, tt.want)
			}
		})
	}
}

func validCatalogueProject() Project {
	return Project{
		Slug:            "ok",
		RepoURL:         "https://github.com/bitofbytes-io/ok",
		LiveURL:         "https://ok.bitofbytes.io",
		LastUpdate:      "October 29, 2026",
		FirstCommitDate: "2026-07-31",
		Hobby:           "piano",
		PreviousUpdates: []ProjectUpdate{{Date: "October 15, 2026", Note: "Earlier."}},
		Screenshots:     []ProjectScreenshot{{Path: "/static/projects/ok/home.webp"}},
	}
}

func catalogueCaseName(slug string, index int) string {
	if strings.TrimSpace(slug) == "" {
		return fmt.Sprintf("empty-slug-%d", index)
	}
	return slug
}

func projectCatalogueProblems(project Project, slugCount int, files fs.FS) []string {
	var problems []string
	switch {
	case strings.TrimSpace(project.Slug) == "":
		problems = append(problems, "slug is empty")
	case slugCount != 1:
		problems = append(problems, fmt.Sprintf("slug %q appears %d times", project.Slug, slugCount))
	}
	if err := catalogueDateError(lastUpdateLayout, project.LastUpdate); err != nil {
		problems = append(problems, "LastUpdate: "+err.Error())
	}
	if err := catalogueDateError(firstCommitLayout, project.FirstCommitDate); err != nil {
		problems = append(problems, "FirstCommitDate: "+err.Error())
	}
	for i, update := range project.PreviousUpdates {
		if err := catalogueDateError(lastUpdateLayout, update.Date); err != nil {
			problems = append(problems, fmt.Sprintf("PreviousUpdates[%d].Date: %v", i, err))
		}
	}
	if err := absoluteHTTPSError(project.RepoURL); err != nil {
		problems = append(problems, "RepoURL: "+err.Error())
	}
	if err := absoluteHTTPSError(project.LiveURL); err != nil {
		problems = append(problems, "LiveURL: "+err.Error())
	}
	if _, ok := allowedHobbies[project.Hobby]; !ok {
		problems = append(problems, fmt.Sprintf("Hobby %q is not allowed", project.Hobby))
	}
	for i, shot := range project.Screenshots {
		if err := screenshotFileError(files, shot.Path); err != nil {
			problems = append(problems, fmt.Sprintf("Screenshots[%d].Path: %v", i, err))
		}
	}
	return problems
}

func catalogueDateError(layout, value string) error {
	parsed, err := time.Parse(layout, value)
	if err != nil {
		return fmt.Errorf("%q is not a %q date: %w", value, layout, err)
	}
	if formatted := parsed.Format(layout); formatted != value {
		return fmt.Errorf("%q is not a calendar date (parsed as %q)", value, formatted)
	}
	return nil
}

func absoluteHTTPSError(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" || !strings.HasPrefix(raw, "https://") {
		return fmt.Errorf("%q is not an absolute https URL", raw)
	}
	return nil
}

func screenshotFileError(files fs.FS, webPath string) error {
	name, ok := screenshotRel(webPath)
	if !ok {
		return fmt.Errorf("%q is not a file under /static/", webPath)
	}
	info, err := fs.Stat(files, name)
	if err != nil {
		return fmt.Errorf("%q: %w", webPath, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%q is not a file", webPath)
	}
	return nil
}

// screenshotRel is the path inside static/ for a site path such as
// /static/projects/carma/local-garage-dashboard.webp. It rejects anything that
// is not a clean relative path under that directory.
func screenshotRel(webPath string) (string, bool) {
	name, ok := strings.CutPrefix(webPath, "/static/")
	if !ok || name == "" || name == "." {
		return "", false
	}
	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, "..") || strings.Contains(name, "\\") {
		return "", false
	}
	if path.Clean(name) != name {
		return "", false
	}
	return name, true
}

// catalogueStaticFiles is the repo static/ directory. go test runs the test
// binary from this package's directory, so ../static is stable under -trimpath,
// which strips the source path runtime.Caller would return.
func catalogueStaticFiles(t *testing.T) fs.FS {
	t.Helper()
	const dir = "../static"
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("static directory: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("static path %s is not a directory", dir)
	}
	return os.DirFS(dir)
}
