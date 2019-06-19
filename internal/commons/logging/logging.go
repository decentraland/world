package logging

import (
	"github.com/sirupsen/logrus"
)

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
			logrus.WithError(err).Error("panic")
		}
	}
}

// New returns a new logrus Logger with our config
func New() *logrus.Logger {
	log := logrus.New()
	formatter := logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime: "@timestamp",
		},
	}

	log.SetFormatter(&formatter)
	return log
}
