package views

// SiteURL is the canonical origin. Traefik redirects www to the apex, so the
// apex is the only host a canonical or Open Graph URL names.
const SiteURL = "https://bitofbytes.io"

// defaultImage is the preview image for pages without a screenshot.
const defaultImage = "/static/apple-touch-icon.png"

// Meta is the per-page SEO data base.gohtml renders into <head>. Page data
// structs embed it.
type Meta struct {
	Title       string
	Description string
	// Path is the canonical path on the site, without a query string.
	Path string
	// Image is a site-relative image path; empty means the default icon.
	Image string
	// JSONLD is a schema.org object rendered as a JSON-LD data block, or nil
	// for none.
	JSONLD any
}

// Canonical is the page's absolute canonical URL.
func (m Meta) Canonical() string {
	return SiteURL + m.Path
}

// ImageURL is the absolute preview image URL, which Open Graph requires.
func (m Meta) ImageURL() string {
	if m.Image == "" {
		return SiteURL + defaultImage
	}
	return SiteURL + m.Image
}

// TwitterCard is the large card when the page has its own screenshot, and the
// small summary card for the default icon.
func (m Meta) TwitterCard() string {
	if m.Image == "" {
		return "summary"
	}
	return "summary_large_image"
}
