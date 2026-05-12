package models

import "testing"

func TestProjectsReturnsMostRecentProjectFirst(t *testing.T) {
	t.Parallel()

	projects := Projects()
	if len(projects) == 0 {
		t.Fatal("Projects() returned no projects")
	}
	if projects[0].Slug != "dined" {
		t.Fatalf("first project slug = %q, want %q", projects[0].Slug, "dined")
	}
}
