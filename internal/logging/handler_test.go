package logging

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestHumanReadableHandlerEnabled(t *testing.T) {
	h := NewHumanReadableHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: slog.LevelWarn})
	if h.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("Enabled(info) = true, want false")
	}
	if !h.Enabled(context.Background(), slog.LevelError) {
		t.Fatal("Enabled(error) = false, want true")
	}
}

func TestHumanReadableHandlerHandle(t *testing.T) {
	var buf bytes.Buffer
	h := NewHumanReadableHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	h.useColor = false

	rec := slog.NewRecord(time.Date(2026, 4, 1, 8, 9, 10, 0, time.UTC), slog.LevelInfo, "hello world", 0)
	rec.AddAttrs(slog.String("source", "test"), slog.Int("count", 2))

	if err := h.Handle(context.Background(), rec); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	got := buf.String()
	for _, want := range []string{"2026-04-01 08:09:10", "INFO", "hello world", "source=test", "count=2"} {
		if !strings.Contains(got, want) {
			t.Fatalf("Handle() output missing %q in %q", want, got)
		}
	}
}

func TestHumanReadableHandlerWithAttrsAndGroup(t *testing.T) {
	h := NewHumanReadableHandler(&bytes.Buffer{}, nil)
	if got := h.WithAttrs([]slog.Attr{slog.String("k", "v")}); got != h {
		t.Fatal("WithAttrs() did not return same handler")
	}
	if got := h.WithGroup("group"); got != h {
		t.Fatal("WithGroup() did not return same handler")
	}
}
