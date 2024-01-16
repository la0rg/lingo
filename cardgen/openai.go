package cardgen

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"

	"github.com/la0rg/lingo/cardgen/prompt"
	"github.com/la0rg/lingo/logger"
)

const tokenLimit = 2000

type OpenAI struct {
	client GPTClient
}

type GPTClient interface {
	CreateChatCompletion(
		ctx context.Context,
		request openai.ChatCompletionRequest,
	) (openai.ChatCompletionResponse, error)
}

func NewOpenAI(client GPTClient) (*OpenAI, error) {
	return &OpenAI{client: client}, nil
}

func (g OpenAI) Generate(ctx context.Context, sourceLanguage, targetLanguage string, words ...string) (*Cards, error) {
	request, err := json.Marshal(request{
		SourceLanguage: sourceLanguage,
		TargetLanguage: targetLanguage,
		Words:          words,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to marshal request: %v", err)
	}

	resp, err := g.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			MaxTokens: tokenLimit,
			Model:     openai.GPT3Dot5Turbo1106,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: prompt.System},
				{Role: openai.ChatMessageRoleUser, Content: string(request)},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cards: %v", err)
	}

	if len(resp.Choices) < 1 {
		return nil, fmt.Errorf("empty OpenAI response")
	}

	logger.WithContext(ctx).Debug("OpenAI response", "response", resp)
	if resp.Choices[0].FinishReason != openai.FinishReasonStop {
		logger.WithContext(ctx).Warn("Incomplete OpenAI response", "finish_reason", resp.Choices[0].FinishReason)
	}

	response := []byte(resp.Choices[0].Message.Content)

	var cards Cards
	if err := json.Unmarshal(response, &cards); err != nil {
		logger.WithContext(ctx).Warn("OpenAI response is not a valid json", "response", resp.Choices[0].Message.Content)
		return nil, fmt.Errorf("unable to parse OpenAI response: %v", err)
	}

	return &cards, nil
}
