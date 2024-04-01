package model

type Cards struct {
	Cards []Card `json:"cards"`
}

type Card struct {
	Word               string        `json:"word" dynamodbav:"word"`
	Collection         string        `json:"collection" dynamodbav:"collection"`
	Translations       []Translation `json:"translations" dynamodbav:"translations"`
	Article            string        `json:"article" dynamodbav:"article,omitempty"`
	Transcription      string        `json:"transcription" dynamodbav:"transcription,omitempty"`
	Context            string        `json:"context" dynamodbav:"context,omitempty"`
	ContextTranslation string        `json:"context_translation" dynamodbav:"context_translation,omitempty"`
	Definition         string        `json:"definition" dynamodbav:"definition,omitempty"`
}

type Translation struct {
	PartOfSpeech string   `json:"pos" dynamodbav:"pos"`
	Translations []string `json:"translations" dynamodbav:"translations"`
}
