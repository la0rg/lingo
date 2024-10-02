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
			{
				Message:      openai.ChatCompletionMessage{Content: mockResponse},
				FinishReason: openai.FinishReasonStop,
			},
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
      "transcription": "/ˈkuːxən/",
      "context": "Ich backe einen Kuchen.",
      "context_translation": "I am baking a cake.",
      "definition": "Ein süßes Gericht, das gebacken wird und typischerweise Mehl, Zucker, Eier und Butter enthält."
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
      "context": "Ein Stück Käse bitte.",
      "context_translation": "A piece of cheese, please.",
      "definition": "Ein festes Nahrungsmittel aus Milch, das normalerweise gelb oder weiß ist und einen salzigen Geschmack hat."
    },
    {
      "word": "verrückt",
      "translations": [
        {
          "pos": "Adjective",
          "translations": [
            "crazy",
            "insane"
          ]
        }
      ],
      "transcription": "/fɛɐˈrʏkt/",
      "context": "Bist du verrückt?",
      "context_translation": "Are you crazy?",
      "definition": "Nicht im Einklang mit der Vernunft oder dem gesunden Menschenverstand; geistig gestört."
    }
  ]
}`
