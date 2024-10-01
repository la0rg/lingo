package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/la0rg/lingo/config"
	"github.com/la0rg/lingo/model"
	"github.com/la0rg/lingo/web"
)

type CardService interface {
	Get(ctx context.Context, collection *model.Collection, word string) (*model.Card, error)
}

type APKGService interface {
	CreateOnS3(ctx context.Context, collection *model.Collection, words ...string) (string, error)
}

type TemplateRenderer interface {
	Render(name web.TemplateName, data any) ([]byte, error)
}

type handlers struct {
	conf *config.Config
	card CardService
	apkg APKGService
	tmpl TemplateRenderer
}

func NewRouter(
	conf *config.Config,
	card CardService,
	apkg APKGService,
	tmpl TemplateRenderer,
) http.Handler {
	handlers := handlers{conf: conf, card: card, apkg: apkg, tmpl: tmpl}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(RequestLogger)
	r.Use(Recoverer)

	r.Route("/collections/{collection}", func(r chi.Router) {
		r.Use(APIKey(conf))
		r.Get("/cards/{word}", ErrorHandler(handlers.Card))
		r.Post("/apkg", ErrorHandler(handlers.APKG))
	})

	r.Get("/", ErrorHandler(handlers.Index))
	r.Get("/styles.css", ErrorHandler(handlers.Styles))

	return r
}
