package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/la0rg/lingo/internal/repository/mocks"
)

func TestMigration_TableExists(t *testing.T) {
	t.Parallel()

	client := mocks.NewDynamoAPI(t) // CreateTable must not be called
	client.EXPECT().DescribeTable(mock.Anything, mock.Anything).Return(&dynamodb.DescribeTableOutput{}, nil)

	assert.NoError(t, NewMigration(client).Run(context.Background()))
}

func TestMigration_Errors(t *testing.T) {
	t.Parallel()

	t.Run("describe", func(t *testing.T) {
		t.Parallel()
		client := mocks.NewDynamoAPI(t)
		client.EXPECT().DescribeTable(mock.Anything, mock.Anything).Return(nil, errors.New("no access"))

		assert.ErrorContains(t, NewMigration(client).Run(context.Background()), "no access")
	})

	t.Run("create", func(t *testing.T) {
		t.Parallel()
		client := mocks.NewDynamoAPI(t)
		client.EXPECT().DescribeTable(mock.Anything, mock.Anything).
			Return(nil, &types.ResourceNotFoundException{})
		client.EXPECT().CreateTable(mock.Anything, cardsTableSchema).Return(nil, errors.New("limit exceeded"))

		assert.ErrorContains(t, NewMigration(client).Run(context.Background()), "limit exceeded")
	})
}
