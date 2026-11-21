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
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/la0rg/lingo/internal/cardgen"
	"github.com/la0rg/lingo/internal/client"
	"github.com/la0rg/lingo/internal/config"
	"github.com/la0rg/lingo/internal/handler"
	"github.com/la0rg/lingo/internal/mock"
	"github.com/la0rg/lingo/internal/repository"
	"github.com/la0rg/lingo/internal/service"
	"github.com/la0rg/lingo/web"
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

	generator, err := cardgen.NewOpenAI(gpt, generatorOpts(conf)...)
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

	templates, err := web.NewTemplates()
	if err != nil {
		return err
	}

	collectionRepo, err := repository.NewCollection()
	if err != nil {
		return err
	}

	cardRepo := repository.NewCard(db)
	apkgRepo := repository.NewAPKG(s3, conf)
	cardService := service.NewCard(generator, cardRepo)
	apkgService := service.NewAPKG(cardService, apkgRepo)

	router := handler.NewRouter(conf, cardService, apkgService, templates, collectionRepo)

	if err := startHTTPServer(conf, router); err != nil {
		return err
	}

	slog.Info("Server stopped")
	return nil
}

func gpt(conf *config.Config) cardgen.GPTClient {
	if conf.MockOpenAI {
		return mock.GPTClient{}
	}

	// the client retries connection errors, 408, 409, 429 and 5xx responses with backoff
	client := openai.NewClient(option.WithAPIKey(conf.OpenAIKey), option.WithMaxRetries(2))
	return &client.Chat.Completions
}

func generatorOpts(conf *config.Config) []cardgen.Option {
	var opts []cardgen.Option
	if conf.OpenAIModel != "" {
		opts = append(opts, cardgen.WithModel(conf.OpenAIModel))
	}
	if conf.OpenAIReasoningEffort != "" {
		opts = append(opts, cardgen.WithReasoningEffort(conf.OpenAIReasoningEffort))
	}
	return opts
}

func dbConfig(ctx context.Context, conf *config.Config) (aws.Config, error) {
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx, awsOpts(conf.Dynamo)...)
	if err != nil {
		return awsConfig, fmt.Errorf("unable to load AWS config: %v", err)
	}

	return awsConfig, nil
}

func s3Config(ctx context.Context, conf *config.Config) (aws.Config, error) {
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx, awsOpts(conf.S3.AWS)...)
	if err != nil {
		return awsConfig, fmt.Errorf("unable to load AWS config: %v", err)
	}

	return awsConfig, nil
}

func awsOpts(conf config.AWS) []func(*awsconfig.LoadOptions) error {
	var options []func(*awsconfig.LoadOptions) error

	if conf.Endpoint != "" {
		options = append(options, awsconfig.WithBaseEndpoint(conf.Endpoint))
	}

	if conf.Key != "" && conf.Secret != "" {
		credentialsProvider := credentials.StaticCredentialsProvider{
			Value: aws.Credentials{AccessKeyID: conf.Key, SecretAccessKey: conf.Secret},
		}

		options = append(options, awsconfig.WithCredentialsProvider(credentialsProvider))
	}

	if conf.Region != "" {
		options = append(options, awsconfig.WithRegion(conf.Region))
	}

	return options
}

// startHTTPServer starts the HTTP server and blocks until it's stopped.
func startHTTPServer(conf *config.Config, handler http.Handler) error {
	addr := fmt.Sprintf(":%d", conf.Port)
	var srv = http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: time.Minute,
		IdleTimeout:  time.Minute * 5,
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
