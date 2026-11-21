package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSystem(t *testing.T) {
	t.Parallel()

	assert.Contains(t, System, "You are a dictionary API")
	assert.Contains(t, System, `"context_translation"`)
}
