package cardgen

type request struct {
	SourceLanguage string   `json:"source_language"`
	TargetLanguage string   `json:"target_language"`
	Words          []string `json:"words"`
}
