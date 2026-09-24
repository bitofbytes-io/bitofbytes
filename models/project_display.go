package models

import (
	"cmp"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// RecentWindow is how recently a project must have been updated to count as
// "updated this month": it lights its piano key and earns the badge.
const RecentWindow = 30 * 24 * time.Hour

const (
	lastUpdateLayout  = "January 2, 2006"
	firstCommitLayout = "2006-01-02"
)

// UpdatedOn is LastUpdate as a date, or the zero time when it cannot be parsed.
func (p Project) UpdatedOn() time.Time {
	t, _ := time.Parse(lastUpdateLayout, p.LastUpdate)
	return t
}

// StartedOn is FirstCommitDate as a date, or the zero time when it cannot be parsed.
func (p Project) StartedOn() time.Time {
	t, _ := time.Parse(firstCommitLayout, p.FirstCommitDate)
	return t
}

// UpdatedWithin reports whether the last update falls within d before now.
func (p Project) UpdatedWithin(now time.Time, d time.Duration) bool {
	updated := p.UpdatedOn()
	if updated.IsZero() {
		return false
	}
	return !updated.After(now) && now.Sub(updated) <= d
}

// UpdatedShort is the last update as "Sep 6".
func (p Project) UpdatedShort() string {
	return formatOr(p.UpdatedOn(), "Jan 2", p.LastUpdate)
}

// UpdatedLong is the last update as "Sep 6, 2026".
func (p Project) UpdatedLong() string {
	return formatOr(p.UpdatedOn(), "Jan 2, 2006", p.LastUpdate)
}

// StartedLong is the first commit as "Jul 13, 2026".
func (p Project) StartedLong() string {
	return formatOr(p.StartedOn(), "Jan 2, 2006", p.FirstCommitDate)
}

// StartedMonth is the first commit as "Jul 2026".
func (p Project) StartedMonth() string {
	return formatOr(p.StartedOn(), "Jan 2006", p.FirstCommitDate)
}

// StartedKey is the label on the project's piano key, like "since Jul '26".
func (p Project) StartedKey() string {
	started := p.StartedOn()
	if started.IsZero() {
		return ""
	}
	return "since " + started.Format("Jan '06")
}

// LatestNote is Notes without its "Most recent work" lead-in, as a sentence.
func (p Project) LatestNote() string {
	note := strings.TrimSpace(p.Notes)
	for _, prefix := range []string{"Most recent work ", "Most recent work: "} {
		if rest, ok := strings.CutPrefix(note, prefix); ok {
			note = rest
			break
		}
	}
	r, size := utf8.DecodeRuneInString(note)
	if size == 0 {
		return ""
	}
	return string(unicode.ToUpper(r)) + note[size:]
}

// Thumbnail is the first screenshot that has an image, or a zero value.
func (p Project) Thumbnail() ProjectScreenshot {
	for _, screenshot := range p.Screenshots {
		if screenshot.Path != "" {
			return screenshot
		}
	}
	return ProjectScreenshot{}
}

// TechPreview is the first few technologies, for compact listings.
func (p Project) TechPreview() []string {
	return p.Tech[:min(len(p.Tech), 4)]
}

// StackSummary is the first three technologies joined with middle dots.
func (p Project) StackSummary() string {
	return strings.Join(p.Tech[:min(len(p.Tech), 3)], " · ")
}

// NumberedHighlight is a highlight with its lower-case roman numeral.
type NumberedHighlight struct {
	Numeral string
	Text    string
}

// NumberedHighlights pairs each highlight with "i.", "ii.", "iii." and so on.
func (p Project) NumberedHighlights() []NumberedHighlight {
	numerals := []string{"i", "ii", "iii", "iv", "v", "vi", "vii", "viii", "ix", "x"}
	out := make([]NumberedHighlight, 0, len(p.Highlights))
	for i, text := range p.Highlights {
		numeral := ""
		if i < len(numerals) {
			numeral = numerals[i] + "."
		}
		out = append(out, NumberedHighlight{Numeral: numeral, Text: text})
	}
	return out
}

// SortByLastUpdate returns a copy of projects, most recently updated first.
// Projects updated on the same day keep their relative order.
func SortByLastUpdate(projects []Project) []Project {
	sorted := slices.Clone(projects)
	slices.SortStableFunc(sorted, func(a, b Project) int {
		return b.UpdatedOn().Compare(a.UpdatedOn())
	})
	return sorted
}

// SortByFirstCommit returns a copy of projects, the earliest started first.
func SortByFirstCommit(projects []Project) []Project {
	sorted := slices.Clone(projects)
	slices.SortStableFunc(sorted, func(a, b Project) int {
		return cmp.Compare(a.FirstCommitDate, b.FirstCommitDate)
	})
	return sorted
}

func formatOr(t time.Time, layout, fallback string) string {
	if t.IsZero() {
		return fallback
	}
	return t.Format(layout)
}
