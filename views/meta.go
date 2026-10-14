package views

// Meta is the per-page search and social metadata rendered by the base
// template's head. Path is the canonical path on the site and Image is a
// site-absolute image path; the template joins both to the canonical host
// (https://bitofbytes.io) so scrapers receive absolute URLs.
type Meta struct {
	Title       string
	Description string
	Path        string
	Image       string
}
