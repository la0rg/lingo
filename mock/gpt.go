package mock

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type GPTClient struct{}

func (GPTClient) CreateChatCompletion(
	ctx context.Context,
	request openai.ChatCompletionRequest,
) (response openai.ChatCompletionResponse, err error) {
	return openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{Message: openai.ChatCompletionMessage{Content: mockResponse}},
		},
	}, nil
}

const mockResponse = `
{
  "cards": [
    {
      "word": "Kuchen",
      "translations": [
        {
          "pos": "Noun",
          "translations": [
            "cake"
          ]
        }
      ],
      "article": "der",
      "transcription": "/ˈkuːχən/",
      "context": "Ich esse gerne Kuchen.",
      "context_translation": "I like to eat cake.",
      "definition": "Ein süßes Gebäck, das gebacken wird und meistens mit Creme oder Glasur versehen ist."
    },
    {
      "word": "Kaffee",
      "translations": [
        {
          "pos": "Noun",
          "translations": [
            "coffee"
          ]
        }
      ],
      "article": "der",
      "transcription": "/ˈkafeː/",
      "context": "Ich trinke jeden Morgen Kaffee.",
      "context_translation": "I drink coffee every morning.",
      "definition": "Ein heißes Getränk, das aus gerösteten Kaffeebohnen hergestellt wird."
    },
    {
      "word": "Käse",
      "translations": [
        {
          "pos": "Noun",
          "translations": [
            "cheese"
          ]
        }
      ],
      "article": "der",
      "transcription": "/ˈkɛːzə/",
      "context": "Ich mag gerne Käse auf meinem Brot.",
      "context_translation": "I like to have cheese on my bread.",
      "definition": "Ein Milchprodukt, das aus gemolkener Milch hergestellt wird und einen charakteristischen Geschmack hat."
    },
    {
      "word": "verrückt",
      "translations": [
        {
          "pos": "Adjective",
          "translations": [
            "crazy",
            "insane",
            "mad"
          ]
        }
      ],
      "transcription": "/fɛrˈrʏkt/",
      "context": "Er ist verrückt.",
      "context_translation": "He is crazy.",
      "definition": "Unvernünftig oder unkontrolliertes Verhalten, das von der Norm abweicht."
    }
  ]
}`
