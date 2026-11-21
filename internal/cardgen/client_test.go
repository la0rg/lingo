package cardgen

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerate_RealClient runs the real SDK client against a fake server to check the request on the wire
// and that transient errors are retried by the client.
func TestGenerate_RealClient(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	var body map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer sk-test", r.Header.Get("Authorization"))

		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		data, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.NoError(t, json.Unmarshal(data, &body))

		content, _ := json.Marshal(apfelJSON)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"1","object":"chat.completion","model":"gpt-6-luna","choices":[{"index":0,` +
			`"finish_reason":"stop","message":{"role":"assistant","content":` + string(content) + `}}],` +
			`"usage":{"prompt_tokens":500,"completion_tokens":200,"total_tokens":700}}`))
	}))
	t.Cleanup(srv.Close)

	client := openai.NewClient(
		option.WithAPIKey("sk-test"), option.WithBaseURL(srv.URL), option.WithMaxRetries(2),
		option.WithHTTPClient(srv.Client()),
	)
	g := newGenerator(t, &client.Chat.Completions)

	card, err := g.Generate(context.Background(), deEn, "apfel")

	require.NoError(t, err)
	assert.Equal(t, "Apfel", card.Word)
	assert.Equal(t, int32(2), calls.Load(), "5xx is retried by the client")

	assert.Equal(t, "gpt-6-luna", body["model"])
	assert.Equal(t, "none", body["reasoning_effort"])
	assert.EqualValues(t, tokenLimit, body["max_completion_tokens"])
	assert.NotContains(t, body, "max_tokens")

	format := body["response_format"].(map[string]any)
	assert.Equal(t, "json_schema", format["type"])
	schema := format["json_schema"].(map[string]any)
	assert.Equal(t, "card", schema["name"])
	assert.Equal(t, true, schema["strict"])
}
