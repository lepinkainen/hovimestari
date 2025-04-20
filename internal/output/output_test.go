package output

import (
	"context"
	"testing"
)

func TestCLIOutputter(t *testing.T) {
	// Create a new CLI outputter
	outputter := NewCLIOutputter()

	// Test the Send method (this just tests that it doesn't return an error)
	err := outputter.Send(context.Background(), "Test message")
	if err != nil {
		t.Errorf("CLIOutputter.Send() returned an error: %v", err)
	}
}

// Note: We don't test the Discord and Telegram outputters here because they require
// actual API calls. In a real-world scenario, we would mock the HTTP client to test
// these outputters without making actual API calls.
