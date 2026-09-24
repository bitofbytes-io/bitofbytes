package models

import (
	"slices"
	"testing"
	"time"
)

func TestProjectDisplayDates(t *testing.T) {
	t.Parallel()

	p := Project{LastUpdate: "September 6, 2026", FirstCommitDate: "2025-10-30"}
	for _, tt := range []struct{ name, got, want string }{
		{"UpdatedShort", p.UpdatedShort(), "Sep 6"},
		{"UpdatedLong", p.UpdatedLong(), "Sep 6, 2026"},
		{"StartedLong", p.StartedLong(), "Oct 30, 2025"},
		{"StartedMonth", p.StartedMonth(), "Oct 2025"},
		{"StartedKey", p.StartedKey(), "since Oct '25"},
	} {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}

	bad := Project{LastUpdate: "soon", FirstCommitDate: "someday"}
	if got := bad.UpdatedLong(); got != "soon" {
		t.Errorf("unparseable UpdatedLong = %q, want the raw value", got)
	}
	if got := bad.StartedKey(); got != "" {
		t.Errorf("unparseable StartedKey = %q, want empty", got)
	}
}

func TestProjectUpdatedWithin(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.September, 24, 12, 0, 0, 0, time.UTC)
	for _, tt := range []struct {
		lastUpdate string
		want       bool
	}{
		{"September 6, 2026", true},
		{"August 26, 2026", true},
		{"August 25, 2026", false},
		{"September 30, 2026", false},
		{"not a date", false},
	} {
		if got := (Project{LastUpdate: tt.lastUpdate}).UpdatedWithin(now, RecentWindow); got != tt.want {
			t.Errorf("UpdatedWithin(%q) = %v, want %v", tt.lastUpdate, got, tt.want)
		}
	}
}

func TestProjectLatestNoteDropsLeadIn(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct{ notes, want string }{
		{"Most recent work improved search.", "Improved search."},
		{"Most recent work: tidied the reader.", "Tidied the reader."},
		{"Shipped the first version.", "Shipped the first version."},
		{"", ""},
	} {
		if got := (Project{Notes: tt.notes}).LatestNote(); got != tt.want {
			t.Errorf("LatestNote(%q) = %q, want %q", tt.notes, got, tt.want)
		}
	}
}

func TestProjectListingHelpers(t *testing.T) {
	t.Parallel()

	p := Project{
		Tech:       []string{"Go", "HTMX", "PostgreSQL", "Docker", "Goose"},
		Highlights: []string{"One.", "Two."},
		Screenshots: []ProjectScreenshot{
			{Title: "Pending"},
			{Title: "Home", Path: "/static/projects/x/home.png"},
		},
	}
	if got := p.TechPreview(); !slices.Equal(got, []string{"Go", "HTMX", "PostgreSQL", "Docker"}) {
		t.Errorf("TechPreview = %v", got)
	}
	if got, want := p.StackSummary(), "Go · HTMX · PostgreSQL"; got != want {
		t.Errorf("StackSummary = %q, want %q", got, want)
	}
	if got := p.Thumbnail().Title; got != "Home" {
		t.Errorf("Thumbnail = %q, want the first screenshot with an image", got)
	}
	if got := p.NumberedHighlights(); got[0].Numeral != "i." || got[1].Numeral != "ii." || got[1].Text != "Two." {
		t.Errorf("NumberedHighlights = %+v", got)
	}
	if got := (Project{}).Thumbnail().Path; got != "" {
		t.Errorf("Thumbnail without screenshots = %q, want empty", got)
	}
}

func TestSortByLastUpdateKeepsSameDayOrder(t *testing.T) {
	t.Parallel()

	projects := []Project{
		{Slug: "older", LastUpdate: "August 3, 2026"},
		{Slug: "first", LastUpdate: "September 6, 2026"},
		{Slug: "second", LastUpdate: "September 6, 2026"},
		{Slug: "newest", LastUpdate: "September 14, 2026"},
	}
	var got []string
	for _, p := range SortByLastUpdate(projects) {
		got = append(got, p.Slug)
	}
	if want := []string{"newest", "first", "second", "older"}; !slices.Equal(got, want) {
		t.Fatalf("SortByLastUpdate = %v, want %v", got, want)
	}
	if projects[0].Slug != "older" {
		t.Fatal("SortByLastUpdate modified its input")
	}
}

func TestProjectsHaveDisplayableDatesAndNoEmDashes(t *testing.T) {
	t.Parallel()

	for _, p := range Projects() {
		if p.UpdatedOn().IsZero() {
			t.Errorf("%s: LastUpdate %q is not in the \"January 2, 2006\" format", p.Slug, p.LastUpdate)
		}
		if p.StartedOn().IsZero() {
			t.Errorf("%s: FirstCommitDate %q is not in the 2006-01-02 format", p.Slug, p.FirstCommitDate)
		}
		for _, text := range append([]string{p.Tagline, p.Summary, p.Notes}, append(p.Highlights, p.Paragraphs...)...) {
			if slices.Contains([]rune(text), '—') {
				t.Errorf("%s: copy contains an em dash: %q", p.Slug, text)
			}
		}
	}
}
