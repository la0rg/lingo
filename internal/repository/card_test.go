package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/internal/repository/mocks"
)

func fullCard() model.Card {
	return model.Card{
		ID:           1715808757001,
		Word:         "Apfel",
		CollectionID: 1715808757,
		Plural:       "Äpfel",
		Translations: model.Translations{
			{PartOfSpeech: "Nomen", Translations: []string{"apple"}},
			{PartOfSpeech: "Adjektiv", Translations: []string{"apple-like", "fruity"}},
		},
		Article:            "der",
		Gender:             "m.",
		Transcription:      "/ˈapfəl/",
		Context:            "Der Apfel ist rot.",
		ContextTranslation: "The apple is red.",
		Definition:         "Eine Frucht.",
	}
}

func TestCard_Put(t *testing.T) {
	t.Parallel()

	client := mocks.NewDynamoAPI(t)
	client.EXPECT().PutItem(mock.Anything, mock.Anything).
		RunAndReturn(func(
			_ context.Context, in *dynamodb.PutItemInput, _ ...func(*dynamodb.Options),
		) (*dynamodb.PutItemOutput, error) {
			assert.Equal(t, "lingo_cards", *in.TableName)
			assert.Equal(t, &types.AttributeValueMemberS{Value: "apfel"}, in.Item["key"], "key is the lowercased word")
			assert.Equal(t, &types.AttributeValueMemberN{Value: "1715808757"}, in.Item["collection_id"])
			assert.Equal(t, &types.AttributeValueMemberS{Value: "Apfel"}, in.Item["word"])
			return &dynamodb.PutItemOutput{}, nil
		})

	require.NoError(t, NewCard(client).Put(context.Background(), fullCard()))
}

func TestCard_Put_Error(t *testing.T) {
	t.Parallel()

	client := mocks.NewDynamoAPI(t)
	client.EXPECT().PutItem(mock.Anything, mock.Anything).Return(nil, errors.New("throttled"))

	assert.ErrorContains(t, NewCard(client).Put(context.Background(), fullCard()), "throttled")
}

func TestCard_Get(t *testing.T) {
	t.Parallel()

	client := mocks.NewDynamoAPI(t)
	client.EXPECT().GetItem(mock.Anything, mock.Anything).
		RunAndReturn(func(
			_ context.Context, in *dynamodb.GetItemInput, _ ...func(*dynamodb.Options),
		) (*dynamodb.GetItemOutput, error) {
			assert.Equal(t, "lingo_cards", *in.TableName)
			assert.Equal(t, map[string]types.AttributeValue{
				"key":           &types.AttributeValueMemberS{Value: "apfel"},
				"collection_id": &types.AttributeValueMemberN{Value: "42"},
			}, in.Key)
			return &dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{
				"word":          &types.AttributeValueMemberS{Value: "Apfel"},
				"collection_id": &types.AttributeValueMemberN{Value: "42"},
			}}, nil
		})

	card, err := NewCard(client).Get(context.Background(), 42, "APFEL")

	require.NoError(t, err)
	assert.Equal(t, &model.Card{Word: "Apfel", CollectionID: 42}, card)
}

func TestCard_Get_NotFound(t *testing.T) {
	t.Parallel()

	client := mocks.NewDynamoAPI(t)
	client.EXPECT().GetItem(mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{}, nil)

	card, err := NewCard(client).Get(context.Background(), 42, "qwrt")

	require.NoError(t, err)
	assert.Nil(t, card)
}

func TestCard_Get_Errors(t *testing.T) {
	t.Parallel()

	t.Run("request", func(t *testing.T) {
		t.Parallel()
		client := mocks.NewDynamoAPI(t)
		client.EXPECT().GetItem(mock.Anything, mock.Anything).Return(nil, errors.New("throttled"))

		_, err := NewCard(client).Get(context.Background(), 42, "apfel")

		assert.ErrorContains(t, err, "throttled")
	})

	t.Run("unmarshal", func(t *testing.T) {
		t.Parallel()
		client := mocks.NewDynamoAPI(t)
		client.EXPECT().GetItem(mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "not a number"}},
		}, nil)

		_, err := NewCard(client).Get(context.Background(), 42, "apfel")

		assert.ErrorContains(t, err, "unmarshal")
	})
}
