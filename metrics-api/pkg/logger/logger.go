package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger defines the logging interface
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	WithFields(fields map[string]interface{}) Logger
	With(fields map[string]interface{}) Logger
	Sync() error
}

// zapLogger implements the Logger interface with zap
type zapLogger struct {
	logger *zap.SugaredLogger
}

// NewLogger creates a new Logger with the specified options
func NewLogger(opts ...Option) Logger {
	// Default configuration
	config := &loggerConfig{
		level:      zapcore.InfoLevel,
		outputType: "json",
		output:     os.Stdout,
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) { enc.AppendString(t.UTC().Format(time.RFC3339Nano)) }),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create encoder based on output type
	var encoder zapcore.Encoder
	if config.outputType == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(config.output),
		config.level,
	)

	// Create zap logger
	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return &zapLogger{ // Fixed: Use type name directly
		logger: logger.Sugar(),
	}
}

func (l *zapLogger) WithFields(fields map[string]interface{}) Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}

	newLogger := l.logger.Desugar().With(zapFields...)
	return &zapLogger{
		logger: newLogger.Sugar(),
	}
}


// NewNopLogger creates a logger that discards all logs
func NewNopLogger() Logger {
	return &zapLogger{ 
		logger: zap.NewNop().Sugar(),
	}
}

// Debug logs a debug message
func (l *zapLogger) Debug(args ...interface{}) {
	if l.logger != nil {
		l.logger.Debug(args...)
	}
}

// Debugf logs a formatted debug message
func (l *zapLogger) Debugf(format string, args ...interface{}) {
	if l.logger != nil {
		l.logger.Debugf(format, args...)
	}
}

// Info logs an info message
func (l *zapLogger) Info(args ...interface{}) {
	if l.logger != nil {
		l.logger.Info(args...)
	}
}

// Infof logs a formatted info message
func (l *zapLogger) Infof(format string, args ...interface{}) {
	if l.logger != nil {
		l.logger.Infof(format, args...)
	}
}

// Warn logs a warning message
func (l *zapLogger) Warn(args ...interface{}) {
	if l.logger != nil {
		l.logger.Warn(args...)
	}
}

// Warnf logs a formatted warning message
func (l *zapLogger) Warnf(format string, args ...interface{}) {
	if l.logger != nil {
		l.logger.Warnf(format, args...)
	}
}

// Error logs an error message
func (l *zapLogger) Error(args ...interface{}) {
	if l.logger != nil {
		l.logger.Error(args...)
	}
}

// Errorf logs a formatted error message
func (l *zapLogger) Errorf(format string, args ...interface{}) {
	if l.logger != nil {
		l.logger.Errorf(format, args...)
	}
}

// Fatal logs a fatal message and exits
func (l *zapLogger) Fatal(args ...interface{}) {
	if l.logger != nil {
		l.logger.Fatal(args...)
	}
}

// Fatalf logs a formatted fatal message and exits
func (l *zapLogger) Fatalf(format string, args ...interface{}) {
	if l.logger != nil {
		l.logger.Fatalf(format, args...)
	}
}

// With returns a logger with the specified fields
func (l *zapLogger) With(fields map[string]interface{}) Logger {
	if l.logger == nil {
		return l
	}
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &zapLogger{
		logger: l.logger.With(args...),
	}
}

// Sync flushes any buffered log entries
func (l *zapLogger) Sync() error {
	if l.logger == nil {
		return nil
	}
	return l.logger.Sync()
}

// loggerConfig holds the configuration for the logger
type loggerConfig struct {
	level      zapcore.Level
	outputType string
	output     io.Writer
}

// Option is a function that configures a loggerConfig
type Option func(*loggerConfig)

// WithLevel sets the log level
func WithLevel(level string) Option {
	return func(c *loggerConfig) {
		switch level {
		case "debug":
			c.level = zapcore.DebugLevel
		case "info":
			c.level = zapcore.InfoLevel
		case "warn":
			c.level = zapcore.WarnLevel
		case "error":
			c.level = zapcore.ErrorLevel
		case "fatal":
			c.level = zapcore.FatalLevel
		default:
			fmt.Fprintf(os.Stderr, "Invalid log level '%s', defaulting to info\n", level)
			c.level = zapcore.InfoLevel
		}
	}
}

// WithOutputType sets the output format (json or console)
func WithOutputType(outputType string) Option {
	return func(c *loggerConfig) {
		if outputType == "console" || outputType == "json" {
			c.outputType = outputType
		}
	}
}

// WithOutput sets the output writer
func WithOutput(output io.Writer) Option {
	return func(c *loggerConfig) {
		c.output = output
	}
}