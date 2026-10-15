package controllers

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/DryWaters/bitofbytes/models"
	"github.com/DryWaters/bitofbytes/views"
)

const (
	sitemapNamespace  = "http://www.sitemaps.org/schemas/sitemap/0.9"
	sitemapDateLayout = "2006-01-02"
)

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

// Sitemap serves sitemap.xml: the home page, the project index and every
// project page. A project's lastmod is its last update and the project index
// takes the newest one. The home page has no lastmod: it shows content that
// carries no date, so any date would be a guess.
func (p Portfolio) Sitemap(w http.ResponseWriter, r *http.Request) {
	var newest time.Time
	for _, project := range p.Projects {
		if updated := project.UpdatedOn(); updated.After(newest) {
			newest = updated
		}
	}

	set := sitemapURLSet{XMLNS: sitemapNamespace}
	set.URLs = append(set.URLs,
		sitemapEntry("/", time.Time{}),
		sitemapEntry("/projects", newest),
	)
	for _, project := range models.SortByLastUpdate(p.Projects) {
		set.URLs = append(set.URLs, sitemapEntry("/projects/"+project.Slug, project.UpdatedOn()))
	}

	body, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		http.Error(w, "There was an error building the sitemap.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, _ = w.Write([]byte(xml.Header))
	_, _ = w.Write(body)
}

func sitemapEntry(path string, lastMod time.Time) sitemapURL {
	entry := sitemapURL{Loc: views.SiteURL + path}
	if !lastMod.IsZero() {
		entry.LastMod = lastMod.Format(sitemapDateLayout)
	}
	return entry
}
