package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/la0rg/lingo/model"
)

type CardsGenerator interface {
	Generate(ctx context.Context, sourceLanguage, targetLanguage string, words ...string) (*model.Cards, error)
}

type handlers struct {
	cards CardsGenerator
}

func NewRouter(cards CardsGenerator) http.Handler {
	handlers := handlers{cards: cards}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(RequestLogger)
	r.Use(Recoverer)

	r.Route("/collections/{collection}", func(r chi.Router) {
		r.Get("/cards/{card}", ErrorHandler(handlers.Card))
	})

	return r
}
