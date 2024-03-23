package client

import (
	"context"
	"fmt"

	lconfig "github.com/la0rg/lingo/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	*s3.Client
	*s3.PresignClient
}

func NewS3(ctx context.Context, conf *lconfig.Config) (*S3, error) {
	// Load the Shared AWS Configuration (~/.aws/config)
	opts := []func(*config.LoadOptions) error{}

	if conf.Local {
		staticResolver := aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					PartitionID:       "aws",
					URL:               "http://127.0.0.1:9000",
					SigningRegion:     "us-east-1",
					HostnameImmutable: true,
				}, nil
			})
		opts = append(opts, config.WithEndpointResolverWithOptions(staticResolver))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %v", err)
	}

	client := s3.NewFromConfig(cfg)
	presignedClient := s3.NewPresignClient(client)

	return &S3{
		Client:        client,
		PresignClient: presignedClient,
	}, nil
}
