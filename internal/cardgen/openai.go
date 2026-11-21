package cardgen

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"

	"github.com/la0rg/lingo/internal/cardgen/prompt"
	"github.com/la0rg/lingo/internal/logger"
	"github.com/la0rg/lingo/internal/model"
)

const (
	// DefaultModel is the cheapest current model; a dictionary lookup doesn't need a bigger one.
	DefaultModel = "gpt-6-luna"
	// DefaultReasoningEffort disables reasoning: it only adds cost and latency for dictionary lookups.
	DefaultReasoningEffort = shared.ReasoningEffortNone

	tokenLimit = 1000
)

// ErrIncomplete is returned when the model stops before completing the card.
var ErrIncomplete = errors.New("incomplete OpenAI response")

// GPTClient is the OpenAI chat completions API (openai.Client.Chat.Completions).
// Transport errors and rate limits are retried by the client itself.
type GPTClient interface {
	New(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (
		*openai.ChatCompletion, error)
}

type OpenAI struct {
	client          GPTClient
	model           string
	reasoningEffort shared.ReasoningEffort
	systemPrompt    string
}

type Option func(*OpenAI)

// WithModel overrides DefaultModel.
func WithModel(model string) Option {
	return func(g *OpenAI) { g.model = model }
}

// WithReasoningEffort overrides DefaultReasoningEffort.
func WithReasoningEffort(effort string) Option {
	return func(g *OpenAI) { g.reasoningEffort = shared.ReasoningEffort(effort) }
}

func NewOpenAI(client GPTClient, opts ...Option) (*OpenAI, error) {
	systemPrompt, err := prompt.System()
	if err != nil {
		return nil, fmt.Errorf("unable to render system prompt: %w", err)
	}

	g := &OpenAI{
		client:          client,
		model:           DefaultModel,
		reasoningEffort: DefaultReasoningEffort,
		systemPrompt:    systemPrompt,
	}
	for _, opt := range opts {
		opt(g)
	}

	return g, nil
}

type request struct {
	SourceLanguage string `json:"source_language"`
	TargetLanguage string `json:"target_language"`
	Word           string `json:"word"`
}

// Generate creates a card for the word. It returns a nil card if the word is not valid in the source language.
func (g *OpenAI) Generate(ctx context.Context, collection *model.Collection, word string) (*model.Card, error) {
	req, err := json.Marshal(request{
		SourceLanguage: collection.SourceLanguage,
		TargetLanguage: collection.TargetLanguage,
		Word:           word,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to marshal request: %w", err)
	}

	resp, err := g.client.New(ctx, g.params(string(req)))
	if err != nil {
		return nil, fmt.Errorf("failed to generate a card: %w", err)
	}

	log := logger.WithContext(ctx)
	log.Debug("OpenAI response",
		"model", resp.Model,
		"prompt_tokens", resp.Usage.PromptTokens,
		"completion_tokens", resp.Usage.CompletionTokens,
		"reasoning_tokens", resp.Usage.CompletionTokensDetails.ReasoningTokens,
	)

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("%w: no choices", ErrIncomplete)
	}

	choice := resp.Choices[0]
	if choice.Message.Refusal != "" {
		log.Warn("OpenAI refused to generate a card", "word", word, "refusal", choice.Message.Refusal)
		return nil, nil
	}

	if choice.FinishReason != "stop" {
		return nil, fmt.Errorf("%w: finish reason %q", ErrIncomplete, choice.FinishReason)
	}

	var card model.Card
	if err := json.Unmarshal([]byte(choice.Message.Content), &card); err != nil {
		log.Warn("OpenAI response is not a valid json", "response", choice.Message.Content)
		return nil, fmt.Errorf("unable to parse OpenAI response: %w", err)
	}

	if card.Word == "" || len(card.Translations) == 0 {
		log.Info("Word is not valid", "word", word)
		return nil, nil
	}

	return &card, nil
}

func (g *OpenAI) params(req string) openai.ChatCompletionNewParams {
	return openai.ChatCompletionNewParams{
		Model:               g.model,
		ReasoningEffort:     g.reasoningEffort,
		MaxCompletionTokens: openai.Int(tokenLimit),
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
				JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:   "card",
					Strict: openai.Bool(true),
					Schema: cardSchema,
				},
			},
		},
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(g.systemPrompt),
			openai.UserMessage(req),
		},
	}
}
