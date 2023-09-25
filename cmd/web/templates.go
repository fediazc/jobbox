package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"cazar.fediaz.net/internal/models"
	"cazar.fediaz.net/ui"
)

type templateData struct {
	IsAuthenticated bool
	Flash           string
	Form            any
	Job             *models.Job
	Jobs            []*models.Job
}

func todaysDate() string {
	return time.Now().Format("2006-01-02")
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("01/02/2006")
}

func humanShortDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("01/02")
}

var functions = template.FuncMap{
	"todaysDate":     todaysDate,
	"humanDate":      humanDate,
	"humanShortDate": humanShortDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Files, "html/pages/*.gohtml")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/base.gohtml",
			"html/partials/*.gohtml",
			page,
		}

		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
