package handler

import (
	"context"
	"net/http"
	"slices"

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
	limiter    *wordLimiter
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
		limiter:    newWordLimiter(conf.Limits.WordsPerMinute),
		card:       card,
		apkg:       apkg,
		tmpl:       tmpl,
		collection: collection,
	}
	apiKey := APIKey(conf)
	html := handlers.ErrorHandlerHTML

	mux := http.NewServeMux()
	mux.Handle("GET /version", ErrorHandler(handlers.Version))

	mux.Handle("GET /api/collections/{collection}/cards/{word}", apiKey(ErrorHandler(handlers.Card)))
	mux.Handle("POST /api/collections/{collection}/apkg", apiKey(ErrorHandler(handlers.APKG)))

	mux.Handle("GET /card", html(handlers.CardPreview))
	mux.Handle("POST /language", html(handlers.Language))
	mux.Handle("GET /collection", html(handlers.Collection))
	mux.Handle("POST /collection", html(handlers.CreateCollection))
	mux.Handle("GET /styles.css", html(handlers.Styles))
	mux.Handle("GET /public/", http.FileServerFS(web.PublicFS))
	mux.Handle("GET /{$}", html(handlers.Index))

	return chain(mux,
		RequestLogger, RequestID, Recoverer, SecurityHeaders, MaxBodySize(maxBodyBytes), clientIP(conf.Limits),
	)
}

// chain wraps h with middlewares, the first one being the outermost.
func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range slices.Backward(middlewares) {
		h = middleware(h)
	}
	return h
}
