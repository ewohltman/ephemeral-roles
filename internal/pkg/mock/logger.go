package mock

import (
	"log/slog"
)

// NewLogger provides a *slog.Logger that discards all output, for use in tests.
func NewLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}
