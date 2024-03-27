package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/la0rg/lingo/cardgen"
	"github.com/la0rg/lingo/client"
	"github.com/la0rg/lingo/config"
	"github.com/la0rg/lingo/handler"
	"github.com/la0rg/lingo/mock"
	"github.com/la0rg/lingo/repository"
	"github.com/la0rg/lingo/service"
)

func Run(ctx context.Context, conf *config.Config) error {
	gpt := initGPT()

	awsConfig, err := initAWSConfig(ctx, conf)
	if err != nil {
		return err
	}

	db := dynamodb.NewFromConfig(awsConfig)

	if err := repository.NewMigration(db).Run(ctx); err != nil {
		return err
	}

	generator, err := cardgen.NewOpenAI(gpt)
	if err != nil {
		return err
	}

	s3, err := client.NewS3(ctx, awsConfig)
	if err != nil {
		return err
	}

	router := handler.NewRouter(conf, generator, new(service.APKG), s3)

	if err := startHTTPServer(conf, router); err != nil {
		return err
	}

	slog.Info("Server stopped")
	return nil
}

func initGPT() cardgen.GPTClient {
	// return openai.NewClient(conf.OpenAIKey)
	return mock.GPTClient{}
}

func initAWSConfig(ctx context.Context, conf *config.Config) (aws.Config, error) {
	opts := []func(*awsconfig.LoadOptions) error{}

	if conf.Local {
		staticResolver := aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					PartitionID:       "aws",
					URL:               "http://localhost:8000",
					SigningRegion:     "us-east-1",
					HostnameImmutable: true,
				}, nil
			})

		credentialsProvider := credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "3f1lzp", SecretAccessKey: "m716ts", SessionToken: "",
				Source: "Mock credentials used above for local instance",
			},
		}

		opts = append(
			opts,
			awsconfig.WithRegion("us-east-1"),
			awsconfig.WithEndpointResolverWithOptions(staticResolver),
			awsconfig.WithCredentialsProvider(credentialsProvider),
		)
	}

	awsConfig, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return awsConfig, fmt.Errorf("unable to load AWS config: %v", err)
	}

	return awsConfig, nil
}

// startHTTPServer starts the HTTP server and blocks until it's stopped.
func startHTTPServer(conf *config.Config, handler http.Handler) error {
	addr := fmt.Sprintf(":%d", conf.Port)
	var srv = http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	idleConnsClosed := make(chan struct{})

	go func() {
		sig := <-interrupt
		slog.Info("Stopping HTTP server", "signal", sig.String())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("HTTP server shutdown failed", "error", err)
		}
		close(idleConnsClosed)
	}()

	slog.Info("Starting HTTP server", "addr", addr)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	<-idleConnsClosed
	return nil
}
