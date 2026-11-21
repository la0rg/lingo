package repository

import (
	"context"
	"errors"
	"io"
	"testing"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/la0rg/lingo/internal/config"
	"github.com/la0rg/lingo/internal/repository/mocks"
)

var s3Conf = &config.Config{S3: config.S3{Bucket: "lingo"}}

func TestAPKG_StoreAndGetURL(t *testing.T) {
	t.Parallel()

	client := mocks.NewS3Client(t)
	client.EXPECT().PutObject(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, in *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			assert.Equal(t, "lingo", *in.Bucket)
			assert.Equal(t, "apkg/1/file.apkg", *in.Key)
			body, err := io.ReadAll(in.Body)
			require.NoError(t, err)
			assert.Equal(t, []byte("archive"), body)
			return &s3.PutObjectOutput{}, nil
		})
	client.EXPECT().PresignGetObject(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(
			_ context.Context, in *s3.GetObjectInput, optFns ...func(*s3.PresignOptions),
		) (*v4.PresignedHTTPRequest, error) {
			assert.Equal(t, "lingo", *in.Bucket)
			assert.Equal(t, "apkg/1/file.apkg", *in.Key)

			var opts s3.PresignOptions
			for _, fn := range optFns {
				fn(&opts)
			}
			assert.Equal(t, expiration, opts.Expires)

			return &v4.PresignedHTTPRequest{URL: "https://s3.example/presigned"}, nil
		})

	url, err := NewAPKG(client, s3Conf).StoreAndGetURL(context.Background(), "apkg/1/file.apkg", []byte("archive"))

	require.NoError(t, err)
	assert.Equal(t, "https://s3.example/presigned", url)
}

func TestAPKG_StoreAndGetURL_Errors(t *testing.T) {
	t.Parallel()

	t.Run("put", func(t *testing.T) {
		t.Parallel()
		client := mocks.NewS3Client(t)
		client.EXPECT().PutObject(mock.Anything, mock.Anything).Return(nil, errors.New("s3 down"))

		_, err := NewAPKG(client, s3Conf).StoreAndGetURL(context.Background(), "k", nil)

		assert.ErrorContains(t, err, "s3 down")
	})

	t.Run("presign", func(t *testing.T) {
		t.Parallel()
		client := mocks.NewS3Client(t)
		client.EXPECT().PutObject(mock.Anything, mock.Anything).Return(&s3.PutObjectOutput{}, nil)
		client.EXPECT().PresignGetObject(mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("no credentials"))

		_, err := NewAPKG(client, s3Conf).StoreAndGetURL(context.Background(), "k", nil)

		assert.ErrorContains(t, err, "no credentials")
	})
}
