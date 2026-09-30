package views

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"

	csrf "filippo.io/csrf/gorilla"
)

type Page struct {
	category string
	htmlTpl  *template.Template
}

func (p Page) Execute(w http.ResponseWriter, r *http.Request, data any) {
	tpl, err := p.htmlTpl.Clone()
	if err != nil {
		slog.Error("cloning template", "error", err)
		http.Error(w, "There was an error rendering the page", http.StatusInternalServerError)
		return
	}
	tpl.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrf.TemplateField(r)
		},
		"category": func() string {
			return p.category
		},
	})
	w.Header().Set("Content-Type", "text/html")
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
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

func ParseFS(fs fs.FS, patterns ...string) (Page, error) {
	var page Page
	category, _, _ := strings.Cut(patterns[0], "/")
	tpl, err := template.New(path.Base(patterns[0])).Funcs(
		template.FuncMap{
			"csrfField": func() (template.HTML, error) {
				return "", fmt.Errorf("csrfField not implemented")
			},
			"category": func() string {
				return ""
			},
		},
	).ParseFS(fs, patterns...)

	if err != nil {
		return page, fmt.Errorf("parsing template: %w", err)
	}

	return Page{
		htmlTpl:  tpl,
		category: category,
	}, nil
}
