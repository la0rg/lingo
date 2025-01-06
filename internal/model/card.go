package model

import "strings"

type Card struct {
	ID                 int64         `json:"-" dynamodbav:"id"`
	Word               string        `json:"word" dynamodbav:"word"`
	CollectionID       int64         `json:"-" dynamodbav:"collection_id"`
	Plural             string        `json:"plural" dynamodbav:"plural"`
	Translations       []Translation `json:"translations" dynamodbav:"translations"`
	Article            string        `json:"article" dynamodbav:"article,omitempty"`
	Gender             string        `json:"gender" dynamodbav:"gender,omitempty"`
	Transcription      string        `json:"transcription" dynamodbav:"transcription,omitempty"`
	Context            string        `json:"context" dynamodbav:"context,omitempty"`
	ContextTranslation string        `json:"context_translation" dynamodbav:"context_translation,omitempty"`
	Definition         string        `json:"definition" dynamodbav:"definition,omitempty"`
}

func (c Card) AllTranslations() string {
	translations := make([]string, 0)
	for _, translation := range c.Translations {
		translations = append(translations, translation.Translations...)
	}
	return strings.Join(translations, ", ")
}

func (c Card) FieldByName(name CardField) string {
	switch name {
	case CardFieldWord:
		return c.Word
	case CardFieldArticle:
		return c.Article
	case CardFieldGender:
		return c.Gender
	case CardFieldTranscription:
		return c.Transcription
	case CardFieldPlural:
		return c.Plural
	case CardFieldTranslation:
		return c.AllTranslations()
	case CardFieldContext:
		return c.Context
	case CardFieldContextTranslation:
		return c.ContextTranslation
	case CardFieldDefinition:
		return c.Definition
	default:
		return ""
	}
}

type Translation struct {
	PartOfSpeech string   `json:"pos" dynamodbav:"pos,omitempty"`
	Translations []string `json:"translations" dynamodbav:"translations"`
}

type CardField = string

const (
	CardFieldWord               CardField = "word"
	CardFieldArticle            CardField = "article"
	CardFieldGender             CardField = "gender"
	CardFieldTranscription      CardField = "transcription"
	CardFieldPlural             CardField = "plural"
	CardFieldTranslation        CardField = "translation"
	CardFieldContext            CardField = "context"
	CardFieldContextTranslation CardField = "context_translation"
	CardFieldDefinition         CardField = "definition"
)
