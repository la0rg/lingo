package model

import (
	"github.com/la0rg/lingo/pkg/apkg"
)

type Collection struct {
	ID             int64
	Name           string
	SourceLanguage string
	TargetLanguage string
	Template       apkg.Model
}

func CollectionSlug(source, target string) string {
	return source + "-" + target
}

var Collections = []Collection{
	{
		ID:             1715808757,
		Name:           "Lingo DE-EN",
		SourceLanguage: "de",
		TargetLanguage: "en",
		Template: apkg.Model{
			ID:        1715808518,
			Name:      "lingo_de_en",
			Css:       deStyles,
			Templates: deTemplates,
			Fields:    deFields,
		},
	},
	{
		ID:             1715808758,
		Name:           "Lingo DE-RU",
		SourceLanguage: "de",
		TargetLanguage: "ru",
		Template: apkg.Model{
			ID:        1715808519,
			Name:      "lingo_de_ru",
			Css:       deStyles,
			Templates: deTemplates,
			Fields:    deFields,
		},
	},
}

var CollectionsMap = buildCollectionsMap(Collections)

var (
	deStyles = `
	.card { font-family: Arial,sans-serif; text-align: center; background-color: #F5F2E7; }
	.word { font-size: 28px; color: #582177; }
	.transc { font-size: 16px; color: #317FC0; font-family: Arial Unicode MS; }
	.transl { font-size: 22px; color: #582177; margin: 0 0 10px; }
	.mid { font-size: 17px; color: #222222; margin: 10px 0 0; }
	.small { font-size: 14px; font-style: italic; color: #111111; margin: 10px 0 0; }
	.word.der { color:#262e7d }
	.word.die { color:#be3737 }
	.word.das { color:#34875d }`

	deTemplates = []apkg.Template{
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
			Qfmt: `
		<div class=transl>{{translation}}</div>
		<div class=mid>{{context_translation}}</div>`,
			Afmt: `
		{{FrontSide}}
		<hr id=answer>
		<div class=\"word {{article}}\">{{article}} {{word}}</div>
		<div class=transc>[{{transctiption}}]</div>
		<div class=mid>{{context}}</div><br/>
		<div class=small>{{definition}}</div>`,
		},
	}

	deFields = []string{"word", "article", "transctiption", "plural", "translation", "context", "context_translation", "definition"}
)

func buildCollectionsMap(collections []Collection) map[string]*Collection {
	m := make(map[string]*Collection, len(collections))
	for i := range collections {
		c := &collections[i]
		m[CollectionSlug(c.SourceLanguage, c.TargetLanguage)] = c
	}
	return m
}
