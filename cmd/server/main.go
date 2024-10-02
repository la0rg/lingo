package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/la0rg/lingo/internal/app"
	"github.com/la0rg/lingo/internal/config"
)

func main() {
	conf, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	initLogger(conf)
	slog.Info("Config loaded", "config", conf)

	if err := app.Run(context.Background(), conf); err != nil {
		log.Fatal(err)
	}
}

func initLogger(conf *config.Config) {
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(conf.LogLevel)); err != nil {
		slog.Warn("Invalid log level in config", "log_level", conf.LogLevel)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)
}
