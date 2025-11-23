package logger

import (
	"context"
	"os"

	"github.com/charmbracelet/log"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const loggerKey contextKey = "logger"

// Setup creates and configures a new logger with TTY auto-detection.
// When running in a TTY (terminal), output is styled with colors.
// When running in systemd or redirected to a file, output is plain/structured.
func Setup(level log.Level) *log.Logger {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    false,
		ReportTimestamp: true,
		Level:           level,
	})

	// Auto-detect TTY for styling
	if !isTerminal(os.Stderr) {
		// Disable colors for systemd/pipes (structured output)
		logger.SetStyles(&log.Styles{})
	}

	return logger
}

// WithLogger adds a logger to the context.
func WithLogger(ctx context.Context, logger *log.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// LogFromCtx retrieves the logger from context.
// If no logger is found, returns a default logger.
func LogFromCtx(ctx context.Context) *log.Logger {
	if logger, ok := ctx.Value(loggerKey).(*log.Logger); ok {
		return logger
	}
	// Fallback to default logger if not in context
	return log.Default()
}

// isTerminal checks if the file descriptor is a terminal.
func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}
