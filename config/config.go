package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	LogLevel string `mapstructure:"log_level"`
	Port     int    `mapstructure:"port"`
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
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil && !errors.As(err, new(viper.ConfigFileNotFoundError)) {
		return nil, fmt.Errorf("unable to read config file: %w", err)
	}

	if err := viper.Unmarshal(&conf); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config: %w", err)
	}

	return &conf, nil
}
