package controllers

import (
	"net/http"

	"github.com/DryWaters/bitofbytes/models"
	"github.com/DryWaters/bitofbytes/views"
)

type Portfolio struct {
	Projects  []models.Project
	Templates PortfolioTemplates
}

type PortfolioTemplates struct {
	Home          views.Page
	ProjectsIndex views.Page
	ProjectDetail views.Page
}

type HomeData struct{}

type ProjectsIndexData struct {
	Projects []models.Project
}

type ProjectDetailData struct {
	Project models.Project
}

func (p Portfolio) Home(w http.ResponseWriter, r *http.Request) {
	p.Templates.Home.Execute(w, r, HomeData{})
}

func (p Portfolio) ProjectsIndex(w http.ResponseWriter, r *http.Request) {
	p.Templates.ProjectsIndex.Execute(w, r, ProjectsIndexData{
		Projects: p.Projects,
	})
}

func (p Portfolio) ProjectDetail(w http.ResponseWriter, r *http.Request) {
	project, ok := p.findProject(r.PathValue("slug"))
	if !ok {
		http.NotFound(w, r)
		return
	}

	p.Templates.ProjectDetail.Execute(w, r, ProjectDetailData{
		Project: project,
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
