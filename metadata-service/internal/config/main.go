package config

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/sethvargo/go-envconfig"
)

type IncusConfig struct {
	// ServerURL is the URL of the Incus server.
	ServerURL string `env:"SERVER_URL,default=https://localhost:8443"`
	// TLSConfig holds the TLS configuration for connecting to the Incus server.
	TLSClientCert string `env:"TLS_CLIENT_CERT,default=/etc/incus/client.crt"`
	// TLSClientKey is the path to the client key for TLS connections.
	TLSClientKey string `env:"TLS_CLIENT_KEY,default=/etc/incus/client.key"`
	TLSServerCert string `env:"TLS_SERVER_CERT,default="` // Optional, can be left empty to use default server certificate handling
	TLSInsecureSkipVerify bool `env:"TLS_INSECURE_SKIP_VERIFY,default=false"` // Skip certificate verification for self-signed certs
}

type DatabaseConfig struct {
	// DBDriver is the database driver to use (e.g., sqlite, postgres).
	DBDriver string `env:"DB_DRIVER,default=sqlite"`
	// DBSource is the data source name for the database connection.
	DBSource string `env:"DB_SOURCE,default=metadata.db"`
}

// Config holds the configuration for the metadata service.
type Config struct {
	// Port is the port on which the metadata service will run.
	Port string `env:"PORT,default=8080"`
	// LogLevel sets the logging level for the service.
	LogLevel zerolog.Level `env:"LOG_LEVEL,default=info"`
	// Incus contains the configuration for connecting to the Incus server.
	Incus *IncusConfig `env:",prefix=INCUS_CONFIG_"`
	// Database contains the configuration for connecting to the database.
	Database *DatabaseConfig `env:",prefix=DATABASE_CONFIG_"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
