package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"	
)

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	*zerolog.Logger
}

// Config holds logger configuration
type Config struct {
	Level  string
	Format string
}

// New creates a new logger instance
func New(cfg Config) *Logger {
	// Set log level
	level, err := zerolog.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output format
	var logger zerolog.Logger
	if cfg.Format == "console" {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Caller().Logger()
	} else {
		logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	}

	return &Logger{Logger: &logger}
}

// WithRequestID adds request ID to logger context
func (l *Logger) WithRequestID(requestID string) *Logger {
	newLogger := l.Logger.With().Str("request_id", requestID).Logger()
	return &Logger{Logger: &newLogger}
}

// WithComponent adds component name to logger context
func (l *Logger) WithComponent(component string) *Logger {
	newLogger := l.Logger.With().Str("component", component).Logger()
	return &Logger{Logger: &newLogger}
}

// WithError adds error to logger context
func (l *Logger) WithError(err error) *Logger {
	newLogger := l.Logger.With().Err(err).Logger()
	return &Logger{Logger: &newLogger}
}

// InfoWithFields logs info message with additional fields
func (l *Logger) InfoWithFields(msg string, fields map[string]interface{}) {
	event := l.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// ErrorWithFields logs error message with additional fields
func (l *Logger) ErrorWithFields(msg string, fields map[string]interface{}) {
	event := l.Error()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Global logger instance
var global *Logger

// SetGlobal sets the global logger instance
func SetGlobal(l *Logger) {
	global = l
}

// Global returns the global logger instance
func Global() *Logger {
	if global == nil {
		global = New(Config{Level: "info", Format: "json"})
	}
	return global
}

// Info logs info message using global logger
func Info() *zerolog.Event {
	return Global().Info()
}

// Error logs error message using global logger
func Error() *zerolog.Event {
	return Global().Error()
}

// Debug logs debug message using global logger
func Debug() *zerolog.Event {
	return Global().Debug()
}

// Warn logs warning message using global logger
func Warn() *zerolog.Event {
	return Global().Warn()
}
