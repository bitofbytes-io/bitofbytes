package controllers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/DryWaters/bitofbytes/models"
	"github.com/DryWaters/bitofbytes/views"
)

type Portfolio struct {
	Projects   []models.Project
	Activities models.Activities
	Templates  PortfolioTemplates
	// Now returns the current time; nil means time.Now. Tests pin it.
	Now func() time.Time
}

type PortfolioTemplates struct {
	Home          views.Page
	ProjectsIndex views.Page
	ProjectDetail views.Page
}

// ProjectView is a project plus what depends on the current date.
type ProjectView struct {
	models.Project
	Recent bool
}

// OctaveKey is one white key of the home page keyboard.
type OctaveKey struct {
	ProjectView
	Note string
}

// BlackKey is a decorative black key sitting on white-key boundary Pos.
type BlackKey struct {
	Pos int
}

type HomeData struct {
	views.Meta
	// JSONLD is the pre-rendered JSON-LD data block for the page, emitted
	// verbatim in the head. It is built from models' data only, so wrapping
	// it as template.JS is safe and keeps html/template from JS-escaping it.
	JSONLD template.JS
	// Featured is the most recently updated project with a screenshot, other
	// than this site itself; HasFeatured is false when there is none.
	Featured    ProjectView
	HasFeatured bool
	// Updates are the most recently updated projects other than Featured,
	// newest first; they may be newer than Featured.
	Updates   []ProjectView
	Keys      []OctaveKey
	BlackKeys []BlackKey
	// KeyCount is the number of keys as a capitalized word ("Eight").
	KeyCount    string
	RecentCount int
	// RecentLabel is RecentCount with its noun ("1 update", "6 updates").
	RecentLabel string
	// Activities are the "now" lines.
	Activities models.Activities
}

type ProjectsIndexData struct {
	views.Meta
	JSONLD   template.JS
	Projects []ProjectView
	Sort     string
}

type ProjectDetailData struct {
	views.Meta
	JSONLD  template.JS
	Project ProjectView
}

const (
	homeUpdateCount = 3
	// siteSlug is this site's own project; Home never features a screenshot
	// of itself.
	siteSlug    = "bitofbytes"
	octaveSize  = 8
	sortUpdated = "updated"
	sortNewest  = "newest"
	// canonicalHost is the apex origin the site is served from (Traefik
	// redirects www to the apex), hardcoded for canonical and og: URLs.
	canonicalHost = "https://bitofbytes.io"
)

var (
	whiteNotes = []string{"C4", "D4", "E4", "F4", "G4", "A4", "B4", "C5"}
	blackKeys  = []BlackKey{{1}, {2}, {4}, {5}, {6}}
	countWords = []string{"No", "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight"}
)

func (p Portfolio) Home(w http.ResponseWriter, r *http.Request) {
	byUpdate := p.views(models.SortByLastUpdate(p.Projects))

	recent := 0
	for _, project := range byUpdate {
		if project.Recent {
			recent++
		}
	}

	// The keyboard holds one octave: the most recently updated projects,
	// laid out in the order they were started.
	onKeys := byUpdate[:min(len(byUpdate), octaveSize)]
	started := make([]models.Project, len(onKeys))
	for i, project := range onKeys {
		started[i] = project.Project
	}
	keys := make([]OctaveKey, 0, len(onKeys))
	for i, project := range p.views(models.SortByFirstCommit(started)) {
		keys = append(keys, OctaveKey{ProjectView: project, Note: whiteNotes[i]})
	}
	var black []BlackKey
	for _, key := range blackKeys {
		if key.Pos < len(keys) {
			black = append(black, key)
		}
	}

	featured, hasFeatured, updates := latestUpdates(byUpdate)

	image := "/static/projects/bitofbytes/home.webp"
	if hasFeatured {
		image = featured.Thumbnail().Path
	}

	p.Templates.Home.Execute(w, r, HomeData{
		Meta: views.Meta{
			Title:       "BitOfBytes | Daniel Waters",
			Description: "The personal landing page and project portfolio for Daniel Waters: real-time video at work, small apps at home.",
			Path:        "/",
			Image:       image,
		},
		JSONLD:      personJSONLD(),
		Featured:    featured,
		HasFeatured: hasFeatured,
		Updates:     updates,
		Keys:        keys,
		BlackKeys:   black,
		KeyCount:    countWords[len(keys)],
		RecentCount: recent,
		RecentLabel: pluralize(recent, "update", "updates"),
		Activities:  p.Activities,
	})
}

// latestUpdates splits projects, most recently updated first, into the
// featured project (the newest one with a screenshot that is not this site)
// and the ledger: the homeUpdateCount newest projects other than the featured
// one, so a project newer than Featured (such as this site) still appears.
func latestUpdates(byUpdate []ProjectView) (featured ProjectView, ok bool, ledger []ProjectView) {
	for _, project := range byUpdate {
		if project.Slug != siteSlug && project.Thumbnail().Path != "" {
			featured, ok = project, true
			break
		}
	}
	for _, project := range byUpdate {
		if len(ledger) == homeUpdateCount {
			break
		}
		if ok && project.Slug == featured.Slug {
			continue
		}
		ledger = append(ledger, project)
	}
	return featured, ok, ledger
}

func pluralize(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return strconv.Itoa(n) + " " + many
}

func (p Portfolio) ProjectsIndex(w http.ResponseWriter, r *http.Request) {
	sort := sortUpdated
	projects := models.SortByLastUpdate(p.Projects)
	if r.URL.Query().Get("sort") == sortNewest {
		sort = sortNewest
		projects = models.SortByFirstCommit(p.Projects)
		slices.Reverse(projects)
	}

	p.Templates.ProjectsIndex.Execute(w, r, ProjectsIndexData{
		Meta: views.Meta{
			Title:       "Projects | BitOfBytes",
			Description: "Small, private apps for everyday life, plus this site. Each one has its own page with screenshots and what changed most recently.",
			Path:        "/projects",
			Image:       "/static/projects/bitofbytes/projects.webp",
		},
		JSONLD:   personJSONLD(),
		Projects: p.views(projects),
		Sort:     sort,
	})
}

func (p Portfolio) ProjectDetail(w http.ResponseWriter, r *http.Request) {
	project, ok := p.findProject(r.PathValue("slug"))
	if !ok {
		http.NotFound(w, r)
		return
	}

	p.Templates.ProjectDetail.Execute(w, r, ProjectDetailData{
		Meta: views.Meta{
			Title:       project.Name + " | BitOfBytes",
			Description: project.Tagline,
			Path:        "/projects/" + project.Slug,
			Image:       project.Thumbnail().Path,
		},
		JSONLD:  projectJSONLD(project),
		Project: p.view(project),
	})
}

func (p Portfolio) findProject(slug string) (models.Project, bool) {
	for _, project := range p.Projects {
		if project.Slug == slug {
			return project, true
		}
	}

	return models.Project{}, false
}

func (p Portfolio) now() time.Time {
	if p.Now != nil {
		return p.Now()
	}
	return time.Now()
}

func (p Portfolio) view(project models.Project) ProjectView {
	return ProjectView{
		Project: project,
		Recent:  project.UpdatedWithin(p.now(), models.RecentWindow),
	}
}

func (p Portfolio) views(projects []models.Project) []ProjectView {
	out := make([]ProjectView, len(projects))
	for i, project := range projects {
		out[i] = p.view(project)
	}
	return out
}

// personJSONLD is the site owner's schema.org Person block, served as a data
// block on the home and projects index pages. Values are static/constants;
// it is safe to emit as template.JS.
func personJSONLD() template.JS {
	const block = `{
  "@context": "https://schema.org",
  "@type": "Person",
  "name": "Daniel Waters",
  "url": "https://bitofbytes.io/",
  "jobTitle": "Senior Software Engineer",
  "worksFor": { "@type": "Organization", "name": "Cisco" },
  "sameAs": [
    "https://github.com/bitofbytes-io",
    "https://www.linkedin.com/in/daniel-waters/"
  ]
}`
	return template.JS(block)
}

// projectJSONLD is a schema.org SoftwareApplication block for a project
// detail page, built from models' data only (no request input), so emitting
// it as template.JS is safe.
func projectJSONLD(project models.Project) template.JS {
	type softwareApp struct {
		Context             string `json:"@context"`
		Type                string `json:"@type"`
		Name                string `json:"name"`
		Description         string `json:"description"`
		URL                 string `json:"url"`
		Image               string `json:"image,omitempty"`
		ApplicationCategory string `json:"applicationCategory"`
		OperatingSystem     string `json:"operatingSystem"`
		DateModified        string `json:"dateModified,omitempty"`
		Author              struct {
			Type string `json:"@type"`
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"author"`
	}

	app := softwareApp{
		Context:             "https://schema.org",
		Type:                "SoftwareApplication",
		Name:                project.Name,
		Description:         project.Tagline,
		URL:                 canonicalHost + "/projects/" + project.Slug,
		ApplicationCategory: "WebApplication",
		OperatingSystem:     "Web",
		Author: struct {
			Type string `json:"@type"`
			Name string `json:"name"`
			URL  string `json:"url"`
		}{
			Type: "Person",
			Name: "Daniel Waters",
			URL:  canonicalHost + "/",
		},
	}
	if image := project.Thumbnail().Path; image != "" {
		app.Image = canonicalHost + image
	}
	if updated := project.UpdatedOn(); !updated.IsZero() {
		app.DateModified = updated.Format("2006-01-02")
	}

	out, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		return template.JS(`{"@context":"https://schema.org","@type":"SoftwareApplication","name":"BitOfBytes"}`)
	}
	return template.JS(out)
}
