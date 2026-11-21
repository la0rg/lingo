package config

import (
	"bytes"
	"log/slog"
	"reflect"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAllKeys(t *testing.T) {
	t.Parallel()

	keys := getAllKeys(reflect.TypeOf(Config{}))

	assert.Subset(t, keys, []string{
		"loglevel", "port", "mockopenai", "openaikey", "apikeys",
		"dynamo.region", "dynamo.endpoint", "dynamo.key", "dynamo.secret",
		"s3.region", "s3.endpoint", "s3.key", "s3.secret", "s3.bucket", // squashed AWS fields
		"limits.maxwords", "limits.maxwordlength", "limits.wordsperminute",
		"limits.clientipheader", "limits.trustedproxies",
	})
	assert.NotContains(t, keys, "s3.aws.region")
}

// Load uses the global viper instance, so these tests must not run in parallel.
func TestLoad_FromEnv(t *testing.T) {
	t.Cleanup(viper.Reset)
	t.Chdir(t.TempDir()) // no config.yaml
	t.Setenv("HOME", t.TempDir())

	t.Setenv("PORT", "9090")
	t.Setenv("OPENAIKEY", "sk-test")
	t.Setenv("S3_BUCKET", "bucket")
	t.Setenv("S3_REGION", "eu-west-1")
	t.Setenv("LIMITS_MAXWORDS", "50")
	t.Setenv("LIMITS_TRUSTEDPROXIES", "10.0.0.0/8,2400:cb00::/32")

	conf, err := Load()
	require.NoError(t, err)

	assert.Equal(t, 9090, conf.Port)
	assert.Equal(t, "info", conf.LogLevel, "default is kept")
	assert.Equal(t, "sk-test", conf.OpenAIKey)
	assert.Equal(t, "bucket", conf.S3.Bucket)
	assert.Equal(t, "eu-west-1", conf.S3.Region)
	assert.Equal(t, Limits{
		MaxWords:       50,
		MaxWordLength:  64,
		WordsPerMinute: 200,
		TrustedProxies: []string{"10.0.0.0/8", "2400:cb00::/32"},
	}, conf.Limits)
}

func TestLoad_InvalidLimits(t *testing.T) {
	t.Cleanup(viper.Reset)
	t.Chdir(t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("LIMITS_MAXWORDS", "500")

	_, err := Load()
	assert.ErrorContains(t, err, "limits.wordsperminute")
}

func TestLimits_Validate(t *testing.T) {
	t.Parallel()

	assert.NoError(t, Limits{MaxWords: 10, WordsPerMinute: 10}.Validate())
	assert.Error(t, Limits{MaxWords: 11, WordsPerMinute: 10}.Validate())
	assert.NoError(t, Limits{TrustedProxies: []string{"173.245.48.0/20", "2400:cb00::/32"}}.Validate())
	assert.ErrorContains(t, Limits{TrustedProxies: []string{"173.245.48.0"}}.Validate(), "173.245.48.0")
}

func TestConfig_LogValueRedactsSecrets(t *testing.T) {
	t.Parallel()

	conf := &Config{
		Port:      8080,
		OpenAIKey: "sk-secret-openai",
		APIKeys:   []string{"api-secret-1", "api-secret-2"},
		Dynamo:    AWS{Region: "us-east-1", Key: "dynamo-key", Secret: "dynamo-secret"},
		S3:        S3{AWS: AWS{Key: "s3-key", Secret: "s3-secret"}, Bucket: "lingo"},
	}

	var buf bytes.Buffer
	slog.New(slog.NewJSONHandler(&buf, nil)).Info("starting", "config", conf)
	out := buf.String()

	for _, secret := range []string{"sk-secret-openai", "api-secret-1", "api-secret-2", "dynamo-key",
		"dynamo-secret", "s3-key", "s3-secret"} {
		assert.NotContains(t, out, secret)
	}
	assert.Contains(t, out, `"openaikey":"[REDACTED]"`)
	assert.Contains(t, out, `"apikeys_count":2`)
	assert.Contains(t, out, `"region":"us-east-1"`)
	assert.Contains(t, out, `"bucket":"lingo"`)
}

func TestConfig_LogValueKeepsEmptySecretsEmpty(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	slog.New(slog.NewJSONHandler(&buf, nil)).Info("starting", "config", Config{})

	assert.Contains(t, buf.String(), `"openaikey":""`, "an empty key must stay visible as missing")
}
