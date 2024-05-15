package model

import "github.com/la0rg/lingo/apkg"

type Collection struct {
	ID             int64
	Slug           string
	Name           string
	SourceLanguage string
	TargetLanguage string
	Template       apkg.Model
}

var Collections = map[string]*Collection{
	"de-en": {
		ID:             1715808757,
		Name:           "Lingo DE-EN",
		Slug:           "de-en",
		SourceLanguage: "de",
		TargetLanguage: "en",
		Template: apkg.Model{
			ID:   1715808518,
			Name: "lingo_de_en",
			Css:  `.word { \n\tfont-size: 28px; \nfont-style: italic; \n}`,
			Templates: []apkg.Template{
				{
					Name: "lingo",
					Qfmt: `<div class=word>{{word}}</div>`,
					Afmt: `{{FrontSide}}\n\n<div class=word>{{translation}}</div>`,
				},
				{
					Name: "lingo (reversed)",
					Qfmt: `<div class=word>{{translation}}</div>`,
					Afmt: `{{FrontSide}}\n\n<div class=word>{{word}}</div>`,
				},
			},
			Fields: []string{"word", "translation"},
		},
	},
}
