package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/la0rg/lingo/internal/config"
	"github.com/la0rg/lingo/internal/model"
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
	publicHandler := http.FileServer(http.FS(web.PublicFS))

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(RequestLogger)
	r.Use(Recoverer)

	r.Route("/api/collections/{collection}", func(r chi.Router) {
		r.Use(APIKey(conf))
		r.Get("/cards/{word}", ErrorHandler(handlers.Card))
		r.Post("/apkg", ErrorHandler(handlers.APKG))
	})

	r.Get("/card", handlers.ErrorHandlerHTML(handlers.CardPreview))
	r.Post("/language", handlers.ErrorHandlerHTML(handlers.Language))
	r.Get("/styles.css", handlers.ErrorHandlerHTML(handlers.Styles))
	r.Get("/public/*", publicHandler.ServeHTTP)
	r.Get("/", handlers.ErrorHandlerHTML(handlers.Index))

	return r
}
