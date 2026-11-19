package main

import (
	"encoding/xml"
	"html"
	"strings"
	"testing"
	"time"

	"github.com/DryWaters/bitofbytes/models"
)

func TestFeedIsAtomOfEveryUpdate(t *testing.T) {
	t.Parallel()

	rr := get(t, "/feed.xml")
	if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/atom+xml") {
		t.Errorf("Content-Type = %q, want application/atom+xml", got)
	}
	if got := rr.Header().Get("Cache-Control"); got != "public, max-age=3600" {
		t.Errorf("Cache-Control = %q, want public, max-age=3600", got)
	}

	var feed struct {
		XMLName xml.Name `xml:"http://www.w3.org/2005/Atom feed"`
		ID      string   `xml:"id"`
		Title   string   `xml:"title"`
		Updated string   `xml:"updated"`
		Author  struct {
			Name string `xml:"name"`
		} `xml:"author"`
		Links []struct {
			Rel  string `xml:"rel,attr"`
			Href string `xml:"href,attr"`
		} `xml:"link"`
		Entries []struct {
			ID      string `xml:"id"`
			Title   string `xml:"title"`
			Updated string `xml:"updated"`
			Link    struct {
				Href string `xml:"href,attr"`
			} `xml:"link"`
			Summary string `xml:"summary"`
		} `xml:"entry"`
	}
	if err := xml.Unmarshal(rr.Body.Bytes(), &feed); err != nil {
		t.Fatalf("feed is not valid Atom XML: %v", err)
	}
	if feed.ID == "" || feed.Title == "" || feed.Author.Name != "Daniel Waters" {
		t.Errorf("feed id/title/author = %q/%q/%q", feed.ID, feed.Title, feed.Author.Name)
	}
	links := map[string]string{}
	for _, link := range feed.Links {
		links[link.Rel] = link.Href
	}
	if links["self"] != "https://bitofbytes.io/feed.xml" || links["alternate"] != "https://bitofbytes.io/updates" {
		t.Errorf("feed links = %v", links)
	}

	updates := models.AllUpdates()
	if len(feed.Entries) != len(updates) {
		t.Fatalf("feed has %d entries, want %d", len(feed.Entries), len(updates))
	}
	if want := updates[0].Date.Format(time.RFC3339); feed.Updated != want {
		t.Errorf("feed updated = %q, want %q", feed.Updated, want)
	}
	for i, entry := range feed.Entries {
		update := updates[i]
		if entry.ID != update.ID() || !strings.HasPrefix(entry.ID, "tag:bitofbytes.io,") {
			t.Errorf("entry %d id = %q, want %q", i, entry.ID, update.ID())
		}
		if entry.Title != update.Name+" update" {
			t.Errorf("entry %d title = %q", i, entry.Title)
		}
		if _, err := time.Parse(time.RFC3339, entry.Updated); err != nil {
			t.Errorf("entry %d updated %q is not RFC 3339", i, entry.Updated)
		}
		if want := "https://bitofbytes.io/projects/" + update.Slug; entry.Link.Href != want {
			t.Errorf("entry %d link = %q, want %q", i, entry.Link.Href, want)
		}
		if entry.Summary != update.Note {
			t.Errorf("entry %d summary = %q, want %q", i, entry.Summary, update.Note)
		}
	}
}

func TestUpdatesPageListsEveryUpdateNewestFirst(t *testing.T) {
	t.Parallel()

	body := html.UnescapeString(get(t, "/updates").Body.String())
	updates := models.AllUpdates()
	if got := strings.Count(body, `class="nc-timeline__item`); got != len(updates) {
		t.Errorf("timeline items = %d, want %d", got, len(updates))
	}
	if got := strings.Count(body, "nc-timeline__item--latest"); got != 1 {
		t.Errorf("latest items = %d, want 1", got)
	}
	last := 0
	for _, update := range updates {
		at := strings.Index(body[last:], update.Note)
		if at < 0 {
			t.Fatalf("missing or out of order: %s %q", update.ID(), update.Note)
		}
		last += at
	}
	if !strings.Contains(body, `href="/feed.xml"`) {
		t.Error("updates page does not link the feed")
	}
}

func TestPagesAdvertiseTheFeed(t *testing.T) {
	t.Parallel()

	for _, page := range pagePaths() {
		body := get(t, page).Body.String()
		if want := `<link rel="alternate" type="application/atom+xml" title="BitOfBytes project updates" href="/feed.xml">`; !strings.Contains(body, want) {
			t.Errorf("%s missing feed autodiscovery link", page)
		}
		if want := `<a href="/feed.xml" type="application/atom+xml">Subscribe</a>`; !strings.Contains(body, want) {
			t.Errorf("%s footer missing Subscribe link", page)
		}
	}

	if home := get(t, "/").Body.String(); !strings.Contains(home, `href="/updates">See all updates`) {
		t.Error("home page missing See all updates link")
	}
}

type atomDoc struct {
	XMLName xml.Name `xml:"http://www.w3.org/2005/Atom feed"`
	ID      string   `xml:"id"`
	Title   string   `xml:"title"`
	Updated string   `xml:"updated"`
	Author  struct {
		Name string `xml:"name"`
	} `xml:"author"`
	Links []struct {
		Rel  string `xml:"rel,attr"`
		Href string `xml:"href,attr"`
	} `xml:"link"`
	Entries []struct {
		ID      string `xml:"id"`
		Title   string `xml:"title"`
		Updated string `xml:"updated"`
		Link    struct {
			Href string `xml:"href,attr"`
		} `xml:"link"`
		Summary string `xml:"summary"`
	} `xml:"entry"`
}

func TestProjectFeedHasOnlyThatProjectsUpdates(t *testing.T) {
	t.Parallel()

	for _, project := range models.Projects() {
		rr := get(t, "/projects/"+project.Slug+"/feed.xml")
		if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/atom+xml") {
			t.Errorf("%s Content-Type = %q, want application/atom+xml", project.Slug, got)
		}
		if got := rr.Header().Get("Cache-Control"); got != "public, max-age=3600" {
			t.Errorf("%s Cache-Control = %q", project.Slug, got)
		}

		var feed atomDoc
		if err := xml.Unmarshal(rr.Body.Bytes(), &feed); err != nil {
			t.Fatalf("%s feed is not valid Atom XML: %v", project.Slug, err)
		}
		if feed.ID != "https://bitofbytes.io/projects/"+project.Slug || feed.Title != project.Name+" updates | BitOfBytes" || feed.Author.Name != "Daniel Waters" {
			t.Errorf("%s feed id/title/author = %q/%q/%q", project.Slug, feed.ID, feed.Title, feed.Author.Name)
		}
		links := map[string]string{}
		for _, link := range feed.Links {
			links[link.Rel] = link.Href
		}
		if want := "https://bitofbytes.io/projects/" + project.Slug + "/feed.xml"; links["self"] != want {
			t.Errorf("%s self link = %q, want %q", project.Slug, links["self"], want)
		}
		if want := "https://bitofbytes.io/projects/" + project.Slug; links["alternate"] != want {
			t.Errorf("%s alternate link = %q, want %q", project.Slug, links["alternate"], want)
		}

		updates := project.Updates()
		if len(feed.Entries) != len(updates) || len(updates) == 0 {
			t.Fatalf("%s feed has %d entries, want %d", project.Slug, len(feed.Entries), len(updates))
		}
		if want := updates[0].Date.Format(time.RFC3339); feed.Updated != want {
			t.Errorf("%s feed updated = %q, want %q", project.Slug, feed.Updated, want)
		}
		for i, entry := range feed.Entries {
			if entry.ID != updates[i].ID() || !strings.HasSuffix(entry.ID, ":"+project.Slug) {
				t.Errorf("%s entry %d id = %q, want %q", project.Slug, i, entry.ID, updates[i].ID())
			}
			if entry.Title != project.Name+" update" || entry.Summary != updates[i].Note {
				t.Errorf("%s entry %d title/summary = %q/%q", project.Slug, i, entry.Title, entry.Summary)
			}
			if want := "https://bitofbytes.io/projects/" + project.Slug; entry.Link.Href != want {
				t.Errorf("%s entry %d link = %q, want %q", project.Slug, i, entry.Link.Href, want)
			}
		}
	}
}

func TestUpdatesPageFiltersByProject(t *testing.T) {
	t.Parallel()

	project := models.Projects()[0]
	updates := project.Updates()
	body := html.UnescapeString(get(t, "/updates?project="+project.Slug).Body.String())

	if got := strings.Count(body, `class="nc-timeline__item`); got != len(updates) {
		t.Errorf("timeline items = %d, want %d", got, len(updates))
	}
	for _, other := range models.Projects()[1:] {
		for _, update := range other.Updates() {
			if strings.Contains(body, `<span class="nc-timeline__text">`+update.Note+`</span>`) {
				t.Errorf("%s update shown on the %s filter: %q", other.Slug, project.Slug, update.Note)
			}
		}
	}
	for _, want := range []string{
		`href="/projects/` + project.Slug + `/feed.xml" type="application/atom+xml">Atom feed`,
		`<a class="nc-chip" href="/updates?project=` + project.Slug + `" aria-current="true">` + project.Name,
		`<a class="nc-chip" href="/updates">All projects`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("filtered page missing %q", want)
		}
	}
	if got := strings.Count(body, `aria-current="true"`); got != 1 {
		t.Errorf("active chips = %d, want 1", got)
	}

	all := html.UnescapeString(get(t, "/updates").Body.String())
	if !strings.Contains(all, `<a class="nc-chip" href="/updates" aria-current="true">All projects`) {
		t.Error("unfiltered page does not mark All projects current")
	}
	for _, other := range models.Projects() {
		if !strings.Contains(all, `href="/updates?project=`+other.Slug+`"`) {
			t.Errorf("updates page has no chip for %s", other.Slug)
		}
	}
}

func TestUpdatesPageIgnoresUnknownProject(t *testing.T) {
	t.Parallel()

	got := get(t, "/updates?project=nope").Body.String()
	if want := get(t, "/updates").Body.String(); got != want {
		t.Error("unknown ?project= changed the updates page")
	}
}

func TestProjectPageLinksItsFeed(t *testing.T) {
	t.Parallel()

	for _, project := range models.Projects() {
		body := get(t, "/projects/"+project.Slug).Body.String()
		for _, want := range []string{
			`href="/projects/` + project.Slug + `/feed.xml" type="application/atom+xml">Atom feed`,
			`href="/updates?project=` + project.Slug + `">All `,
		} {
			if !strings.Contains(body, want) {
				t.Errorf("%s page missing %q", project.Slug, want)
			}
		}
	}
}
