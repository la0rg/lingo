package apkg

import (
	"time"

	"github.com/la0rg/lingo/apkg/apkgdb"
)

type Model struct {
	ID        int64
	Name      string
	Css       string
	Templates []Template
	Fields    []string
}

func (m Model) buildModel(deckID int64) apkgdb.Model {
	return apkgdb.Model{
		ID:        m.ID,
		Name:      m.Name,
		Mod:       time.Now().Unix(),
		Usn:       1,
		Did:       deckID,
		Css:       apkgdb.UnescapedString(m.Css),
		Templates: m.buildTemplates(),
		Fields:    m.buildFields(),
		Tags:      []string{},
	}
}

func (m Model) buildTemplates() []apkgdb.Template {
	var templates = make([]apkgdb.Template, 0, len(m.Templates))

	for _, template := range m.Templates {
		templates = append(templates, apkgdb.Template{
			Name: template.Name,
			Ord:  len(templates),
			Qfmt: apkgdb.UnescapedString(template.Qfmt),
			Afmt: apkgdb.UnescapedString(template.Afmt),
		})
	}
	return templates
}

func (m Model) buildFields() []apkgdb.Field {
	var fields = make([]apkgdb.Field, 0, len(m.Fields))

	for _, name := range m.Fields {
		fields = append(fields, apkgdb.Field{
			Name: name,
			Ord:  len(fields),
			Font: "Arial",
			Size: 20,
		})
	}
	return fields
}
