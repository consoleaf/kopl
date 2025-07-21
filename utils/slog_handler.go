package utils

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// simpleCLIHandler is a minimal slog.Handler for CLI output.
type simpleCLIHandler struct {
	outStdout io.Writer
	outStderr io.Writer
	mu        sync.Mutex // Protects writes to the underlying writers
	level     slog.Level
}

// NewSimpleCLIHandler creates a new simpleCLIHandler.
func NewSimpleCLIHandler() slog.Handler {
	_, exists := os.LookupEnv("DEBUG")
	level := slog.LevelInfo
	if exists {
		level = slog.LevelDebug
	}

	return &simpleCLIHandler{
		outStdout: os.Stdout,
		outStderr: os.Stderr,
		level:     level,
	}
}

// Enabled always returns true as we filter by output stream, not by dropping records.
func (h *simpleCLIHandler) Enabled(_ context.Context, level slog.Level) bool {
	return true
}

// Handle writes the log record's message and attributes to stdout (for info) or stderr (for others).
// It omits time and level. Attributes are simply appended.
func (h *simpleCLIHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var buf strings.Builder
	buf.WriteString(r.Message)

	// Append attributes simply as " key=value"
	r.Attrs(func(attr slog.Attr) bool {
		buf.WriteString(" ")
		buf.WriteString(attr.Key)
		buf.WriteString("=")
		buf.WriteString(attr.Value.String())
		return true
	})
	buf.WriteString("\n")

	if r.Level < h.level {
		return nil
	}

	output := h.outStderr
	if r.Level == slog.LevelInfo {
		output = h.outStdout
	}

	_, err := output.Write([]byte(buf.String()))
	return err
}

// WithAttrs returns a new handler. Attributes are handled on a per-record basis.
func (h *simpleCLIHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Attributes are processed directly in the Handle method per record.
	// We return a new handler instance as per slog.Handler contract.
	return &simpleCLIHandler{
		outStdout: h.outStdout,
		outStderr: h.outStderr,
	}
}

// WithGroup returns a new handler. Groups are not explicitly formatted in this simple version.
func (h *simpleCLIHandler) WithGroup(name string) slog.Handler {
	// For this extremely simple version, groups are ignored in formatting.
	// We still return a new handler to satisfy the slog.Handler contract.
	return &simpleCLIHandler{
		outStdout: h.outStdout,
		outStderr: h.outStderr,
	}
}
