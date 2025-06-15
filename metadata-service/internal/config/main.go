package config

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/sethvargo/go-envconfig"
)

// Config holds the configuration for the metadata service.
type Config struct {
	// Port is the port on which the metadata service will run.
	Port string `env:"PORT,default=8080"`
	// LogLevel sets the logging level for the service.
	LogLevel zerolog.Level `env:"LOG_LEVEL,default=info"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}