package cardgen

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/la0rg/lingo/internal/cardgen/mocks"
	"github.com/la0rg/lingo/internal/model"
)

var deEn = &model.Collection{SourceLanguage: "de", TargetLanguage: "en"}

const apfelJSON = `{"word":"Apfel","plural":"Äpfel","translations":[{"pos":"Nomen","translations":["apple"]}],` +
	`"article":"der","gender":"m.","transcription":"/ˈapfəl/","context":"Der Apfel ist rot.",` +
	`"context_translation":"The apple is red.","definition":"Eine Frucht."}`

func completion(content, finishReason string) *openai.ChatCompletion {
	return &openai.ChatCompletion{Choices: []openai.ChatCompletionChoice{{
		Message:      openai.ChatCompletionMessage{Content: content},
		FinishReason: finishReason,
	}}}
}

func newGenerator(t *testing.T, client GPTClient, opts ...Option) *OpenAI {
	t.Helper()
	g, err := NewOpenAI(client, opts...)
	require.NoError(t, err)
	return g
}

func TestGenerate(t *testing.T) {
	t.Parallel()

	client := mocks.NewGPTClient(t)
	var params openai.ChatCompletionNewParams
	client.EXPECT().New(mock.Anything, mock.Anything).
		RunAndReturn(func(
			_ context.Context, p openai.ChatCompletionNewParams, _ ...option.RequestOption,
		) (*openai.ChatCompletion, error) {
			params = p
			return completion(apfelJSON, "stop"), nil
		}).Once()

	card, err := newGenerator(t, client).Generate(context.Background(), deEn, "apfel")

	require.NoError(t, err)
	assert.Equal(t, &model.Card{
		Word:               "Apfel",
		Plural:             "Äpfel",
		Translations:       model.Translations{{PartOfSpeech: "Nomen", Translations: []string{"apple"}}},
		Article:            "der",
		Gender:             "m.",
		Transcription:      "/ˈapfəl/",
		Context:            "Der Apfel ist rot.",
		ContextTranslation: "The apple is red.",
		Definition:         "Eine Frucht.",
	}, card)

	assert.Equal(t, DefaultModel, params.Model)
	assert.Equal(t, shared.ReasoningEffortNone, params.ReasoningEffort)
	assert.Equal(t, int64(tokenLimit), params.MaxCompletionTokens.Value)

	schema := params.ResponseFormat.OfJSONSchema
	require.NotNil(t, schema, "structured outputs must be used")
	assert.True(t, schema.JSONSchema.Strict.Value)
	assert.Equal(t, cardSchema, schema.JSONSchema.Schema)

	require.Len(t, params.Messages, 2)
	require.NotNil(t, params.Messages[0].OfSystem)
	assert.Contains(t, params.Messages[0].OfSystem.Content.OfString.Value, "You are a dictionary API")
	require.NotNil(t, params.Messages[1].OfUser)
	assert.JSONEq(t, `{"source_language":"de","target_language":"en","word":"apfel"}`,
		params.Messages[1].OfUser.Content.OfString.Value)
}

func TestGenerate_Options(t *testing.T) {
	t.Parallel()

	client := mocks.NewGPTClient(t)
	client.EXPECT().New(mock.Anything, mock.MatchedBy(func(p openai.ChatCompletionNewParams) bool {
		return p.Model == "gpt-6.1-terra" && p.ReasoningEffort == shared.ReasoningEffortLow
	})).Return(completion(apfelJSON, "stop"), nil)

	g := newGenerator(t, client, WithModel("gpt-6.1-terra"), WithReasoningEffort("low"))
	_, err := g.Generate(context.Background(), deEn, "apfel")

	require.NoError(t, err)
}

func TestGenerate_InvalidWord(t *testing.T) {
	t.Parallel()

	cases := map[string]*openai.ChatCompletion{
		"empty word":      completion(`{"word":"","translations":[]}`, "stop"),
		"no translations": completion(`{"word":"Qwrt","translations":[]}`, "stop"),
		"refusal": {Choices: []openai.ChatCompletionChoice{{
			Message: openai.ChatCompletionMessage{Refusal: "I can't help with that."}, FinishReason: "stop",
		}}},
	}

	for name, resp := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			client := mocks.NewGPTClient(t)
			client.EXPECT().New(mock.Anything, mock.Anything).Return(resp, nil).Once() // no retries

			card, err := newGenerator(t, client).Generate(context.Background(), deEn, "qwrt")

			require.NoError(t, err)
			assert.Nil(t, card)
		})
	}
}

func TestGenerate_Errors(t *testing.T) {
	t.Parallel()

	apiErr := errors.New("401 unauthorized")
	cases := []struct {
		name    string
		resp    *openai.ChatCompletion
		err     error
		wantErr error
		message string
	}{
		{"api error", nil, apiErr, apiErr, ""},
		{"no choices", &openai.ChatCompletion{}, nil, ErrIncomplete, ""},
		{"truncated", completion(`{"word":"Apf`, "length"), nil, ErrIncomplete, `"length"`},
		{"content filter", completion("", "content_filter"), nil, ErrIncomplete, ""},
		{"invalid json", completion("not json", "stop"), nil, nil, "unable to parse"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			client := mocks.NewGPTClient(t)
			client.EXPECT().New(mock.Anything, mock.Anything).Return(c.resp, c.err).Once()

			card, err := newGenerator(t, client).Generate(context.Background(), deEn, "apfel")

			assert.Nil(t, card)
			require.Error(t, err)
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
			}
			if c.message != "" {
				assert.ErrorContains(t, err, c.message)
			}
		})
	}
}

// The schema must stay in sync with the json tags of model.Card.
func TestCardSchemaMatchesModel(t *testing.T) {
	t.Parallel()

	var fields map[string]any
	data, err := json.Marshal(model.Card{Translations: model.Translations{{}}})
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &fields))

	properties := cardSchema["properties"].(map[string]any)
	assert.Len(t, properties, len(fields))
	for name := range fields {
		assert.Contains(t, properties, name)
		assert.Contains(t, cardSchema["required"], name, "strict mode requires every property")
	}
}
