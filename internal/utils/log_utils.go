package utils

import (
	"context"
	"fmt"
	"goa.design/clue/log"
)

// aggregatedKVFielder is a helper struct that aggregates key-value pairs for logging.
type aggregatedKVFielder struct {
	fields []log.KV // Slice of key-value pairs for structured logging
}

// LogFields returns the aggregated fields as a slice of log.KV for logging.
func (a *aggregatedKVFielder) LogFields() []log.KV {
	return a.fields
}

// LogUtil provides a set of utility functions for structured logging with configurable formatting.
type LogUtil struct {
	Format log.FormatFunc // Format function to define the output format (JSON or Terminal)
}

// newLogUtil initializes a new LogUtil with optional configuration settings.
// It sets a default format based on the output destination (terminal or non-terminal).
func newLogUtil() LogUtil {
	format := log.FormatTerminal // Default format is JSON
	if log.IsTerminal() {
		format = log.FormatTerminal // Use terminal-friendly format if outputting to a terminal
	}
	return LogUtil{
		Format: format,
	}
}

// Debug logs a message at the Debug level with structured key-value pairs.
// It converts the provided data structure to key-value format and applies the specified format.
func (l LogUtil) Debug(ctx context.Context, kv interface{}) {
	ctx = log.Context(ctx, log.WithDebug(), log.WithFormat(l.Format)) // Set up context with debug logging and formatting
	userKVMap := Data.StructToKVMap(kv)                               // Convert structure to key-value map

	// Aggregate fields for logging
	aggregated := &aggregatedKVFielder{}
	for k, v := range userKVMap {
		aggregated.fields = append(aggregated.fields, log.KV{K: k, V: fmt.Sprintf("%v", v)})
	}

	log.Debug(ctx, aggregated) // Log at Debug level with aggregated fields
}

// Info logs a message at the Info level with structured key-value pairs.
// It dynamically sets the format based on the output destination and converts data to key-value format.
func (l LogUtil) Info(ctx context.Context, kv interface{}) {
	format := log.FormatTerminal // Default format is JSON
	if log.IsTerminal() {
		format = log.FormatTerminal // Use terminal format if outputting to a terminal
	}
	ctx = log.Context(ctx, log.WithDebug(), log.WithFormat(format))

	userKVMap := Data.StructToKVMap(kv) // Convert structure to key-value map
	aggregated := &aggregatedKVFielder{}

	// Aggregate fields for logging
	for k, v := range userKVMap {
		aggregated.fields = append(aggregated.fields, log.KV{K: k, V: fmt.Sprintf("%v", v)})
	}

	log.Info(ctx, aggregated) // Log at Info level with aggregated fields
}

// Error logs an error message with structured key-value pairs and an error object.
// It applies the specified format, converts data to key-value format, and logs the error.
func (l LogUtil) Error(ctx context.Context, kv interface{}, err error) {
	format := log.FormatTerminal // Default format is JSON
	if log.IsTerminal() {
		format = log.FormatTerminal // Use terminal format if outputting to a terminal
	}
	ctx = log.Context(ctx, log.WithDebug(), log.WithFormat(format))

	userKVMap := Data.StructToKVMap(kv) // Convert structure to key-value map
	aggregated := &aggregatedKVFielder{}

	// Aggregate fields for logging
	for k, v := range userKVMap {
		aggregated.fields = append(aggregated.fields, log.KV{K: k, V: fmt.Sprintf("%v", v)})
	}

	log.Error(ctx, err, aggregated) // Log the error with the aggregated fields
}

// Log is a global instance of LogUtil for convenient access to logging functions throughout the application.
var (
	Log LogUtil = newLogUtil()
)
