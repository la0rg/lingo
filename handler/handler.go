package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/la0rg/lingo/cardgen"
)

type CardsGenerator interface {
	Generate(ctx context.Context, sourceLanguage, targetLanguage string, words ...string) (*cardgen.Cards, error)
}

type handlers struct {
	cards CardsGenerator
}

func NewRouter(cards CardsGenerator) http.Handler {
	handlers := handlers{cards: cards}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(RequestLogger)
	r.Use(middleware.Recoverer)

	r.HandleFunc("/card/{card_id}", ErrorHandler(handlers.GetCardHandler))

	return r
}
