package models

import "testing"

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
