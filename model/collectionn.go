package model

type Collection struct {
	ID             string
	SourceLanguage string
	TargetLanguage string
}

var Collections = map[string]*Collection{
	"de-en": {
		ID:             "de-en",
		SourceLanguage: "de",
		TargetLanguage: "en",
	},
	"de-ru": {
		ID:             "de-ru",
		SourceLanguage: "de",
		TargetLanguage: "ru",
	},
}
