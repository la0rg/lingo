package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/la0rg/lingo/apkg"
	"github.com/la0rg/lingo/model"
)

var de_en_model = apkg.Model{
	ID:   555555,
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
}

var apkgConfig = apkg.DefaultConfig(6493485093845, "lingo_config")

var ErrEmptyCards = fmt.Errorf("empty cards")

type APKG struct{}

func (a *APKG) Create(ctx context.Context, cards *model.Cards) (*apkg.Collection, error) {
	if cards.Cards == nil || len(cards.Cards) < 1 {
		return nil, ErrEmptyCards
	}

	notes := make([]apkg.Note, 0)
	id := time.Now().UnixMilli()

	for _, card := range cards.Cards {
		translations := make([]string, 0)
		for _, translation := range card.Translations {
			translations = append(translations, translation.Translations...)
		}

		notes = append(notes, apkg.Note{
			ID:     id, // TODO: save ID in DB
			Model:  de_en_model,
			Fields: []string{card.Word, strings.Join(translations, ", ")},
		})

		id++
	}

	return &apkg.Collection{
		Decks: []apkg.Deck{{
			ID:     58098085085,
			Name:   "lingo_de_en",
			Config: apkgConfig,
			Notes:  notes,
		}},
	}, nil
}
