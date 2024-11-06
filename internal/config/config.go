package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	LogLevel   string
	Port       int
	MockOpenAI bool
	OpenAIKey  string
	APIKeys    []string
	Dynamo     AWS
	S3         S3
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

	return &conf, nil
}

func bindEnvs() error {
	configType := reflect.TypeOf(Config{})
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

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

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
