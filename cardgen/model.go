package cardgen

import "github.com/la0rg/lingo/model"

type request struct {
	SourceLanguage string   `json:"source_language"`
	TargetLanguage string   `json:"target_language"`
	Words          []string `json:"words"`
}

type response struct {
	Cards []model.Card `json:"cards"`
}
