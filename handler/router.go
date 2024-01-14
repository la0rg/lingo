package handler

import (
	"context"
	"net/http"

	"github.com/la0rg/lingo/cardgen"
)

type generator interface {
	Generate(ctx context.Context, sourceLanguage, targetLanguage string, words ...string) (*cardgen.Cards, error)
}

type handlers struct {
	cards generator
}

func Router(cards generator) http.Handler {
	handlers := handlers{cards: cards}

	mux := http.NewServeMux()

	mux.HandleFunc("/card", handlers.GetCardHandler)

	return mux
}
