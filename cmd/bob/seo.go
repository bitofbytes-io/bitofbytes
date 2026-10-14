package main

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/DryWaters/bitofbytes/models"
)

// canonicalHost is the apex origin (Traefik redirects www to the apex).
const canonicalHost = "https://bitofbytes.io"

const robotsTXT = "User-agent: *\nAllow: /\n\nSitemap: https://bitofbytes.io/sitemap.xml\n"

// serveRobotsTXT serves the site's robots.txt.
func serveRobotsTXT(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(robotsTXT))
}

// sitemapURL is one <url> entry in the sitemap.
type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

// urlset is the sitemap document root.
type urlset struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

// serveSitemapXML renders the sitemap from the project catalog. Every page
// gets an entry; project entries carry <lastmod> from the project's parsed
// last-update date so crawlers see fresh pages.
func serveSitemapXML(w http.ResponseWriter, r *http.Request) {
	set := urlset{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	set.URLs = append(set.URLs,
		sitemapURL{Loc: canonicalHost + "/"},
		sitemapURL{Loc: canonicalHost + "/projects"},
	)
	for _, project := range models.Projects() {
		entry := sitemapURL{Loc: canonicalHost + "/projects/" + project.Slug}
		if updated := project.UpdatedOn(); !updated.IsZero() {
			entry.LastMod = updated.Format(time.DateOnly)
		}
		set.URLs = append(set.URLs, entry)
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write([]byte(xml.Header))
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	_ = enc.Encode(set)
}
