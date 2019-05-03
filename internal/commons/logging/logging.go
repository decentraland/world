package logging

import (
	"github.com/sirupsen/logrus"
)

func SetLevel(logger *logrus.Logger, level string) error {
	lvl, err := logrus.ParseLevel(level)

	if err != nil {
		return err
	}

	logger.SetLevel(lvl)
	return nil
}

func LogPanic() {
	if r := recover(); r != nil {
		err, ok := r.(error)
		if ok {
			logrus.WithError(err).Error("panic")
		}
	}
}

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
