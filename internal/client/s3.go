package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	*s3.Client
	*s3.PresignClient
}

func NewS3(ctx context.Context, awsConfig aws.Config) (*S3, error) {
	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		// Custom endpoints (MinIO, local dev) don't support virtual-hosted-style bucket addressing.
		o.UsePathStyle = awsConfig.BaseEndpoint != nil
	})
	presignedClient := s3.NewPresignClient(client)

	return &S3{
		Client:        client,
		PresignClient: presignedClient,
	}, nil
}
