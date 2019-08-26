package logging

import (
	"os"
	"runtime/debug"

	"github.com/rs/zerolog"
)

// Logger ...
type Logger = zerolog.Logger

// LogPanic can be used deferred in the main function to capture panic logs
func LogPanic(log Logger) {
	if r := recover(); r != nil {
		err, ok := r.(error)
		if ok {
			debug.PrintStack()
			log.Error().Err(err).Msg("panic")
		}
	}
}

// LoggerConfig represents the logger config
type LoggerConfig struct {
	Level string
}

// New returns a new logger
func New(config *LoggerConfig) (Logger, error) {
	logger := zerolog.New(os.Stdout)

	lvl, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		return logger, err
	}

	logger = logger.Level(lvl).With().Timestamp().Logger()
	return logger, nil
}

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.TimestampFieldName = "@timestmap"
}
