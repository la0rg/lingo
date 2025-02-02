package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/la0rg/lingo/apkg"
	"github.com/la0rg/lingo/internal/logger"
	"github.com/la0rg/lingo/internal/model"
)

var apkgConfig = apkg.DefaultConfig(6493485093845, "lingo_config")

var ErrEmptyCards = fmt.Errorf("empty cards")

type APKG struct {
	card     CardGetter
	apkgRepo APKGRepository
}

func NewAPKG(card CardGetter, apkgRepo APKGRepository) *APKG {
	return &APKG{card: card, apkgRepo: apkgRepo}
}

type CardGetter interface {
	GetAll(ctx context.Context, collection *model.Collection, words ...string) (*GetAllResult, error)
}

func (s *APKG) CreateOnS3(
	ctx context.Context, collection *model.Collection, words ...string,
) (*model.CollectionFile, error) {
	logger := logger.WithContext(ctx).With("collection", collection.ID, "words", words)
	logger.Info("Creating an APKG file")

	result, err := s.card.GetAll(ctx, collection, words...)
	if err != nil {
		return nil, fmt.Errorf("unable to get cards: %w", err)
	}

	apkgCollection, err := s.buildAPKG(collection, result.Cards)
	if err != nil {
		return nil, fmt.Errorf("unable to create APKG collection: %w", err)
	}

	archive, err := apkg.Build(apkgCollection)
	if err != nil {
		return nil, fmt.Errorf("unable to build APKG file: %w", err)
	}

	url, err := s.apkgRepo.StoreAndGetURL(ctx, s.s3Key(collection), archive)
	if err != nil {
		return nil, fmt.Errorf("unable to save APKG file to S3: %w", err)
	}

	return &model.CollectionFile{
		URL:            url,
		CardsCount:     len(result.Cards),
		CollectionName: collection.Name,
		MissingWords:   result.MissingWords,
	}, nil
}

func (s *APKG) buildAPKG(collection *model.Collection, cards []model.Card) (*apkg.Collection, error) {
	if len(cards) == 0 {
		return nil, ErrEmptyCards
	}

	notes := make([]apkg.Note, 0)

	for _, card := range cards {
		note := apkg.Note{ID: card.ID, Model: collection.Template}
		for _, field := range collection.Template.Fields {
			note.Fields = append(note.Fields, card.FieldByName(field))
		}
		notes = append(notes, note)
	}

	return &apkg.Collection{
		Decks: []apkg.Deck{{
			ID:     collection.ID,
			Name:   collection.Name,
			Config: apkgConfig,
			Notes:  notes,
		}},
	}, nil
}

func (s *APKG) s3Key(collection *model.Collection) string {
	return fmt.Sprintf(
		"apkg/%d/lingo_%d_%s.apkg",
		collection.ID, time.Now().Unix(), uuid.New().String(),
	)
}
