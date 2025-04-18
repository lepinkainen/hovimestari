package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiClient is a client for the Google Gemini API
type GeminiClient struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewGeminiClient creates a new Gemini client with the given API key
func NewGeminiClient(apiKey string) (*GeminiClient, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Use Gemini Pro model
	model := client.GenerativeModel("gemini-2.0-flash")

	return &GeminiClient{
		client: client,
		model:  model,
	}, nil
}

// Close closes the Gemini client
func (c *GeminiClient) Close() error {
	return c.client.Close()
}

// GenerateBrief generates a brief based on the provided memories
func (c *GeminiClient) GenerateBrief(ctx context.Context, memories []string, userInfo map[string]string) (string, error) {
	// Build the prompt
	var promptBuilder strings.Builder

	promptBuilder.WriteString("You are Hovimestari, a helpful butler assistant. Your task is to generate a daily brief in Finnish for your user based on the following information:\n\n")

	// Add user information if available
	if len(userInfo) > 0 {
		promptBuilder.WriteString("User Information:\n")
		for key, value := range userInfo {
			promptBuilder.WriteString(fmt.Sprintf("- %s: %s\n", key, value))
		}
		promptBuilder.WriteString("\n")
	}

	// Add memories
	if len(memories) > 0 {
		promptBuilder.WriteString("Relevant Information:\n")
		for _, memory := range memories {
			promptBuilder.WriteString(fmt.Sprintf("- %s\n", memory))
		}
		promptBuilder.WriteString("\n")
	}

	promptBuilder.WriteString("Please generate a concise, well-organized daily brief in Finnish. Use a formal, butler-like tone. Include only relevant information and organize it in a clear, readable format. If there are calendar events, list them chronologically. If there are tasks or reminders, prioritize them appropriately.\n")

	// Generate the response
	resp, err := c.model.GenerateContent(ctx, genai.Text(promptBuilder.String()))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	// Extract the text from the response
	text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	return string(text), nil
}

// GenerateResponse generates a response to a user query
func (c *GeminiClient) GenerateResponse(ctx context.Context, query string, memories []string) (string, error) {
	// Build the prompt
	var promptBuilder strings.Builder

	promptBuilder.WriteString("You are Hovimestari, a helpful butler assistant. Your task is to respond to the user's query in Finnish based on the following information:\n\n")

	// Add the user's query
	promptBuilder.WriteString(fmt.Sprintf("User Query: %s\n\n", query))

	// Add memories if available
	if len(memories) > 0 {
		promptBuilder.WriteString("Relevant Information:\n")
		for _, memory := range memories {
			promptBuilder.WriteString(fmt.Sprintf("- %s\n", memory))
		}
		promptBuilder.WriteString("\n")
	}

	promptBuilder.WriteString("Please respond in Finnish using a formal, butler-like tone. Be helpful, concise, and respectful. If you don't have enough information to answer the query, politely say so and ask for more details if necessary.\n")

	// Generate the response
	resp, err := c.model.GenerateContent(ctx, genai.Text(promptBuilder.String()))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	// Extract the text from the response
	text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	return string(text), nil
}
