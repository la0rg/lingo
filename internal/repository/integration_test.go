//go:build integration

package repository

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newLocalDynamo connects to DynamoDB Local (`make up`), or DYNAMO_ENDPOINT when set.
func newLocalDynamo(t *testing.T) *dynamodb.Client {
	t.Helper()

	endpoint := os.Getenv("DYNAMO_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8000"
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithBaseEndpoint(endpoint),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dynamodb", "dynamodb", "")),
	)
	require.NoError(t, err)

	client := dynamodb.NewFromConfig(cfg)
	_, err = client.ListTables(context.Background(), &dynamodb.ListTablesInput{Limit: aws.Int32(1)})
	require.NoError(t, err, "DynamoDB Local is not reachable at %s, run `make up`", endpoint)

	return client
}

func TestIntegration_MigrationIsIdempotent(t *testing.T) {
	client := newLocalDynamo(t)

	require.NoError(t, NewMigration(client).Run(context.Background()))
	require.NoError(t, NewMigration(client).Run(context.Background()))
}

func TestIntegration_CardRoundTrip(t *testing.T) {
	client := newLocalDynamo(t)
	ctx := context.Background()
	require.NoError(t, NewMigration(client).Run(ctx))
	repo := NewCard(client)

	card := fullCard()
	card.Word = "IntegrationApfel"
	require.NoError(t, repo.Put(ctx, card))

	got, err := repo.Get(ctx, card.CollectionID, "integrationapfel")
	require.NoError(t, err)
	assert.Equal(t, &card, got, "every column is read back")

	other, err := repo.Get(ctx, card.CollectionID+1, "integrationapfel")
	require.NoError(t, err)
	assert.Nil(t, other, "cards are scoped to a collection")

	missing, err := repo.Get(ctx, card.CollectionID, "does-not-exist")
	require.NoError(t, err)
	assert.Nil(t, missing)
}

func TestIntegration_CardOptionalFieldsOmitted(t *testing.T) {
	client := newLocalDynamo(t)
	ctx := context.Background()
	require.NoError(t, NewMigration(client).Run(ctx))
	repo := NewCard(client)

	card := fullCard()
	card.Word, card.Article, card.Gender, card.Transcription = "IntegrationMinimal", "", "", ""
	card.Context, card.ContextTranslation, card.Definition = "", "", ""
	require.NoError(t, repo.Put(ctx, card))

	got, err := repo.Get(ctx, card.CollectionID, "integrationminimal")
	require.NoError(t, err)
	assert.Equal(t, &card, got)
}
