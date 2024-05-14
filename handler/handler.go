package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/la0rg/lingo/config"
	"github.com/la0rg/lingo/model"
)

type CardService interface {
	Get(ctx context.Context, collection *model.Collection, word string) (*model.Card, error)
}

type APKGService interface {
	CreateOnS3(ctx context.Context, collection *model.Collection, words ...string) (string, error)
}

type handlers struct {
	conf *config.Config
	card CardService
	apkg APKGService
}

func NewRouter(conf *config.Config, card CardService, apkg APKGService) http.Handler {
	handlers := handlers{conf: conf, card: card, apkg: apkg}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(RequestLogger)
	r.Use(Recoverer)
	r.Use(APIKey(conf))

	r.Route("/collections/{collection}", func(r chi.Router) {
		r.Get("/cards/{word}", ErrorHandler(handlers.Card))
		r.Post("/apkg", ErrorHandler(handlers.APKG))
	})

	return r
}
