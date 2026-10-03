package views

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
)

type Page struct {
	htmlTpl *template.Template
}

func (p Page) Execute(w http.ResponseWriter, r *http.Request, data any) {
	w.Header().Set("Content-Type", "text/html")
	var buf bytes.Buffer
	err := p.htmlTpl.Execute(&buf, data)
	if err != nil {
		slog.Error("executing template", "error", err)
		http.Error(w, "There was an error executing the template.", http.StatusInternalServerError)
		return
	}

	io.Copy(w, &buf)
}

func Must(p Page, err error) Page {
	if err != nil {
		panic(err)
	}
	return p
}

// ParseFS parses patterns into a page. The page's category, the directory of
// the first pattern ("home" for "home/index.gohtml"), is bound once here as the
// {{ category }} template function. {{ assetVersion }} returns assetVersion,
// which templates append to static asset URLs so each release busts the cache.
func ParseFS(assetVersion string, fsys fs.FS, patterns ...string) (Page, error) {
	if len(patterns) == 0 {
		return Page{}, errors.New("parsing template: no patterns")
	}
	category, _, _ := strings.Cut(patterns[0], "/")
	tpl, err := template.New(path.Base(patterns[0])).Funcs(template.FuncMap{
		"category":     func() string { return category },
		"assetVersion": func() string { return assetVersion },
	}).ParseFS(fsys, patterns...)
	if err != nil {
		return Page{}, fmt.Errorf("parsing template: %w", err)
	}

	return Page{htmlTpl: tpl}, nil
}
