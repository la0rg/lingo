package handler

import (
	"net/http"

	"github.com/la0rg/lingo/web"
)

func (h *handlers) Collection(w http.ResponseWriter, r *http.Request) error {
	return h.HTML(w, r, web.Collection, defaultLanguageSelector)
}
