// pkg/logger/logger.go
package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Log levels
const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
)

// Formats
const (
	JSONFormat = "json"
	TextFormat = "text"
)

// Config holds the logger configuration
type Config struct {
	Level       string
	Format      string
	FilePath    string
	MaxSize     int // megabytes
	MaxBackups  int // number of files
	MaxAge      int // days
	Compression bool
}

// Logger is a wrapper around zerolog.Logger
type Logger struct {
	*zerolog.Logger
	config Config
	mu     sync.Mutex
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// New creates a new logger with the specified configuration
func New(config Config) *Logger {
	// Set up zerolog global settings
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.TimestampFieldName = "timestamp"
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "message"
	zerolog.CallerFieldName = "caller"

	// Create the logger instance
	logger := &Logger{
		config: config,
	}

	// Initialize the logger
	logger.configure()

	return logger
}

// configure sets up the logger based on the current configuration
func (l *Logger) configure() {
	l.mu.Lock()
	defer l.mu.Unlock()

	var writers []io.Writer

	// Set up console writer
	var consoleWriter io.Writer = os.Stdout // Default to os.Stdout
	if l.config.Format == TextFormat {
		cw := zerolog.ConsoleWriter{ // Create ConsoleWriter instance
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		consoleWriter = cw // Assign it as an io.Writer
	}
	writers = append(writers, consoleWriter)

	// Set up file writer if specified
	if l.config.FilePath != "" {
		fileWriter := &lumberjack.Logger{
			Filename:   l.config.FilePath,
			MaxSize:    l.config.MaxSize,
			MaxBackups: l.config.MaxBackups,
			MaxAge:     l.config.MaxAge,
			Compress:   l.config.Compression,
		}
		writers = append(writers, fileWriter)
	}

	// Create multi-writer for output to multiple destinations
	mw := io.MultiWriter(writers...)

	// Create the logger
	var level zerolog.Level
	switch l.config.Level {
	case DebugLevel:
		level = zerolog.DebugLevel
	case InfoLevel:
		level = zerolog.InfoLevel
	case WarnLevel:
		level = zerolog.WarnLevel
	case ErrorLevel:
		level = zerolog.ErrorLevel
	default:
		level = zerolog.InfoLevel
	}

	zl := zerolog.New(mw).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()

	l.Logger = &zl
}

// UpdateConfig updates the logger configuration
func (l *Logger) UpdateConfig(config Config) {
	l.config = config
	l.configure()
}

// Default gets the default logger, creating it if necessary
func Default() *Logger {
	once.Do(func() {
		defaultLogger = New(Config{
			Level:  InfoLevel,
			Format: TextFormat,
		})
	})
	return defaultLogger
}

// SetDefault sets the default logger
func SetDefault(logger *Logger) {
	defaultLogger = logger
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	// Create a new logger with the added field
	newLogger := l.Logger.With().Interface(key, value).Logger()

	// Return a new Logger with the updated zerolog.Logger
	return &Logger{
		Logger: &newLogger,
		config: l.config,
		mu:     sync.Mutex{},
	}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	// Create a context for building the new logger
	ctx := l.Logger.With()

	// Add all fields
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}

	// Create the new logger
	newLogger := ctx.Logger()

	// Return a new Logger with the updated zerolog.Logger
	return &Logger{
		Logger: &newLogger,
		config: l.config,
		mu:     sync.Mutex{},
	}
}

// Debug logs a debug message
func Debug(msg string, args ...interface{}) {
	if len(args) > 0 {
		Default().Debug().Msg(fmt.Sprintf(msg, args...))
	} else {
		Default().Debug().Msg(msg)
	}
}

// Info logs an info message
func Info(msg string, args ...interface{}) {
	if len(args) > 0 {
		Default().Info().Msg(fmt.Sprintf(msg, args...))
	} else {
		Default().Info().Msg(msg)
	}
}

// Warn logs a warning message
func Warn(msg string, args ...interface{}) {
	if len(args) > 0 {
		Default().Warn().Msg(fmt.Sprintf(msg, args...))
	} else {
		Default().Warn().Msg(msg)
	}
}

// Error logs an error message
func Error(msg string, args ...interface{}) {
	if len(args) > 0 {
		Default().Error().Msg(fmt.Sprintf(msg, args...))
	} else {
		Default().Error().Msg(msg)
	}
}

// DebugWithFields logs a debug message with fields
func DebugWithFields(msg string, fields map[string]interface{}) {
	ctx := Default().Debug()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	ctx.Msg(msg)
}

// InfoWithFields logs an info message with fields
func InfoWithFields(msg string, fields map[string]interface{}) {
	ctx := Default().Info()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	ctx.Msg(msg)
}

// WarnWithFields logs a warning message with fields
func WarnWithFields(msg string, fields map[string]interface{}) {
	ctx := Default().Warn()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	ctx.Msg(msg)
}

// ErrorWithFields logs an error message with fields
func ErrorWithFields(msg string, fields map[string]interface{}) {
	ctx := Default().Error()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	ctx.Msg(msg)
}

// WithError adds an error to the logger context
func (l *Logger) WithError(err error) *Logger {
	// Create a new logger with the added error
	newLogger := l.Logger.With().Err(err).Logger()

	// Return a new Logger with the updated zerolog.Logger
	return &Logger{
		Logger: &newLogger,
		config: l.config,
		mu:     sync.Mutex{},
	}
}

// ErrorWithError logs an error message with error details
func ErrorWithError(msg string, err error) {
	Default().WithError(err).Error().Msg(msg)
}