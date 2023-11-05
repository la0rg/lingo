package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/la0rg/lingo/cardgen"
	"github.com/la0rg/lingo/mock"
)

func main() {
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
			"Kaffee",
			"käse",
			"verrückt",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Card generated", "result", result)
}
