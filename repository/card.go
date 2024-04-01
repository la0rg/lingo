package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/la0rg/lingo/model"
)

type Card struct {
	client *dynamodb.Client
}

func NewCard(client *dynamodb.Client) *Card {
	return &Card{client: client}
}

func (c *Card) Put(ctx context.Context, card model.Card) error {
	item, err := attributevalue.MarshalMap(card)
	if err != nil {
		return fmt.Errorf("unable to marshal card: %v", err)
	}

	_, err = c.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: cardsTableSchema.TableName,
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("unable to save a card: %v", err)
	}

	return nil
}
