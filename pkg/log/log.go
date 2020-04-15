package log

import "github.com/sirupsen/logrus"

func NewLoggerWithLevel(level logrus.Level) *logrus.Entry {
	log := logrus.New()
	log.SetLevel(level)

	return logrus.NewEntry(log)
}
