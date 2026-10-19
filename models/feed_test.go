package models

import (
	"slices"
	"testing"
	"time"
)

func TestUpdatesOfFlattensNewestFirst(t *testing.T) {
	t.Parallel()

	projects := []Project{
		{
			Slug:       "carma",
			Name:       "Carma",
			LastUpdate: "October 15, 2026",
			Notes:      "Most recent work added VIN decoding.",
			PreviousUpdates: []ProjectUpdate{
				{Date: "October 4, 2026", Note: "Most recent work secured cookies."},
			},
		},
		{
			Slug:       "noted",
			Name:       "Noted",
			LastUpdate: "October 15, 2026",
			Notes:      "Most recent work added search.",
			PreviousUpdates: []ProjectUpdate{
				{Date: "October 9, 2026", Note: "Fixed the reader."},
				{Date: "not a date", Note: "Dropped."},
			},
		},
	}

	got := updatesOf(projects)
	want := []FeedEntry{
		{Slug: "carma", Name: "Carma", Date: time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), Note: "Added VIN decoding."},
		{Slug: "noted", Name: "Noted", Date: time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), Note: "Added search."},
		{Slug: "noted", Name: "Noted", Date: time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC), Note: "Fixed the reader."},
		{Slug: "carma", Name: "Carma", Date: time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC), Note: "Secured cookies."},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("updatesOf() =\n%v\nwant\n%v", got, want)
	}
	if id := got[0].ID(); id != "tag:bitofbytes.io,2026-10-15:carma" {
		t.Errorf("ID() = %q", id)
	}
	if date := got[3].DateLong(); date != "Oct 4, 2026" {
		t.Errorf("DateLong() = %q", date)
	}
}

// Every update in the catalog must parse, or AllUpdates would drop it, and
// Atom IDs must be unique.
func TestAllUpdatesCoversEveryUpdateWithUniqueIDs(t *testing.T) {
	t.Parallel()

	want := 0
	for _, project := range Projects() {
		want += 1 + len(project.PreviousUpdates)
	}
	entries := AllUpdates()
	if len(entries) != want {
		t.Fatalf("AllUpdates() has %d entries, want %d (an update date does not parse)", len(entries), want)
	}

	ids := map[string]bool{}
	for i, entry := range entries {
		if ids[entry.ID()] {
			t.Errorf("duplicate ID %q", entry.ID())
		}
		ids[entry.ID()] = true
		if entry.Note == "" {
			t.Errorf("%s has an empty note", entry.ID())
		}
		if i > 0 && entry.Date.After(entries[i-1].Date) {
			t.Errorf("entry %d (%s) is newer than the one before it", i, entry.ID())
		}
	}
}
