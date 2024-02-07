package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/la0rg/lingo/apkg"
	"github.com/la0rg/lingo/logger"
	"github.com/la0rg/lingo/model"
)

func (h *handlers) APKG(w http.ResponseWriter, r *http.Request) error {
	collectionName := strings.ToLower(chi.URLParam(r, "collection"))

	collection, ok := collections[collectionName]
	if !ok {
		return ErrCollectionNotFound
	}

	var request model.APKGRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return BadRequest("invalid request format")
	}

	logger := logger.WithContext(r.Context()).With("collection", collectionName, "request", request)
	logger.Info("Creating an APKG file")

	cards, err := h.cards.Generate(r.Context(), collection.sourceLanguage, collection.targetLanguage, request.Words...)
	if err != nil {
		return fmt.Errorf("unable to generate cards: %w", err)
	}

	apkgCollectionn, err := h.apkg.Create(r.Context(), cards)
	if err != nil {
		return fmt.Errorf("unable to create APKG file: %w", err)
	}

	archive, err := apkg.Build(apkgCollectionn)
	if err != nil {
		return fmt.Errorf("unable to build APKG file: %w", err)
	}

	s3Key := fmt.Sprintf("apkg/%s/lingo_%s.apkg", uuid.New().String(), collectionName)
	url, err := h.saveToS3(r.Context(), s3Key, archive)
	if err != nil {
		return fmt.Errorf("unable to save APKG file to S3: %w", err)
	}

	JSON(w, r, model.APKGResponse{URL: url})
	return nil
}

func (h *handlers) saveToS3(ctx context.Context, key string, archive []byte) (string, error) {
	// Load the Shared AWS Configuration (~/.aws/config)
	opts := []func(*config.LoadOptions) error{}

	if h.conf.Local {
		staticResolver := aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					PartitionID:       "aws",
					URL:               "http://localhost:9000", // or where ever you ran minio
					SigningRegion:     "us-east-1",
					HostnameImmutable: true,
				}, nil
			})
		opts = append(opts, config.WithEndpointResolverWithOptions(staticResolver))
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		return "", err
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("lingo"),
		Key:    aws.String(key),
		Body:   bytes.NewReader(archive),
	})
	if err != nil {
		return "", fmt.Errorf("unable to save file to S3: %v", err)
	}

	presignedResponse, err := s3.NewPresignClient(client).PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String("lingo"),
			Key:    aws.String(key),
		},
		func(opts *s3.PresignOptions) {
			opts.Expires = 60 * time.Second
		},
	)
	if err != nil {
		return "", fmt.Errorf("unable to generate presigned URL: %v", err)
	}

	return presignedResponse.URL, nil
}
