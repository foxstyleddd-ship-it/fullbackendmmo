package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes the global logger
func Init(level string, format string) {
	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	// Configure format
	if format == "pretty" {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	} else {
		// JSON format (default)
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}

	log.Info().
		Str("level", logLevel.String()).
		Str("format", format).
		Msg("Logger initialized")
}

// WithContext returns a logger with context fields
func WithContext(service string) zerolog.Logger {
	return log.With().
		Str("service", service).
		Timestamp().
		Logger()
}

// WithRequest returns a logger with request context
func WithRequest(requestID, accountID, characterID string) zerolog.Logger {
	return log.With().
		Str("request_id", requestID).
		Str("account_id", accountID).
		Str("character_id", characterID).
		Timestamp().
		Logger()
}
