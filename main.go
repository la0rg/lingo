package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/la0rg/lingo/cardgen"
	"github.com/la0rg/lingo/config"
	"github.com/la0rg/lingo/mock"
)

func main() {
	conf, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	initLogger(conf)

	//gpt := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	gpt := mock.GPTClient{}

	generator, err := cardgen.NewOpenAI(gpt)
	if err != nil {
		log.Fatal(err)
	}

	result, err := generator.Generate(
		context.Background(), "de", "en",
		"Kuchen",
		"käse",
		"verrückt",
	)
	if err != nil {
		log.Fatal(err)
	}

	slog.Debug("Testing debug log", "port", conf.Port)
	slog.Info("Card generated", "result", result)
}

func initLogger(conf *config.Config) {
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(conf.LogLevel)); err != nil {
		slog.Warn("Invalid log level in config", "log_level", conf.LogLevel)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)
}
