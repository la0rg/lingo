package cardgen

type request struct {
	SourceLanguage string   `json:"source_language"`
	TargetLanguage string   `json:"target_language"`
	Words          []string `json:"words"`
}

type Cards struct {
	Cards []Card `json:"cards"`
}

type Card struct {
	Word               string        `json:"word"`
	Translations       []Translation `json:"translations"`
	Article            string        `json:"article"`
	Transcription      string        `json:"transcription"`
	Context            string        `json:"context"`
	ContextTranslation string        `json:"context_translation"`
	Definition         string        `json:"definition"`
}

type Translation struct {
	PartOfSpeech string   `json:"pos"`
	Translations []string `json:"translations"`
}
