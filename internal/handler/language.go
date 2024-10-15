package handler

import (
	"fmt"
	"net/http"

	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/web"
)

func (h *handlers) Language(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("unable to parse language selector form: %v", err)
	}

	sourceActive, targetActive := r.Form.Get("sourceActive"), r.Form.Get("targetActive")
	if sourceActive == "" || targetActive == "" {
		return BadRequest("language selector: source or target language is empty")
	}

	sourceNew, targetNew := r.Form.Get("sourceNew"), r.Form.Get("targetNew")
	languageSelector := model.UpdateLanguageSelector(
		sourceActive, targetActive, sourceNew, targetNew,
	)
	return h.HTML(w, r, web.LanguageSelector, languageSelector)
}
