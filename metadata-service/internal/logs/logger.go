package logs

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// Logger is a global logger instance used throughout the application.
var Logger zerolog.Logger
// InitLogger initializes the global logger with the specified log level and output.
func InitLogger(level zerolog.Level) {
	// Set the global logger with the specified log level and output
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	Logger = zerolog.New(os.Stdout).Level(level)
	
	Logger = Logger.With().Str("service", "metadata-service").Timestamp().Logger()
}