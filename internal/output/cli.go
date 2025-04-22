package output

import (
	"context"
	"fmt"
)

// CLIOutputter outputs content to the command line
type CLIOutputter struct{}

// NewCLIOutputter creates a new CLI outputter
func NewCLIOutputter() *CLIOutputter {
	return &CLIOutputter{}
}

// Send prints the content to the command line
func (o *CLIOutputter) Send(ctx context.Context, content string) error {
	fmt.Println(content)
	return nil
}
