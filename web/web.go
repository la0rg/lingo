package web

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed css/styles.css
var cssFS embed.FS

type TemplateName string

const (
	Index TemplateName = "index.html"
)

type Templates struct {
	templates *template.Template
}

func NewTemplates() (*Templates, error) {
	templates, err := template.ParseFS(templateFS, "templates/*.html")
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

func Styles() ([]byte, error) {
	body, err := cssFS.ReadFile("css/styles.css")
	if err != nil {
		return nil, fmt.Errorf("failed to read styles.css: %v", err)
	}

	return body, nil
}
