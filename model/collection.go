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
			Css: `
			.card { font-family: Arial,sans-serif; text-align: center; background-color: #F5F2E7; }
        	.word { font-size: 28px; color: #582177; }
        	.transc { font-size: 16px; color: #317FC0; font-family: Arial Unicode MS; }
        	.transl { font-size: 22px; color: #582177; margin: 0 0 10px; }
        	.mid { font-size: 17px; color: #222222; margin: 10px 0 0; }
        	.small { font-size: 14px; font-style: italic; color: #111111; margin: 10px 0 0; }
			.word.der { color:#262e7d }
			.word.die { color:#be3737 }
			.word.das { color:#34875d }
			`,
			Templates: []apkg.Template{
				{
					Name: "lingo",
					Qfmt: `
					<div class=\"word {{article}}\">{{article}} {{word}}</div>
					<div class=transc>[{{transctiption}}]</div>
					<div class=mid>{{context}}</div><br/>
					<div class=small>{{definition}}</div>`,
					Afmt: `
					{{FrontSide}}
        			<hr id=answer>
        			<div class=transl>{{translation}}</div>
        			<div class=mid>{{context_translation}}</div>`,
				},
				{
					Name: "lingo (reversed)",
					Qfmt: `<hr id=answer><div class=word>{{translation}}</div><div class=small>{{context_translation}}</div>`,
					Afmt: `{{FrontSide}}<div class=word>{{article}} {{word}}</div><div hidden></div><div class=mid>{{plural}}</div><div class=small>{{context}}</div>`,
				},
			},
			Fields: []string{"word", "article", "transctiption", "plural", "translation", "context", "context_translation", "definition"},
		},
	},
}
