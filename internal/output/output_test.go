package output

import (
	"context"
	"testing"
)

func TestCLIOutputter(t *testing.T) {
	outputter := NewCLIOutputter()
	err := outputter.Send(context.Background(), "Test message")
	if err != nil {
		t.Errorf("CLIOutputter.Send() returned an error: %v", err)
	}
}

func TestEscapeMarkdownV2(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "h1 heading becomes bold",
			input: "# Hello",
			want:  "**Hello**",
		},
		{
			name:  "h2 heading becomes bold",
			input: "## Hello",
			want:  "**Hello**",
		},
		{
			name:  "h3 heading becomes bold",
			input: "### Hello",
			want:  "**Hello**",
		},
		{
			name:  "heading in multiline text",
			input: "### Tämän päivän ohjelma\nsome text",
			want:  "**Tämän päivän ohjelma**\nsome text",
		},
		{
			name:  "bold preserved",
			input: "**bold text**",
			want:  "**bold text**",
		},
		{
			name:  "special chars escaped",
			input: "hello.world",
			want:  "hello\\.world",
		},
		{
			name:  "non-heading line unchanged",
			input: "just a line",
			want:  "just a line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeMarkdownV2(tt.input)
			if got != tt.want {
				t.Errorf("escapeMarkdownV2(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// Note: We don't test the Discord and Telegram outputters here because they require
// actual API calls. In a real-world scenario, we would mock the HTTP client to test
// these outputters without making actual API calls.
