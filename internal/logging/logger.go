// Package logging provides a simple setup for the global logger used throughout the office tracker application.
// It initializes a text handler that writes log messages to standard output, with a default log level of INFO.
package logging

import (
	"log/slog"
	"os"
)

// Setup initializes the global logger with a text handler that writes to standard output.
func Setup() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// Component returns a logger with the "component" field set to the specified name, for easier log filtering.
func Component(name string) *slog.Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler).With("component", name)
}
