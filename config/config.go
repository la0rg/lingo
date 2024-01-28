package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	LogLevel  string `mapstructure:"log_level"`
	Port      int    `mapstructure:"port"`
	OpenAIKey string `mapstructure:"openai_key"`
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
	r := reflect.TypeOf(Config{})
	for i := 0; i < r.NumField(); i++ {
		f := r.Field(i)
		tag := f.Tag.Get("mapstructure")

		if err := viper.BindEnv(tag, strings.ToUpper(tag)); err != nil {
			return err
		}
	}

	return nil
}
