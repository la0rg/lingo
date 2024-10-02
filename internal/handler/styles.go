package handler

import (
	"net/http"

	"github.com/la0rg/lingo/web"
)

func (h *handlers) Styles(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(web.Styles)
	return err
}
