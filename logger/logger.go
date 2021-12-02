package logger

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

type Logger logrus.Logger

func CreateLogger(debug bool) *Logger {
	logLevel := logrus.InfoLevel
	if debug {
		logLevel = logrus.DebugLevel
	}

	return &Logger{
		Out: os.Stderr,
		Formatter: &logrus.TextFormatter{
			FullTimestamp:    true,
			TimestampFormat:  "02 Jan 06 15:04",
			PadLevelText:     true,
			QuoteEmptyFields: true,
		},
		Hooks:        make(logrus.LevelHooks),
		Level:        logLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
}

func (l *Logger) Error(err error) {
	(*logrus.Logger)(l).Error(fmt.Sprintf("%+v", err))
}

func (l *Logger) Info(format string, args ...interface{}) {
	(*logrus.Logger)(l).Infof(format, args...)
}
