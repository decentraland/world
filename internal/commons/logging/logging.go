package logging

import (
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

type Logger = logrus.Logger

// SetLevel set logger level given a string level
func SetLevel(logger *logrus.Logger, level string) error {
	lvl, err := logrus.ParseLevel(level)

	if err != nil {
		return err
	}

	logger.SetLevel(lvl)
	return nil
}

// LogPanic can be used deferred in the main function to capture panic logs
func LogPanic() {
	if r := recover(); r != nil {
		err, ok := r.(error)
		if ok {
			debug.PrintStack()
			logrus.WithError(err).Error("panic")
		}
	}
}

// LoggerConfig represents the logger config
type LoggerConfig struct {
	JSONDisabled bool
}

// New returns a new logrus Logger with our config
func New(config *LoggerConfig) *logrus.Logger {
	log := logrus.New()
	if !config.JSONDisabled {
		formatter := logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime: "@timestamp",
			},
		}

		log.SetFormatter(&formatter)
	}
	return log
}
