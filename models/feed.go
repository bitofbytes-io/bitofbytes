package models

import (
	"slices"
	"time"
)

// FeedEntry is one dated project update in the site-wide changelog.
type FeedEntry struct {
	Slug string
	Name string
	Date time.Time
	// Note is the update with its "Most recent work" lead-in removed.
	Note string
}

// ID is the entry's stable Atom ID, "tag:bitofbytes.io,2026-10-15:carma".
func (e FeedEntry) ID() string {
	return "tag:bitofbytes.io," + e.Date.Format(firstCommitLayout) + ":" + e.Slug
}

// DateLong is the date as "Sep 6, 2026".
func (e FeedEntry) DateLong() string {
	return e.Date.Format("Jan 2, 2006")
}

// AllUpdates is every project's latest and previous updates, newest first.
func AllUpdates() []FeedEntry {
	return updatesOf(Projects())
}

// updatesOf flattens the updates of projects, newest first. Updates on the
// same day keep the order of projects. An update whose date cannot be parsed
// is left out, since it has no place on a timeline.
func updatesOf(projects []Project) []FeedEntry {
	var entries []FeedEntry
	add := func(project Project, date, note string) {
		t, err := time.Parse(lastUpdateLayout, date)
		if err != nil {
			return
		}
		entries = append(entries, FeedEntry{Slug: project.Slug, Name: project.Name, Date: t, Note: displayUpdateNote(note)})
	}
	for _, project := range projects {
		add(project, project.LastUpdate, project.Notes)
		for _, update := range project.PreviousUpdates {
			add(project, update.Date, update.Note)
		}
	}
	slices.SortStableFunc(entries, func(a, b FeedEntry) int {
		return b.Date.Compare(a.Date)
	})
	return entries
}
