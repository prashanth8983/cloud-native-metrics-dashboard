// pkg/logger/logger_test.go
package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

// captureOutput captures stdout and returns it as a string
func captureOutput(f func()) string {
	// Save the original stdout
	originalStdout := os.Stdout
	
	// Create a pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Execute the function that produces output
	f()
	
	// Close the write end of the pipe to flush it
	w.Close()
	
	// Read all from the pipe
	var buf bytes.Buffer
	io.Copy(&buf, r)
	
	// Restore the original stdout
	os.Stdout = originalStdout
	
	return buf.String()
}

func TestLoggerLevels(t *testing.T) {
	// Create a new logger with JSON format for easier parsing
	logger := New(Config{
		Level:  DebugLevel,
		Format: JSONFormat,
	})
	
	// Set it as the default
	SetDefault(logger)
	
	// Test debug level
	output := captureOutput(func() {
		Debug("debug message")
	})
	
	var logData map[string]interface{}
	err := json.Unmarshal([]byte(output), &logData)
	if err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}
	
	if logData["level"] != "debug" {
		t.Errorf("Expected level to be debug, got %v", logData["level"])
	}
	
	if logData["message"] != "debug message" {
		t.Errorf("Expected message to be 'debug message', got %v", logData["message"])
	}
	
	// Test info level
	output = captureOutput(func() {
		Info("info message")
	})
	
	err = json.Unmarshal([]byte(output), &logData)
	if err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}
	
	if logData["level"] != "info" {
		t.Errorf("Expected level to be info, got %v", logData["level"])
	}
	
	// Test warn level
	output = captureOutput(func() {
		Warn("warn message")
	})
	
	err = json.Unmarshal([]byte(output), &logData)
	if err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}
	
	if logData["level"] != "warn" {
		t.Errorf("Expected level to be warn, got %v", logData["level"])
	}
	
	// Test error level
	output = captureOutput(func() {
		Error("error message")
	})
	
	err = json.Unmarshal([]byte(output), &logData)
	if err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}
	
	if logData["level"] != "error" {
		t.Errorf("Expected level to be error, got %v", logData["level"])
	}
}

func TestLoggerFields(t *testing.T) {
	// Create a new logger with JSON format
	logger := New(Config{
		Level:  InfoLevel,
		Format: JSONFormat,
	})
	
	// Set it as the default
	SetDefault(logger)
	
	// Test with single field
	output := captureOutput(func() {
		Default().WithField("key", "value").Info().Msg("message with field")
	})
	
	var logData map[string]interface{}
	err := json.Unmarshal([]byte(output), &logData)
	if err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}
	
	if logData["key"] != "value" {
		t.Errorf("Expected field 'key' to be 'value', got %v", logData["key"])
	}
	
	// Test with multiple fields
	fields := map[string]interface{}{
		"int":    123,
		"string": "test",
		"bool":   true,
	}
	
	output = captureOutput(func() {
		Default().WithFields(fields).Info().Msg("message with fields")
	})
	
	err = json.Unmarshal([]byte(output), &logData)
	if err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}
	
	if int(logData["int"].(float64)) != 123 {
		t.Errorf("Expected field 'int' to be 123, got %v", logData["int"])
	}
	
	if logData["string"] != "test" {
		t.Errorf("Expected field 'string' to be 'test', got %v", logData["string"])
	}
	
	if logData["bool"] != true {
		t.Errorf("Expected field 'bool' to be true, got %v", logData["bool"])
	}
	
	// Test InfoWithFields
	output = captureOutput(func() {
		InfoWithFields("message with helper", fields)
	})
	
	err = json.Unmarshal([]byte(output), &logData)
	if err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}
	
	if int(logData["int"].(float64)) != 123 {
		t.Errorf("Expected field 'int' to be 123, got %v", logData["int"])
	}
}

func TestLoggerFormats(t *testing.T) {
	// Test JSON format
	jsonLogger := New(Config{
		Level:  InfoLevel,
		Format: JSONFormat,
	})
	
	SetDefault(jsonLogger)
	
	output := captureOutput(func() {
		Info("json format test")
	})
	
	// Should parse as JSON
	var logData map[string]interface{}
	err := json.Unmarshal([]byte(output), &logData)
	if err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}
	
	// Test text format
	textLogger := New(Config{
		Level:  InfoLevel,
		Format: TextFormat,
	})
	
	SetDefault(textLogger)
	
	output = captureOutput(func() {
		Info("text format test")
	})
	
	// Should be human-readable text
	if !strings.Contains(output, "text format test") {
		t.Errorf("Text format didn't contain expected message: %s", output)
	}
	
	// Shouldn't parse as JSON
	err = json.Unmarshal([]byte(output), &logData)
	if err == nil {
		t.Errorf("Text format output parsed as JSON, but should not have")
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	// Create a logger with warn level
	warnLogger := New(Config{
		Level:  WarnLevel,
		Format: JSONFormat,
	})
	
	SetDefault(warnLogger)
	
	// Debug and Info should not produce output
	debugOutput := captureOutput(func() {
		Debug("debug message")
	})
	
	if debugOutput != "" {
		t.Errorf("Expected no output for Debug with WarnLevel logger, got: %s", debugOutput)
	}
	
	infoOutput := captureOutput(func() {
		Info("info message")
	})
	
	if infoOutput != "" {
		t.Errorf("Expected no output for Info with WarnLevel logger, got: %s", infoOutput)
	}
	
	// Warn and Error should produce output
	warnOutput := captureOutput(func() {
		Warn("warn message")
	})
	
	if warnOutput == "" {
		t.Error("Expected output for Warn with WarnLevel logger, got nothing")
	}
	
	errorOutput := captureOutput(func() {
		Error("error message")
	})
	
	if errorOutput == "" {
		t.Error("Expected output for Error with WarnLevel logger, got nothing")
	}
}

func TestLoggerConfigUpdate(t *testing.T) {
	// Create a logger with debug level
	logger := New(Config{
		Level:  DebugLevel,
		Format: JSONFormat,
	})
	
	SetDefault(logger)
	
	// Debug should produce output
	debugOutput := captureOutput(func() {
		Debug("debug message")
	})
	
	if debugOutput == "" {
		t.Error("Expected output for Debug with DebugLevel logger, got nothing")
	}
	
	// Update to error level
	logger.UpdateConfig(Config{
		Level:  ErrorLevel,
		Format: JSONFormat,
	})
	
	// Debug should not produce output now
	debugOutput = captureOutput(func() {
		Debug("debug message")
	})
	
	if debugOutput != "" {
		t.Errorf("Expected no output for Debug after updating to ErrorLevel, got: %s", debugOutput)
	}
	
	// Error should still produce output
	errorOutput := captureOutput(func() {
		Error("error message")
	})
	
	if errorOutput == "" {
		t.Error("Expected output for Error after updating to ErrorLevel, got nothing")
	}
}

func TestLoggerWithError(t *testing.T) {
	// Create a logger with JSON format
	logger := New(Config{
		Level:  InfoLevel,
		Format: JSONFormat,
	})
	
	SetDefault(logger)
	
	// Create a test error
	testErr := errors.New("test error")
	
	// Log with error
	output := captureOutput(func() {
		logger.WithError(testErr).Error().Msg("error occurred")
	})
	
	var logData map[string]interface{}
	err := json.Unmarshal([]byte(output), &logData)
	if err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}
	
	// Check the error field
	errorField, ok := logData["error"].(string)
	if !ok {
		t.Fatalf("Expected error field to be string, got: %T", logData["error"])
	}
	
	if errorField != "test error" {
		t.Errorf("Expected error field to be 'test error', got: %s", errorField)
	}
	
	// Test ErrorWithError helper
	output = captureOutput(func() {
		ErrorWithError("helper error message", testErr)
	})
	
	err = json.Unmarshal([]byte(output), &logData)
	if err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}
	
	// Check the message and error field
	if logData["message"] != "helper error message" {
		t.Errorf("Expected message to be 'helper error message', got: %v", logData["message"])
	}
	
	errorField, ok = logData["error"].(string)
	if !ok {
		t.Fatalf("Expected error field to be string, got: %T", logData["error"])
	}
	
	if errorField != "test error" {
		t.Errorf("Expected error field to be 'test error', got: %s", errorField)
	}
}