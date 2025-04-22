package output

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	// Create the message
	message := discordMessage{
		Content: content,
	}

	// Marshal the message to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord message: %w", err)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", o.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create Discord webhook request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Discord webhook request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Discord webhook request failed with status code %d", resp.StatusCode)
	}

	return nil
}
