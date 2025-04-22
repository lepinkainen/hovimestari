package output

import (
	"context"
)

// Outputter is an interface for sending brief content to various destinations
type Outputter interface {
	// Send sends the content to the destination
	Send(ctx context.Context, content string) error
}
