package repository

import (
	"bytes"
	"context"
	"fmt"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
)

type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

type APKG struct {
	s3 S3Client
}

func NewAPKG(s3 S3Client) *APKG {
	return &APKG{s3: s3}
}

func (r *APKG) StoreAndGetURL(ctx context.Context, key string, archive []byte) (string, error) {
	_, err := r.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("lingo"),
		Key:    aws.String(key),
		Body:   bytes.NewReader(archive),
	})
	if err != nil {
		return "", fmt.Errorf("unable to save file to S3: %v", err)
	}

	presignedResponse, err := r.s3.PresignGetObject(
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
