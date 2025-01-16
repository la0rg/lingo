package web

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"path/filepath"
)

//go:embed src/templates/*
var templateFS embed.FS

//go:embed dist/styles.css
var Styles []byte

//go:embed public/*
var PublicFS embed.FS

type TemplateName string

const (
	Index              TemplateName = "index.html"
	Collection         TemplateName = "collection.html"
	CollectionDownload TemplateName = "collection_download.html"
	CardPreview        TemplateName = "card.html"
	CardPreviewDE      TemplateName = "card_de.html"
	CardPreviewRU      TemplateName = "card_ru.html"
	LanguageSelector   TemplateName = "language_selector.html"
	Error              TemplateName = "error.html"
)

type Templates struct {
	templates map[string]*template.Template
}

func NewTemplates() (*Templates, error) {
	templates := make(map[string]*template.Template)

	templateFiles, err := fs.Glob(templateFS, "src/templates/*/*.html")
	if err != nil {
		return nil, fmt.Errorf("unable to get page templates: %v", err)
	}

	for _, file := range templateFiles {
		name := filepath.Base(file)

		template, err := template.ParseFS(
			templateFS,
			"src/templates/base.html",
			"src/templates/partials/*.html",
			file,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to parse '%v' template: %v", name, err)
		}

		templates[name] = template
	}

	slog.Info("Templates are loaded", "count", len(templates))
	return &Templates{templates: templates}, nil
}

func (t *Templates) Render(name TemplateName, data any) ([]byte, error) {
	template, ok := t.templates[string(name)]
	if !ok {
		return nil, fmt.Errorf("template '%v' not found", name)
	}

	var buf bytes.Buffer
	err := template.ExecuteTemplate(&buf, string(name), data)
	if err != nil {
		return nil, fmt.Errorf("unable to render template: %v", err)
	}

	return buf.Bytes(), nil
}
