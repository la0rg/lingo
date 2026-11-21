package model

import (
	"fmt"
	"reflect"
	"strings"
)

type Card struct {
	ID                 int64        `json:"-" dynamodbav:"id" cardfield:"id"`
	Word               string       `json:"word" dynamodbav:"word" cardfield:"word"`
	CollectionID       int64        `json:"-" dynamodbav:"collection_id" cardfield:"collection_id"`
	Plural             string       `json:"plural" dynamodbav:"plural" cardfield:"plural"`
	Translations       Translations `json:"translations" dynamodbav:"translations" cardfield:"translation"`
	Article            string       `json:"article" dynamodbav:"article,omitempty" cardfield:"article"`
	Gender             string       `json:"gender" dynamodbav:"gender,omitempty" cardfield:"gender"`
	Transcription      string       `json:"transcription" dynamodbav:"transcription,omitempty" cardfield:"transcription"`
	Context            string       `json:"context" dynamodbav:"context,omitempty" cardfield:"context"`
	ContextTranslation string       `json:"context_translation" dynamodbav:"context_translation,omitempty" cardfield:"context_translation"` //nolint:lll
	Definition         string       `json:"definition" dynamodbav:"definition,omitempty" cardfield:"definition"`
}

func (c Card) FieldByName(name CardField) string {
	v := reflect.ValueOf(c)
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if tag := field.Tag.Get("cardfield"); tag == name {
			fieldValue := v.Field(i).Interface()

			if stringer, ok := fieldValue.(fmt.Stringer); ok {
				return stringer.String()
			}

			return v.Field(i).String()
		}
	}
	return ""
}

// GetAllResult is the outcome of fetching cards for a list of words.
type GetAllResult struct {
	Cards        []Card
	MissingWords []string
}

type Translations []Translation

func (t Translations) String() string {
	str := make([]string, 0)
	for _, translation := range t {
		str = append(str, translation.Translations...)
	}
	return strings.Join(str, ", ")
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
