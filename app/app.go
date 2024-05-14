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
	"github.com/sashabaranov/go-openai"

	"github.com/la0rg/lingo/cardgen"
	"github.com/la0rg/lingo/client"
	"github.com/la0rg/lingo/config"
	"github.com/la0rg/lingo/handler"
	"github.com/la0rg/lingo/repository"
	"github.com/la0rg/lingo/service"
)

func Run(ctx context.Context, conf *config.Config) error {
	gpt := gpt(conf)

	dbConfig, err := dbConfig(ctx, conf)
	if err != nil {
		return err
	}
	db := dynamodb.NewFromConfig(dbConfig)

	if err := repository.NewMigration(db).Run(ctx); err != nil {
		return err
	}

	generator, err := cardgen.NewOpenAI(gpt)
	if err != nil {
		return err
	}

	s3Config, err := s3Config(ctx, conf)
	if err != nil {
		return err
	}

	s3, err := client.NewS3(ctx, s3Config)
	if err != nil {
		return err
	}

	cardRepo := repository.NewCard(db)
	apkgRepo := repository.NewAPKG(s3)
	cardService := service.NewCard(generator, cardRepo)
	apkgService := service.NewAPKG(generator, apkgRepo)

	router := handler.NewRouter(conf, cardService, apkgService)

	if err := startHTTPServer(conf, router); err != nil {
		return err
	}

	slog.Info("Server stopped")
	return nil
}

func gpt(conf *config.Config) cardgen.GPTClient {
	return openai.NewClient(conf.OpenAIKey)
	// return mock.GPTClient{}
}

func dbConfig(ctx context.Context, conf *config.Config) (aws.Config, error) {
	var opts []func(*awsconfig.LoadOptions) error

	if conf.Local {
		opts = localAWSOpts("http://localhost:8000", "us-east-1", "dynamodb", "dynamodb")
	}

	awsConfig, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return awsConfig, fmt.Errorf("unable to load AWS config: %v", err)
	}

	return awsConfig, nil
}

func s3Config(ctx context.Context, conf *config.Config) (aws.Config, error) {
	var opts []func(*awsconfig.LoadOptions) error

	if conf.Local {
		opts = localAWSOpts("http://localhost:9000", "us-east-1", "minioadmin", "minioadmin")
	}

	awsConfig, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return awsConfig, fmt.Errorf("unable to load AWS config: %v", err)
	}

	return awsConfig, nil
}

func localAWSOpts(url, region, key, secret string) []func(*awsconfig.LoadOptions) error {
	staticResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:       "aws",
				URL:               url,
				SigningRegion:     region,
				HostnameImmutable: true,
			}, nil
		})

	credentialsProvider := credentials.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID: key, SecretAccessKey: secret, SessionToken: "",
			Source: "Mock credentials used above for local instance",
		},
	}

	return []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
		awsconfig.WithEndpointResolverWithOptions(staticResolver),
		awsconfig.WithCredentialsProvider(credentialsProvider),
	}
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
