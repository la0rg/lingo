package cardgen

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	oai "github.com/sashabaranov/go-openai"

	"github.com/la0rg/lingo/internal/cardgen/prompt"
	"github.com/la0rg/lingo/internal/logger"
	"github.com/la0rg/lingo/internal/model"
)

const (
	tokenLimit  = 1000
	openAImodel = oai.GPT4o
	retries     = 2
	retryDelay  = 100 * time.Millisecond
)

type OpenAI struct {
	client GPTClient
}

type request struct {
	SourceLanguage string `json:"source_language"`
	TargetLanguage string `json:"target_language"`
	Word           string `json:"word"`
}

type GPTClient interface {
	CreateChatCompletion(ctx context.Context, request oai.ChatCompletionRequest) (oai.ChatCompletionResponse, error)
}

func NewOpenAI(client GPTClient) (*OpenAI, error) {
	return &OpenAI{client: client}, nil
}

func (g OpenAI) Generate(ctx context.Context, collection *model.Collection, word string) (*model.Card, error) {
	request, err := json.Marshal(request{
		SourceLanguage: collection.SourceLanguage,
		TargetLanguage: collection.TargetLanguage,
		Word:           word,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to marshal request: %v", err)
	}

	return g.retryRequest(ctx, string(request))
}

// retryRequest simply retry to remedy random OpenAI hallucinations
func (g OpenAI) retryRequest(ctx context.Context, req string) (*model.Card, error) {
	var card *model.Card
	var err error

	for i := 0; i <= retries; i++ {
		card, err = g.makeRequest(ctx, req)
		if err == nil && card != nil {
			return card, nil
		}

		if i == retries {
			break
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryDelay):
		}
	}
	return card, err
}

func (g OpenAI) makeRequest(ctx context.Context, req string) (*model.Card, error) {
	resp, err := g.client.CreateChatCompletion(
		ctx,
		oai.ChatCompletionRequest{
			MaxCompletionTokens: tokenLimit,
			Model:               openAImodel,
			ResponseFormat: &oai.ChatCompletionResponseFormat{
				Type: oai.ChatCompletionResponseFormatTypeJSONObject,
			},
			Messages: []oai.ChatCompletionMessage{
				{Role: oai.ChatMessageRoleSystem, Content: prompt.System},
				{Role: oai.ChatMessageRoleUser, Content: req},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cards: %v", err)
	}

	if len(resp.Choices) < 1 {
		return nil, nil
	}

	logger.WithContext(ctx).Debug("OpenAI response", "response", resp)
	if resp.Choices[0].FinishReason != oai.FinishReasonStop {
		logger.WithContext(ctx).Warn("Incomplete OpenAI response", "finish_reason", resp.Choices[0].FinishReason)
	}

	content := []byte(resp.Choices[0].Message.Content)

	var card model.Card
	if err := json.Unmarshal(content, &card); err != nil {
		logger.WithContext(ctx).Warn("OpenAI response is not a valid json", "response", resp.Choices[0].Message.Content)
		return nil, fmt.Errorf("unable to parse OpenAI response: %v", err)
	}

	if card.Word == "" || len(card.Translations) == 0 {
		logger.WithContext(ctx).Warn("OpenAI response is empty", "response", resp.Choices[0].Message.Content)
		return nil, nil
	}

	return &card, nil
}
