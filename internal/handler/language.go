package handler

import (
	"net/http"

	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/web"
)

func (h *handlers) Language(w http.ResponseWriter, r *http.Request) error {
	return h.HTML(w, r, web.LanguageSelector, model.Languages)
}
