package web

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
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
	LanguageSelector   TemplateName = "language_selector.html"
	Error              TemplateName = "error.html"
)

type Templates struct {
	templates *template.Template
}

func NewTemplates() (*Templates, error) {
	templates, err := template.ParseFS(templateFS, "src/templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("unable to parse html templates: %v", err)
	}

	return &Templates{templates: templates}, nil
}

func (t *Templates) Render(name TemplateName, data any) ([]byte, error) {
	var buf bytes.Buffer
	err := t.templates.ExecuteTemplate(&buf, string(name), data)
	if err != nil {
		return nil, fmt.Errorf("unable to render template: %v", err)
	}

	return buf.Bytes(), nil
}
