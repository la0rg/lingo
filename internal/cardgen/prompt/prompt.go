package prompt

import (
	"bytes"
	_ "embed"
	"text/template"
)

//go:embed system.md.tmpl
var systemTemplate string

var system = template.Must(template.New("system").Option("missingkey=error").Parse(systemTemplate))

// System renders the system prompt for card generation.
func System() (string, error) {
	var buf bytes.Buffer
	if err := system.Execute(&buf, nil); err != nil {
		return "", err
	}
	return buf.String(), nil
}
