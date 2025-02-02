package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCard_FieldByName(t *testing.T) {
	card := Card{
		Word:   "word",
		Plural: "plural",
		Translations: Translations{
			{
				PartOfSpeech: "noun",
				Translations: []string{"translation1", "translation2"},
			},
			{
				PartOfSpeech: "verb",
				Translations: []string{"translation3", "translation4"},
			},
		},
		Article:            "article",
		Gender:             "gender",
		Transcription:      "transcription",
		Context:            "context",
		ContextTranslation: "context_translation",
		Definition:         "definition",
	}

	cases := []struct {
		field    CardField
		expected string
	}{
		{CardFieldWord, "word"},
		{CardFieldPlural, "plural"},
		{CardFieldTranslation, "translation1, translation2, translation3, translation4"},
		{CardFieldArticle, "article"},
		{CardFieldGender, "gender"},
		{CardFieldTranscription, "transcription"},
		{CardFieldContext, "context"},
		{CardFieldContextTranslation, "context_translation"},
		{CardFieldDefinition, "definition"},
	}

	for _, c := range cases {
		t.Run(string(c.field), func(t *testing.T) {
			assert.Equal(t, c.expected, card.FieldByName(c.field))
		})
	}
}
