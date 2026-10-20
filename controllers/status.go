package controllers

import (
	"net/http"

	"github.com/DryWaters/bitofbytes/internal/status"
	"github.com/DryWaters/bitofbytes/views"
)

type Status struct {
	Snapshot func() []status.Entry
	Page     views.Page
}

func (s Status) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	s.Page.Execute(w, r, struct {
		views.Meta
		Projects []status.Entry
	}{
		Meta:     views.Meta{Title: "Status | " + siteName, Description: "Project availability.", Path: "/status"},
		Projects: s.Snapshot(),
	})
}
