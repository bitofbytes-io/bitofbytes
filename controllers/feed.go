package controllers

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/DryWaters/bitofbytes/models"
	"github.com/DryWaters/bitofbytes/views"
)

const atomNamespace = "http://www.w3.org/2005/Atom"

type atomFeed struct {
	XMLName  xml.Name    `xml:"feed"`
	XMLNS    string      `xml:"xmlns,attr"`
	ID       string      `xml:"id"`
	Title    string      `xml:"title"`
	Subtitle string      `xml:"subtitle"`
	Updated  string      `xml:"updated"`
	Links    []atomLink  `xml:"link"`
	Author   atomAuthor  `xml:"author"`
	Entries  []atomEntry `xml:"entry"`
}

type atomLink struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Href string `xml:"href,attr"`
}

type atomAuthor struct {
	Name string `xml:"name"`
	URI  string `xml:"uri"`
}

type atomEntry struct {
	ID      string   `xml:"id"`
	Title   string   `xml:"title"`
	Updated string   `xml:"updated"`
	Link    atomLink `xml:"link"`
	Summary string   `xml:"summary"`
}

// Feed serves feed.xml: every project update as an Atom entry, newest first.
// The updates only change with a deploy, so readers may cache it for an hour.
func (p Portfolio) Feed(w http.ResponseWriter, r *http.Request) {
	writeFeed(w, atomFeed{
		XMLNS:    atomNamespace,
		ID:       views.SiteURL + "/updates",
		Title:    "BitOfBytes project updates",
		Subtitle: "Every update to Daniel Waters's side projects, newest first.",
		Links: []atomLink{
			{Rel: "self", Type: "application/atom+xml", Href: views.SiteURL + "/feed.xml"},
			{Rel: "alternate", Type: "text/html", Href: views.SiteURL + "/updates"},
		},
	}, p.Updates)
}

// ProjectFeed serves /projects/{slug}/feed.xml: the Atom feed of one project's
// updates. An unknown slug is the branded 404 page.
func (p Portfolio) ProjectFeed(w http.ResponseWriter, r *http.Request) {
	project, ok := p.findProject(r.PathValue("slug"))
	if !ok {
		p.NotFound(w, r)
		return
	}
	writeFeed(w, atomFeed{
		XMLNS:    atomNamespace,
		ID:       views.SiteURL + "/projects/" + project.Slug,
		Title:    project.Name + " updates | " + siteName,
		Subtitle: "Every update to " + project.Name + ", newest first.",
		Links: []atomLink{
			{Rel: "self", Type: "application/atom+xml", Href: views.SiteURL + ProjectFeedPath(project.Slug)},
			{Rel: "alternate", Type: "text/html", Href: views.SiteURL + "/projects/" + project.Slug},
		},
	}, project.Updates())
}

// ProjectFeedPath is the path of a project's Atom feed.
func ProjectFeedPath(slug string) string {
	return "/projects/" + slug + "/feed.xml"
}

// writeFeed fills in the author, entries and updated time of feed and writes
// it as Atom.
func writeFeed(w http.ResponseWriter, feed atomFeed, updates []models.FeedEntry) {
	feed.Author = atomAuthor{Name: "Daniel Waters", URI: views.SiteURL + "/"}
	var newest time.Time
	for _, update := range updates {
		if update.Date.After(newest) {
			newest = update.Date
		}
		feed.Entries = append(feed.Entries, atomEntry{
			ID:      update.ID(),
			Title:   update.Name + " update",
			Updated: update.Date.Format(time.RFC3339),
			Link:    atomLink{Rel: "alternate", Type: "text/html", Href: views.SiteURL + "/projects/" + update.Slug},
			Summary: update.Note,
		})
	}
	feed.Updated = newest.Format(time.RFC3339)

	body, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		http.Error(w, "There was an error building the feed.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write([]byte(xml.Header))
	_, _ = w.Write(body)
}

// UpdatesIndex serves /updates. ?project=<slug> narrows it to one project's
// updates; an unknown slug is ignored, so a stale link still shows the list.
func (p Portfolio) UpdatesIndex(w http.ResponseWriter, r *http.Request) {
	data := UpdatesIndexData{
		Meta:     updatesMeta(),
		Updates:  p.Updates,
		Total:    len(p.Updates),
		FeedHref: "/feed.xml",
	}
	for _, project := range p.Projects {
		if updates := project.Updates(); len(updates) > 0 {
			data.Filters = append(data.Filters, UpdateFilter{Slug: project.Slug, Name: project.Name, Count: len(updates)})
		}
	}
	if project, ok := p.findProject(r.URL.Query().Get("project")); ok {
		data.Project = project.Slug
		data.ProjectName = project.Name
		data.Updates = project.Updates()
		data.FeedHref = ProjectFeedPath(project.Slug)
	}
	p.Templates.UpdatesIndex.Execute(w, r, data)
}
