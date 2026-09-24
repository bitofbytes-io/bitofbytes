package controllers

import (
	"net/http"
	"slices"
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
	Updates     []ProjectView
	Keys        []OctaveKey
	BlackKeys   []BlackKey
	KeyCount    string
	RecentCount int
	Activities  models.Activities
}

type ProjectsIndexData struct {
	Projects []ProjectView
	Sort     string
}

type ProjectDetailData struct {
	Project ProjectView
}

const (
	homeUpdateCount = 4
	octaveSize      = 8
	sortUpdated     = "updated"
	sortNewest      = "newest"
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

	p.Templates.Home.Execute(w, r, HomeData{
		Updates:     byUpdate[:min(len(byUpdate), homeUpdateCount)],
		Keys:        keys,
		BlackKeys:   black,
		KeyCount:    countWords[len(keys)],
		RecentCount: recent,
		Activities:  p.Activities,
	})
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
