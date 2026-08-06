package models

import (
	"slices"
	"testing"
)

func TestProjectsReturnsMostRecentProjectsFirst(t *testing.T) {
	t.Parallel()

	projects := Projects()
	if len(projects) == 0 {
		t.Fatal("Projects() returned no projects")
	}

	for i := 1; i < len(projects); i++ {
		if projects[i-1].FirstCommitDate < projects[i].FirstCommitDate {
			t.Fatalf(
				"Projects() not sorted by first commit date descending: project %q date %q before project %q date %q",
				projects[i-1].Slug,
				projects[i-1].FirstCommitDate,
				projects[i].Slug,
				projects[i].FirstCommitDate,
			)
		}
	}
}

func TestProjectsIncludesCarmaAndNotedAheadOfExistingProjects(t *testing.T) {
	t.Parallel()

	projects := Projects()
	if got, want := len(projects), 8; got != want {
		t.Fatalf("Projects() length = %d, want %d", got, want)
	}

	gotFirstTwo := []string{projects[0].Slug, projects[1].Slug}
	wantFirstTwo := []string{"carma", "noted"}
	if !slices.Equal(gotFirstTwo, wantFirstTwo) {
		t.Fatalf("first two project slugs = %v, want %v", gotFirstTwo, wantFirstTwo)
	}

	for _, slug := range []string{"dined", "permitpal", "learnd", "dejaview", "bitofbytes", "anthology"} {
		if !slices.ContainsFunc(projects, func(project Project) bool { return project.Slug == slug }) {
			t.Errorf("Projects() missing existing project %q", slug)
		}
	}
}
