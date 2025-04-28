package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
)

// Color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[37m"
)

// HumanReadableHandler is a slog.Handler that formats logs in a human-readable format
type HumanReadableHandler struct {
	out      io.Writer
	level    slog.Level
	mu       sync.Mutex
	useColor bool
}

// NewHumanReadableHandler creates a new HumanReadableHandler
func NewHumanReadableHandler(out io.Writer, opts *slog.HandlerOptions) *HumanReadableHandler {
	h := &HumanReadableHandler{
		out:      out,
		useColor: true, // Default to using color
	}

	if opts != nil {
		h.level = opts.Level.Level()
	}

	return h
}

// Enabled implements slog.Handler.
func (h *HumanReadableHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle implements slog.Handler.
func (h *HumanReadableHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Format time as YYYY-MM-DD HH:MM:SS
	timeStr := r.Time.Format("2006-01-02 15:04:05")

	// Format level with color
	levelStr := r.Level.String()
	if h.useColor {
		switch r.Level {
		case slog.LevelDebug:
			levelStr = colorGray + "DEBUG" + colorReset
		case slog.LevelInfo:
			levelStr = colorGreen + "INFO " + colorReset
		case slog.LevelWarn:
			levelStr = colorYellow + "WARN " + colorReset
		case slog.LevelError:
			levelStr = colorRed + "ERROR" + colorReset
		}
	}

	// Format the message
	msg := r.Message

	// Build the key-value pairs
	var kvPairs strings.Builder
	r.Attrs(func(a slog.Attr) bool {
		kvPairs.WriteString(fmt.Sprintf("  %s=%v", a.Key, a.Value.Any()))
		return true
	})

	// Write the formatted log line
	fmt.Fprintf(h.out, "%s  %s  %-40s%s\n", timeStr, levelStr, msg, kvPairs.String())

	return nil
}

// WithAttrs implements slog.Handler.
func (h *HumanReadableHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// For simplicity, we're not implementing attribute grouping
	// In a more complete implementation, we would create a new handler with the attributes
	return h
}

// WithGroup implements slog.Handler.
func (h *HumanReadableHandler) WithGroup(name string) slog.Handler {
	// For simplicity, we're not implementing attribute grouping
	return h
}
