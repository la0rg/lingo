//go:build manual

package apkg

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	model := Model{
		ID:   555555,
		Name: "lingo_test",
		Css:  `.word { \n\tfont-size: 28px; \nfont-style: italic; \n}`,
		Templates: []Template{
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

	notes := []Note{
		{ID: 1111111, Model: model, Fields: []string{"world", "welt"}},
		{ID: 1111112, Model: model, Fields: []string{"red", "Rot"}},
	}

	collection := Collection{
		Decks: []Deck{{
			ID:     58098085085,
			Name:   "lingo_test",
			Config: DefaultConfig(6493485093845, "lingo_test"),
			Notes:  notes,
		}},
	}

	result, err := Build(collection)
	assert.NoError(t, err)

	err = os.WriteFile("test.apkg", result, 0644)
	assert.NoError(t, err)
}
