package output

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// DiscordOutputter sends content to a Discord webhook
type DiscordOutputter struct {
	WebhookURL string
}

// NewDiscordOutputter creates a new Discord outputter
func NewDiscordOutputter(webhookURL string) *DiscordOutputter {
	return &DiscordOutputter{
		WebhookURL: webhookURL,
	}
}

// discordMessage represents a Discord webhook message
type discordMessage struct {
	Content string `json:"content"`
}

// Send sends the content to a Discord webhook
func (o *DiscordOutputter) Send(ctx context.Context, content string) error {
	slog.Info("Sending message to Discord webhook", "content_length", len(content))

	// Create the message
	message := discordMessage{
		Content: content,
	}

	// Marshal the message to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		slog.Error("Failed to marshal Discord message", "error", err)
		return fmt.Errorf("failed to marshal Discord message: %w", err)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", o.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("Failed to create Discord webhook request", "error", err)
		return fmt.Errorf("failed to create Discord webhook request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	slog.Debug("Sending HTTP request to Discord webhook")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to send Discord webhook request", "error", err)
		return fmt.Errorf("failed to send Discord webhook request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("Failed to close response body", "error", err)
		}
	}()

	// Check the response
	slog.Info("Received response from Discord webhook", "status_code", resp.StatusCode)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Read response body for error details
		body, readErr := io.ReadAll(resp.Body)
		if readErr == nil {
			slog.Error("Discord webhook request failed",
				"status_code", resp.StatusCode,
				"response_body", string(body))
		} else {
			slog.Error("Discord webhook request failed, couldn't read response body",
				"status_code", resp.StatusCode,
				"read_error", readErr)
		}
		return fmt.Errorf("discord webhook request failed with status code %d", resp.StatusCode)
	}

	slog.Info("Successfully sent message to Discord webhook")
	return nil
}
