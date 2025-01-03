package cardgen

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"

	"github.com/la0rg/lingo/internal/cardgen/prompt"
	"github.com/la0rg/lingo/internal/logger"
	"github.com/la0rg/lingo/internal/model"
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

func (g OpenAI) Generate(ctx context.Context, collection *model.Collection, words ...string) ([]model.Card, error) {
	request, err := json.Marshal(request{
		SourceLanguage: collection.SourceLanguage,
		TargetLanguage: collection.TargetLanguage,
		Words:          words,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to marshal request: %v", err)
	}

	resp, err := g.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			MaxTokens: tokenLimit,
			Model:     openai.GPT4o,
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
		return nil, nil
	}

	logger.WithContext(ctx).Debug("OpenAI response", "response", resp)
	if resp.Choices[0].FinishReason != openai.FinishReasonStop {
		logger.WithContext(ctx).Warn("Incomplete OpenAI response", "finish_reason", resp.Choices[0].FinishReason)
	}

	content := []byte(resp.Choices[0].Message.Content)

	var res response
	if err := json.Unmarshal(content, &res); err != nil {
		logger.WithContext(ctx).Warn("OpenAI response is not a valid json", "response", resp.Choices[0].Message.Content)
		return nil, fmt.Errorf("unable to parse OpenAI response: %v", err)
	}

	// TODO: Should it be here?
	id := time.Now().UnixMilli()
	for i := range res.Cards {
		res.Cards[i].ID = id + int64(i)
		res.Cards[i].CollectionID = collection.ID
	}

	return res.Cards, nil
}
