package output

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
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

// escapeMarkdownV2 escapes special characters for Telegram's MarkdownV2 format
func escapeMarkdownV2(text string) string {
	// Characters that need to be escaped in MarkdownV2:
	// '_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!'
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	
	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

// Send sends the content to a Telegram chat with markdown formatting
func (o *TelegramOutputter) Send(ctx context.Context, content string) error {
	slog.Info("Sending message to Telegram", "chat_id", o.ChatID, "content_length", len(content))

	// Construct the Telegram Bot API URL
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", o.BotToken)

	// Escape the content for MarkdownV2 format
	escapedContent := escapeMarkdownV2(content)

	// Create the message payload
	payload := map[string]string{
		"chat_id":    o.ChatID,
		"text":       escapedContent,
		"parse_mode": "MarkdownV2",
	}

	// Marshal the payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Failed to marshal Telegram message", "error", err)
		return fmt.Errorf("failed to marshal Telegram message: %w", err)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("Failed to create Telegram API request", "error", err)
		return fmt.Errorf("failed to create Telegram API request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	slog.Debug("Sending HTTP request to Telegram API", "url", url)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to send Telegram API request", "error", err)
		return fmt.Errorf("failed to send Telegram API request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response
	slog.Info("Received response from Telegram API", "status_code", resp.StatusCode)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Read response body for error details
		body, readErr := io.ReadAll(resp.Body)
		if readErr == nil {
			slog.Error("Telegram API request failed",
				"status_code", resp.StatusCode,
				"response_body", string(body))
		} else {
			slog.Error("Telegram API request failed, couldn't read response body",
				"status_code", resp.StatusCode,
				"read_error", readErr)
		}
		return fmt.Errorf("Telegram API request failed with status code %d", resp.StatusCode)
	}

	slog.Info("Successfully sent message to Telegram", "chat_id", o.ChatID)
	return nil
}
