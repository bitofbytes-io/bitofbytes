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
