package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/la0rg/lingo/cardgen"
	"github.com/la0rg/lingo/config"
	"github.com/la0rg/lingo/handler"
	"github.com/la0rg/lingo/mock"
)

func main() {
	conf, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	initLogger(conf)

	// gpt := openai.NewClient(conf.OpenAIKey)
	gpt := mock.GPTClient{}

	generator, err := cardgen.NewOpenAI(gpt)
	if err != nil {
		log.Fatal(err)
	}

	router := handler.NewRouter(generator)

	if err := startHTTPServer(conf, router); err != nil {
		log.Fatal("HTTP server failed", err)
	}

	slog.Info("Server stopped")
}

func initLogger(conf *config.Config) {
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(conf.LogLevel)); err != nil {
		slog.Warn("Invalid log level in config", "log_level", conf.LogLevel)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)
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
