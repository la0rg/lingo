package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/la0rg/lingo/internal/model"
)

// DynamoAPI is the subset of the DynamoDB client used by the repositories.
type DynamoAPI interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (
		*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (
		*dynamodb.PutItemOutput, error)
	DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (
		*dynamodb.DescribeTableOutput, error)
	CreateTable(ctx context.Context, params *dynamodb.CreateTableInput, optFns ...func(*dynamodb.Options)) (
		*dynamodb.CreateTableOutput, error)
}

type Card struct {
	partitionKey string
	sortKey      string

	client DynamoAPI
}

func NewCard(client DynamoAPI) *Card {
	return &Card{
		partitionKey: "key",
		sortKey:      "collection_id",
		client:       client,
	}
}

func (r *Card) Put(ctx context.Context, card model.Card) error {
	item, err := attributevalue.MarshalMap(card)
	if err != nil {
		return fmt.Errorf("unable to marshal card: %v", err)
	}

	item[r.partitionKey] = &types.AttributeValueMemberS{Value: r.buildPartitionKey(card.Word)}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: cardsTableSchema.TableName,
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("unable to save a card: %v", err)
	}

	return nil
}

func (r *Card) Get(ctx context.Context, collection int64, word string) (*model.Card, error) {
	key := map[string]types.AttributeValue{
		r.partitionKey: &types.AttributeValueMemberS{Value: r.buildPartitionKey(word)},
		r.sortKey:      &types.AttributeValueMemberN{Value: strconv.FormatInt(collection, 10)},
	}

	response, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		Key: key, TableName: cardsTableSchema.TableName,
	})
	if err != nil {
		return nil, fmt.Errorf("unnable to fetch a card from db: %v", err)
	}

	if len(response.Item) == 0 {
		return nil, nil
	}

	var card model.Card
	if err := attributevalue.UnmarshalMap(response.Item, &card); err != nil {
		return nil, fmt.Errorf("unable to unmarshal card from a db: %v", err)
	}

	return &card, nil
}

func (r *Card) buildPartitionKey(word string) string {
	return strings.ToLower(word)
}
