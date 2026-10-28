package llm

import (
	"strings"
	"testing"
)

func TestCleanMarkdownWrapper(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "content wrapped in markdown block",
			input:    "```markdown\nHyvää huomenta! Tänään on keskiviikko.\n```",
			expected: "Hyvää huomenta! Tänään on keskiviikko.",
		},
		{
			name:     "content with extra whitespace",
			input:    "  ```markdown\n  Hyvää huomenta! Tänään on keskiviikko.\n  ```  ",
			expected: "Hyvää huomenta! Tänään on keskiviikko.",
		},
		{
			name:     "plain content without wrapper",
			input:    "Hyvää huomenta! Tänään on keskiviikko.",
			expected: "Hyvää huomenta! Tänään on keskiviikko.",
		},
		{
			name:     "content with internal markdown",
			input:    "```markdown\n**Hyvää huomenta!** Tänään on *keskiviikko*.\n```",
			expected: "**Hyvää huomenta!** Tänään on *keskiviikko*.",
		},
		{
			name:     "multiline content with markdown wrapper",
			input:    "```markdown\nHyvää huomenta! Tänään on keskiviikko, 13. elokuuta 2025.\n\nSään ennuste Järvenpäähän huomiselle:\n- Lämpötila: 12-25°C\n```",
			expected: "Hyvää huomenta! Tänään on keskiviikko, 13. elokuuta 2025.\n\nSään ennuste Järvenpäähän huomiselle:\n- Lämpötila: 12-25°C",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "only markdown wrapper",
			input:    "```markdown\n```",
			expected: "",
		},
		{
			name:     "partial markdown wrapper (start only)",
			input:    "```markdown\nHyvää huomenta!",
			expected: "```markdown\nHyvää huomenta!",
		},
		{
			name:     "partial markdown wrapper (end only)",
			input:    "Hyvää huomenta!\n```",
			expected: "Hyvää huomenta!\n```",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanMarkdownWrapper(tt.input)
			if result != tt.expected {
				t.Errorf("cleanMarkdownWrapper(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildBriefPromptIncludesDayType(t *testing.T) {
	c := &Client{prompts: map[string][]string{"dailyBrief": {"Context:", PromptContextPlaceholder}}}

	prompt := c.BuildBriefPrompt(nil, map[string]string{
		"Date":        "Friday, 1 May 2026",
		"DayType":     "Holiday",
		"HolidayName": "Vappu",
	}, "Finnish")

	for _, want := range []string{"- Day Type: Holiday\n", "- Holiday: Vappu\n"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt)
		}
	}

	prompt = c.BuildBriefPrompt(nil, map[string]string{"DayType": "Weekday"}, "Finnish")
	if strings.Contains(prompt, "- Holiday:") {
		t.Fatalf("prompt should not mention a holiday on a weekday:\n%s", prompt)
	}
}
