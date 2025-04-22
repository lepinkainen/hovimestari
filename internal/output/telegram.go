package output

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// TelegramOutputter sends content to a Telegram chat
type TelegramOutputter struct {
	BotToken string
	ChatID   string
}

// NewTelegramOutputter creates a new Telegram outputter
func NewTelegramOutputter(botToken, chatID string) *TelegramOutputter {
	return &TelegramOutputter{
		BotToken: botToken,
		ChatID:   chatID,
	}
}

// Send sends the content to a Telegram chat
func (o *TelegramOutputter) Send(ctx context.Context, content string) error {
	// Construct the Telegram Bot API URL
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", o.BotToken)

	// Create the message payload
	payload := map[string]string{
		"chat_id": o.ChatID,
		"text":    content,
	}

	// Marshal the payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Telegram message: %w", err)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create Telegram API request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Telegram API request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Telegram API request failed with status code %d", resp.StatusCode)
	}

	return nil
}
