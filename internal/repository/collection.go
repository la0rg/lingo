package repository

import (
	"embed"

	"github.com/la0rg/lingo/apkg"
	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/web"
)

//go:embed data/*
var dataFS embed.FS

type Collection struct {
	m map[string]*model.Collection
}

func NewCollection() (*Collection, error) {
	m := make(map[string]*model.Collection, len(collections))
	var err error
	for i := range collections {
		// copy the definition, so the package-level one keeps file names and can be loaded again
		c := new(model.Collection)
		*c = collections[i]
		c.Template.Templates = append([]apkg.Template(nil), collections[i].Template.Templates...)

		// read data files
		c.Template.Css, err = readData(c.Template.Css)
		if err != nil {
			return nil, err
		}

		for i := range c.Template.Templates {
			c.Template.Templates[i].Qfmt, err = readData(c.Template.Templates[i].Qfmt)
			if err != nil {
				return nil, err
			}

			c.Template.Templates[i].Afmt, err = readData(c.Template.Templates[i].Afmt)
			if err != nil {
				return nil, err
			}
		}

		m[model.CollectionSlug(c.SourceLanguage, c.TargetLanguage)] = c
	}

	return &Collection{m}, nil
}

func (c *Collection) GetBySlug(slug string) *model.Collection {
	return c.m[slug]
}

func readData(name string) (string, error) {
	b, err := dataFS.ReadFile(name)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

var collections = []model.Collection{
	{
		ID:               1715808757,
		Name:             "Lingo DE-EN",
		SourceLanguage:   "de",
		TargetLanguage:   "en",
		DefaultWord:      "Apfel",
		HTMLTemplateName: web.CardPreviewDE,
		Template: apkg.Model{
			ID:   1715808522,
			Name: "lingo_de_en",
			Css:  "data/styles_de.css",
			Templates: []apkg.Template{
				{Name: "lingo_de_en", Qfmt: "data/qfmt_de.html", Afmt: "data/afmt.html"},
				{Name: "lingo_de_en_reversed", Qfmt: "data/qfmt_reversed.html", Afmt: "data/afmt_reversed_de.html"},
			},
			Fields: []string{
				model.CardFieldWord,
				model.CardFieldArticle,
				model.CardFieldTranscription,
				model.CardFieldPlural,
				model.CardFieldTranslation,
				model.CardFieldContext,
				model.CardFieldContextTranslation,
				model.CardFieldDefinition,
			},
		},
	},
	{
		ID:               1715808758,
		Name:             "Lingo DE-RU",
		SourceLanguage:   "de",
		TargetLanguage:   "ru",
		DefaultWord:      "Apfel",
		HTMLTemplateName: web.CardPreviewDE,
		Template: apkg.Model{
			ID:   1715808523,
			Name: "lingo_de_ru",
			Css:  "data/styles_de.css",
			Templates: []apkg.Template{
				{Name: "lingo_de_en", Qfmt: "data/qfmt_de.html", Afmt: "data/afmt.html"},
				{Name: "lingo_de_en_reversed", Qfmt: "data/qfmt_reversed.html", Afmt: "data/afmt_reversed_de.html"},
			},
			Fields: []string{
				model.CardFieldWord,
				model.CardFieldArticle,
				model.CardFieldTranscription,
				model.CardFieldPlural,
				model.CardFieldTranslation,
				model.CardFieldContext,
				model.CardFieldContextTranslation,
				model.CardFieldDefinition,
			},
		},
	},
	{
		ID:               1715808759,
		Name:             "Lingo RU-EN",
		SourceLanguage:   "ru",
		TargetLanguage:   "en",
		DefaultWord:      "Мир",
		HTMLTemplateName: web.CardPreviewRU,
		Template: apkg.Model{
			ID:   1715808524,
			Name: "lingo_ru_en",
			Css:  "data/styles_ru.css",
			Templates: []apkg.Template{
				{Name: "lingo_ru_en", Qfmt: "data/qfmt_ru.html", Afmt: "data/afmt.html"},
				{Name: "lingo_ru_en_reversed", Qfmt: "data/qfmt_reversed.html", Afmt: "data/afmt_reversed_ru.html"},
			},
			Fields: []string{
				model.CardFieldWord,
				model.CardFieldGender,
				model.CardFieldTranscription,
				model.CardFieldPlural,
				model.CardFieldTranslation,
				model.CardFieldContext,
				model.CardFieldContextTranslation,
				model.CardFieldDefinition,
			},
		},
	},
	{
		ID:               1715808760,
		Name:             "Lingo RU-DE",
		SourceLanguage:   "ru",
		TargetLanguage:   "de",
		DefaultWord:      "Мир",
		HTMLTemplateName: web.CardPreviewRU,
		Template: apkg.Model{
			ID:   1715808525,
			Name: "lingo_ru_de",
			Css:  "data/styles_ru.css",
			Templates: []apkg.Template{
				{Name: "lingo_ru_de", Qfmt: "data/qfmt_ru.html", Afmt: "data/afmt.html"},
				{Name: "lingo_ru_de_reversed", Qfmt: "data/qfmt_reversed.html", Afmt: "data/afmt_reversed_ru.html"},
			},
			Fields: []string{
				model.CardFieldWord,
				model.CardFieldGender,
				model.CardFieldTranscription,
				model.CardFieldPlural,
				model.CardFieldTranslation,
				model.CardFieldContext,
				model.CardFieldContextTranslation,
				model.CardFieldDefinition,
			},
		},
	},
	{
		ID:               1715808761,
		Name:             "Lingo EN-DE",
		SourceLanguage:   "en",
		TargetLanguage:   "de",
		DefaultWord:      "Book",
		HTMLTemplateName: web.CardPreview,
		Template: apkg.Model{
			ID:   1715808526,
			Name: "lingo_en_de",
			Css:  "data/styles.css",
			Templates: []apkg.Template{
				{Name: "lingo_en_de", Qfmt: "data/qfmt.html", Afmt: "data/afmt.html"},
				{Name: "lingo_en_de_reversed", Qfmt: "data/qfmt_reversed.html", Afmt: "data/afmt_reversed.html"},
			},
			Fields: []string{
				model.CardFieldWord,
				model.CardFieldTranscription,
				model.CardFieldPlural,
				model.CardFieldTranslation,
				model.CardFieldContext,
				model.CardFieldContextTranslation,
				model.CardFieldDefinition,
			},
		},
	},
	{
		ID:               1715808762,
		Name:             "Lingo EN-RU",
		SourceLanguage:   "en",
		TargetLanguage:   "ru",
		DefaultWord:      "Book",
		HTMLTemplateName: web.CardPreview,
		Template: apkg.Model{
			ID:   1715808527,
			Name: "lingo_en_ru",
			Css:  "data/styles.css",
			Templates: []apkg.Template{
				{Name: "lingo_en_ru", Qfmt: "data/qfmt.html", Afmt: "data/afmt.html"},
				{Name: "lingo_en_ru_reversed", Qfmt: "data/qfmt_reversed.html", Afmt: "data/afmt_reversed.html"},
			},
			Fields: []string{
				model.CardFieldWord,
				model.CardFieldTranscription,
				model.CardFieldPlural,
				model.CardFieldTranslation,
				model.CardFieldContext,
				model.CardFieldContextTranslation,
				model.CardFieldDefinition,
			},
		},
	},
}
