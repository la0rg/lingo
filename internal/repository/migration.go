package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var cardsTableSchema = &dynamodb.CreateTableInput{
	AttributeDefinitions: []types.AttributeDefinition{{
		AttributeName: aws.String("key"),
		AttributeType: types.ScalarAttributeTypeS,
	}, {
		AttributeName: aws.String("collection_id"),
		AttributeType: types.ScalarAttributeTypeN,
	}},
	KeySchema: []types.KeySchemaElement{{
		AttributeName: aws.String("key"),
		KeyType:       types.KeyTypeHash,
	}, {
		AttributeName: aws.String("collection_id"),
		KeyType:       types.KeyTypeRange,
	}},
	TableName: aws.String("lingo_cards"),
	ProvisionedThroughput: &types.ProvisionedThroughput{
		ReadCapacityUnits:  aws.Int64(25),
		WriteCapacityUnits: aws.Int64(25),
	},
}

type Migration struct {
	client DynamoAPI
}

func NewMigration(client DynamoAPI) Migration {
	return Migration{client: client}
}

func (m Migration) Run(ctx context.Context) error {
	if err := m.createIfNotExists(ctx, cardsTableSchema); err != nil {
		return err
	}

	slog.Info("Migration completed")
	return nil
}

func (m Migration) createIfNotExists(ctx context.Context, input *dynamodb.CreateTableInput) error {
	exists, err := m.tableExists(ctx, *input.TableName)
	if err != nil {
		return err
	}

	if exists {
		return nil // TODO update
	}

	return m.createTable(ctx, input)
}

func (m Migration) tableExists(ctx context.Context, tableName string) (bool, error) {
	_, err := m.client.DescribeTable(
		ctx, &dynamodb.DescribeTableInput{TableName: aws.String(tableName)},
	)
	if err != nil {
		var notFoundEx *types.ResourceNotFoundException
		if errors.As(err, &notFoundEx) {
			return false, nil
		}

		return false, fmt.Errorf("unable to describe a table: %v", err)
	}

	return true, nil
}

func (m Migration) createTable(ctx context.Context, input *dynamodb.CreateTableInput) error {
	slog.Info("Creating table", "table", aws.ToString(input.TableName))

	_, err := m.client.CreateTable(ctx, input)
	if err != nil {
		return fmt.Errorf("unnable to create a table: %v", err)
	}

	waiter := dynamodb.NewTableExistsWaiter(m.client)
	err = waiter.Wait(ctx, &dynamodb.DescribeTableInput{TableName: input.TableName}, time.Minute)
	if err != nil {
		return fmt.Errorf("unable to create a table: %v", err)
	}

	return nil
}
