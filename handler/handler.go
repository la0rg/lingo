package handler

import (
	"context"
	"net/http"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/la0rg/lingo/apkg"
	"github.com/la0rg/lingo/config"
	"github.com/la0rg/lingo/model"
)

type CardsGenerator interface {
	Generate(ctx context.Context, sourceLanguage, targetLanguage string, words ...string) (*model.Cards, error)
}

type APKGService interface {
	Create(ctx context.Context, cards *model.Cards) (*apkg.Collection, error)
}

type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

type handlers struct {
	conf  *config.Config
	cards CardsGenerator
	apkg  APKGService
	s3    S3Client
}

func NewRouter(conf *config.Config, cards CardsGenerator, apkg APKGService, s3 S3Client) http.Handler {
	handlers := handlers{conf: conf, cards: cards, apkg: apkg, s3: s3}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(RequestLogger)
	r.Use(Recoverer)
	r.Use(APIKey(conf))

	r.Route("/collections/{collection}", func(r chi.Router) {
		r.Get("/cards/{card}", ErrorHandler(handlers.Card))
		r.Post("/apkg", ErrorHandler(handlers.APKG))
	})

	return r
}
