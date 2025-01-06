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
		c := &collections[i]

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
		DefaultWord:      "Bad",
		HTMLTemplateName: web.CardPreviewDE,
		Template: apkg.Model{
			ID:   1715808522,
			Name: "lingo_de_en",
			Css:  "data/styles_de.css",
			Templates: []apkg.Template{
				{Name: "lingo_de_en", Qfmt: "data/qfmt_de.html", Afmt: "data/afmt_de.html"},
				{Name: "lingo_de_en_reversed", Qfmt: "data/qfmt_reversed_de.html", Afmt: "data/afmt_reversed_de.html"},
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
				{Name: "lingo_de_en", Qfmt: "data/qfmt_de.html", Afmt: "data/afmt_de.html"},
				{Name: "lingo_de_en_reversed", Qfmt: "data/qfmt_reversed_de.html", Afmt: "data/afmt_reversed_de.html"},
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
				{Name: "lingo_ru_en", Qfmt: "data/qfmt_ru.html", Afmt: "data/afmt_ru.html"},
				{Name: "lingo_ru_en_reversed", Qfmt: "data/qfmt_reversed_ru.html", Afmt: "data/afmt_reversed_ru.html"},
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
				{Name: "lingo_ru_de", Qfmt: "data/qfmt_ru.html", Afmt: "data/afmt_ru.html"},
				{Name: "lingo_ru_de_reversed", Qfmt: "data/qfmt_reversed_ru.html", Afmt: "data/afmt_reversed_ru.html"},
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
}
