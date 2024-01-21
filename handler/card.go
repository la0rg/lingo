package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/la0rg/lingo/logger"
)

type Collection struct {
	sourceLanguage string
	targetLanguage string
}

var collections = map[string]Collection{
	"de-en": {sourceLanguage: "de", targetLanguage: "en"},
}

var ErrCollectionNotFound = fmt.Errorf("collection not found")

func (h *handlers) GetCard(w http.ResponseWriter, r *http.Request) error {
	collectionName := strings.ToLower(chi.URLParam(r, "collection"))

	collection, ok := collections[collectionName]
	if !ok {
		return ErrCollectionNotFound
	}

	card := strings.ToLower(chi.URLParam(r, "card"))

	logger.WithContext(r.Context()).Info("Get card", "collection", collectionName, "card", card)

	result, err := h.cards.Generate(r.Context(), collection.sourceLanguage, collection.targetLanguage, card)
	if err != nil {
		return fmt.Errorf("unable to generate cards: %v", err)
	}

	JSON(w, r, result)
	return nil
}
