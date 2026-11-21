package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	LogLevel   string
	Port       int
	MockOpenAI bool
	OpenAIKey  string
	// OpenAIModel overrides the default card generation model (cardgen.DefaultModel).
	OpenAIModel string
	// OpenAIReasoningEffort overrides the default reasoning effort (cardgen.DefaultReasoningEffort).
	OpenAIReasoningEffort string
	APIKeys               []string
	Dynamo                AWS
	S3                    S3
	Limits                Limits
}

// Limits protect the paid OpenAI-backed endpoints from abuse.
type Limits struct {
	// MaxWords is the maximum number of words accepted in a single collection request.
	MaxWords int
	// MaxWordLength is the maximum length of a single word, in runes.
	MaxWordLength int
	// WordsPerMinute is the per-client number of words accepted by the public (unauthenticated) endpoints.
	// It must be >= MaxWords.
	WordsPerMinute int
	// ClientIPHeader is a proxy header that holds the client IP (e.g. "CF-Connecting-IP").
	// Leave it empty to use the TCP remote address.
	ClientIPHeader string
	// TrustedProxies are the CIDR prefixes of the proxy. ClientIPHeader is read only for requests coming from them.
	// Leave it empty only if the network guarantees that the proxy is the only way to reach the server.
	TrustedProxies []string
}

// Validate checks that the limits are consistent.
func (l Limits) Validate() error {
	if l.WordsPerMinute < l.MaxWords {
		return fmt.Errorf("limits.wordsperminute (%d) must be >= limits.maxwords (%d)", l.WordsPerMinute, l.MaxWords)
	}

	for _, prefix := range l.TrustedProxies {
		if _, err := netip.ParsePrefix(prefix); err != nil {
			return fmt.Errorf("invalid limits.trustedproxies prefix %q: %w", prefix, err)
		}
	}

	return nil
}

const redacted = "[REDACTED]"

// LogValue implements slog.LogValuer so secrets never end up in logs.
func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("loglevel", c.LogLevel),
		slog.Int("port", c.Port),
		slog.Bool("mockopenai", c.MockOpenAI),
		slog.String("openaikey", redactString(c.OpenAIKey)),
		slog.String("openaimodel", c.OpenAIModel),
		slog.String("openaireasoningeffort", c.OpenAIReasoningEffort),
		slog.Int("apikeys_count", len(c.APIKeys)),
		slog.Any("dynamo", c.Dynamo),
		slog.Any("s3", c.S3),
		slog.Any("limits", c.Limits),
	)
}

func (a AWS) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("region", a.Region),
		slog.String("endpoint", a.Endpoint),
		slog.String("key", redactString(a.Key)),
		slog.String("secret", redactString(a.Secret)),
	)
}

func (s S3) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("aws", s.AWS),
		slog.String("bucket", s.Bucket),
	)
}

func redactString(s string) string {
	if s == "" {
		return ""
	}
	return redacted
}

type S3 struct {
	AWS    `mapstructure:",squash"`
	Bucket string
}

type AWS struct {
	Region   string
	Endpoint string
	Key      string
	Secret   string
}

func Load() (*Config, error) {
	var conf = Config{
		LogLevel: "info",
		Port:     8080,
		Limits: Limits{
			MaxWords:       100,
			MaxWordLength:  64,
			WordsPerMinute: 200,
		},
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.lingo")
	viper.AddConfigPath(".")

	if err := bindEnvs(); err != nil {
		return nil, fmt.Errorf("unable to bind envs: %w", err)
	}

	err := viper.ReadInConfig()
	if err != nil && !errors.As(err, new(viper.ConfigFileNotFoundError)) {
		return nil, fmt.Errorf("unable to read config file: %w", err)
	}

	if err := viper.Unmarshal(&conf); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config: %w", err)
	}

	if err := conf.Limits.Validate(); err != nil {
		return nil, err
	}

	return &conf, nil
}

func bindEnvs() error {
	configType := reflect.TypeFor[Config]()
	keys := getAllKeys(configType)

	for _, key := range keys {
		envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		if err := viper.BindEnv(key, envKey); err != nil {
			return err
		}
	}

	return nil
}

func getAllKeys(t reflect.Type) []string {
	var result []string

	for f := range t.Fields() {
		name := strings.ToLower(f.Name)
		squashParentName := strings.Contains(f.Tag.Get("mapstructure"), "squash")

		if f.Type.Kind() == reflect.Struct {
			subKeys := getAllKeys(f.Type)
			for _, key := range subKeys {
				if squashParentName {
					result = append(result, key)
					continue
				}

				result = append(result, fmt.Sprintf("%s.%s", name, key))
			}

			continue
		}

		result = append(result, name)
	}

	return result
}
