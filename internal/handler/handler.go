package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/la0rg/lingo/internal/config"
	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/web"
)

type CardService interface {
	Get(ctx context.Context, collection *model.Collection, word string) (*model.Card, error)
}

type APKGService interface {
	CreateOnS3(ctx context.Context, collection *model.Collection, words ...string) (*model.CollectionFile, error)
}

type TemplateRenderer interface {
	Render(name web.TemplateName, data any) ([]byte, error)
}

type CollectionGetter interface {
	GetBySlug(slug string) *model.Collection
}

type handlers struct {
	conf       *config.Config
	limiter    *WordLimiter
	card       CardService
	apkg       APKGService
	tmpl       TemplateRenderer
	collection CollectionGetter
}

func NewRouter(
	conf *config.Config,
	card CardService,
	apkg APKGService,
	tmpl TemplateRenderer,
	collection CollectionGetter,
) http.Handler {
	handlers := handlers{
		conf:       conf,
		limiter:    NewWordLimiter(conf.Limits.WordsPerMinute, conf.Limits.WordsBurst, conf.Limits.ClientIPHeader),
		card:       card,
		apkg:       apkg,
		tmpl:       tmpl,
		collection: collection,
	}
	publicHandler := http.FileServer(http.FS(web.PublicFS))

	r := chi.NewRouter()
	r.Use(RequestLogger, RequestID, Recoverer, SecurityHeaders, MaxBodySize(maxBodyBytes))
	r.Get("/version", ErrorHandler(handlers.Version))

	r.Route("/api/collections/{collection}", func(r chi.Router) {
		r.Use(APIKey(conf))
		r.Get("/cards/{word}", ErrorHandler(handlers.Card))
		r.Post("/apkg", ErrorHandler(handlers.APKG))
	})

	r.Get("/card", handlers.ErrorHandlerHTML(handlers.CardPreview))
	r.Post("/language", handlers.ErrorHandlerHTML(handlers.Language))
	r.Get("/collection", handlers.ErrorHandlerHTML(handlers.Collection))
	r.Post("/collection", handlers.ErrorHandlerHTML(handlers.CreateCollection))
	r.Get("/styles.css", handlers.ErrorHandlerHTML(handlers.Styles))
	r.Get("/public/*", publicHandler.ServeHTTP)
	r.Get("/", handlers.ErrorHandlerHTML(handlers.Index))

	return r
}
