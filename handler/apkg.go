package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/la0rg/lingo/apkg"
	"github.com/la0rg/lingo/logger"
	"github.com/la0rg/lingo/model"
)

func (h *handlers) APKG(w http.ResponseWriter, r *http.Request) error {
	collectionName := strings.ToLower(chi.URLParam(r, "collection"))

	collection, ok := collections[collectionName]
	if !ok {
		return ErrCollectionNotFound
	}

	var request model.APKGRequest
	// if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
	// 	return BadRequest("invalid request format")
	// }

	logger := logger.WithContext(r.Context()).With("collection", collectionName, "request", request)
	logger.Info("Creating an APKG file")

	cards, err := h.cards.Generate(r.Context(), collection.sourceLanguage, collection.targetLanguage, request.Words...)
	if err != nil {
		return fmt.Errorf("unable to generate cards: %w", err)
	}

	apkgCollectionn, err := h.apkg.Create(r.Context(), cards)
	if err != nil {
		return fmt.Errorf("unable to create APKG file: %w", err)
	}

	archive, err := apkg.Build(apkgCollectionn)
	if err != nil {
		return fmt.Errorf("unable to build APKG file: %w", err)
	}

	w.Header().Set("Content-Disposition", "attachment; filename=collection_de_en.apkg")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(archive)))
	w.Write(archive)

	return nil
}
