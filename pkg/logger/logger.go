package logger

import (
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Logger wraps logrus.Logger to provide a more streamlined API
type Logger struct {
	*logrus.Logger
}

// Fields type for structured logging fields
type Fields logrus.Fields

// NewLogger creates a new Logger with the given configuration
func NewLogger(level, format, output string) *Logger {
	log := logrus.New()

	// Configure output
	switch strings.ToLower(output) {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	default:
		file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.WithError(err).Error("Failed to open log file, falling back to stdout")
			log.SetOutput(os.Stdout)
		} else {
			log.SetOutput(file)
		}
	}

	// Configure log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		log.WithError(err).Warnf("Failed to parse log level '%s', falling back to info", level)
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)

	// Configure format
	switch strings.ToLower(format) {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	default:
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	}

	return &Logger{log}
}

// WithField adds a field to the log entry
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

// WithFields adds multiple fields to the log entry
func (l *Logger) WithFields(fields Fields) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields(fields))
}

// WithError adds an error field to the log entry
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// Configure changes logger configuration after creation
func (l *Logger) Configure(level, format string, output io.Writer) {
	if level != "" {
		logLevel, err := logrus.ParseLevel(level)
		if err == nil {
			l.SetLevel(logLevel)
		} else {
			l.WithError(err).Warnf("Failed to parse log level '%s'", level)
		}
	}

	if format != "" {
		switch strings.ToLower(format) {
		case "json":
			l.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			})
		case "text":
			l.SetFormatter(&logrus.TextFormatter{
				FullTimestamp:   true,
				TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			})
		default:
			l.Warnf("Unsupported log format '%s'", format)
		}
	}

	if output != nil {
		l.SetOutput(output)
	}
}
