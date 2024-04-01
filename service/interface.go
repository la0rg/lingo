package service

import (
	"context"

	"github.com/la0rg/lingo/model"
)

type CardGenerator interface {
	Generate(ctx context.Context, collection *model.Collection, words ...string) (*model.Cards, error)
}

type CardRepository interface {
	Put(ctx context.Context, card model.Card) error
}

type APKGRepository interface {
	StoreAndGetURL(ctx context.Context, key string, archive []byte) (string, error)
}
