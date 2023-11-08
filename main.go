package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/la0rg/lingo/cardgen"
	"github.com/la0rg/lingo/mock"
)

func main() {
	var logLevel slog.Level
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		if err := logLevel.UnmarshalText([]byte(level)); err != nil {
			log.Fatal(err)
		}
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	//gpt := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	gpt := mock.GPTClient{}

	generator, err := cardgen.NewOpenAI(gpt)
	if err != nil {
		log.Fatal(err)
	}

	result, err := generator.Generate(context.Background(), cardgen.Request{
		SourceLanguage: "de",
		TargetLanguage: "en",
		Words: []string{
			"Kuchen",
			"käse",
			"verrückt",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Card generated", "result", result)
}
