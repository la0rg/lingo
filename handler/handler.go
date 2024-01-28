package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/la0rg/lingo/apkg"
	"github.com/la0rg/lingo/model"
)

type CardsGenerator interface {
	Generate(ctx context.Context, sourceLanguage, targetLanguage string, words ...string) (*model.Cards, error)
}

type APKGService interface {
	Create(ctx context.Context, cards *model.Cards) (*apkg.Collection, error)
}

type handlers struct {
	cards CardsGenerator
	apkg  APKGService
}

func NewRouter(cards CardsGenerator, apkg APKGService) http.Handler {
	handlers := handlers{cards: cards, apkg: apkg}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(RequestLogger)
	r.Use(Recoverer)

	r.Route("/collections/{collection}", func(r chi.Router) {
		r.Get("/cards/{card}", ErrorHandler(handlers.Card))
		r.Get("/apkg", ErrorHandler(handlers.APKG))
	})

	return r
}
