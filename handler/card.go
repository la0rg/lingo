package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/la0rg/lingo/model"
)

var ErrCollectionNotFound = fmt.Errorf("collection not found")

func (h *handlers) Card(w http.ResponseWriter, r *http.Request) error {
	collectionName := strings.ToLower(chi.URLParam(r, "collection"))

	// TODO: replace with collection db lookup
	collection, ok := model.Collections[collectionName]
	if !ok {
		return ErrCollectionNotFound
	}

	word := strings.TrimSpace(chi.URLParam(r, "word"))
	if len(word) == 0 {
		return BadRequest("empty word")
	}

	card, err := h.card.Get(r.Context(), collection, word)
	if err != nil {
		return err
	}

	JSON(w, r, card)
	return nil
}
