package handler

import (
	"net/http"

	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/web"
)

func (h *handlers) Language(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return BadRequest("Invalid request format")
	}

	source, target := r.Form.Get("source"), r.Form.Get("target")
	if source == "" || target == "" {
		return ErrInvalidLanguage
	}

	sourceNew, targetNew := r.Form.Get("sourceNew"), r.Form.Get("targetNew")
	languageSelector := model.UpdateLanguageSelector(
		source, target, sourceNew, targetNew,
	)
	return h.HTML(w, r, web.LanguageSelector, languageSelector)
}
