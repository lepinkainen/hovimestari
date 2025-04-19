package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Client is a client for the Google Gemini API
type Client struct {
	client  *genai.Client
	model   *genai.GenerativeModel
	prompts map[string][]string
}

// NewClient creates a new Gemini client with the given API key and prompts
func NewClient(apiKey string, prompts map[string][]string) (*Client, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Use Gemini Pro model
	model := client.GenerativeModel("gemini-2.0-flash")

	return &Client{
		client:  client,
		model:   model,
		prompts: prompts,
	}, nil
}

// Close closes the Gemini client
func (c *Client) Close() error {
	return c.client.Close()
}

// Generate generates content using the Gemini API with the specified prompt content and output language
func (c *Client) Generate(ctx context.Context, promptKey string, outputLanguage string, promptContent string) (string, error) {
	// For debugging purposes, you can print the prompt
	// fmt.Println("Prompt for Gemini:", promptContent)

	// Generate the response
	resp, err := c.model.GenerateContent(ctx, genai.Text(promptContent))
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

// GenerateBrief generates a brief based on the provided memories
func (c *Client) GenerateBrief(ctx context.Context, memories []string, userInfo map[string]string, outputLanguage string) (string, error) {
	// Build the context information
	var contextBuilder strings.Builder

	// Add user information if available
	if len(userInfo) > 0 {
		// Extract specific information for special handling
		date := userInfo["Date"]
		location := userInfo["Location"]
		family := userInfo["Family"]
		weather := userInfo["Weather"]
		futureWeather := userInfo["FutureWeather"]
		weatherChanges := userInfo["WeatherChanges"]
		birthdays := userInfo["Birthdays"]

		if date != "" {
			contextBuilder.WriteString(fmt.Sprintf("- Current Date: %s\n", date))
		}

		if location != "" {
			contextBuilder.WriteString(fmt.Sprintf("- Location: %s\n", location))
		}

		if family != "" {
			contextBuilder.WriteString(fmt.Sprintf("- Family Members: %s\n", family))
		}

		if weather != "" {
			contextBuilder.WriteString(fmt.Sprintf("- Today's Weather: %s\n", weather))
		}

		if futureWeather != "" {
			contextBuilder.WriteString("- Upcoming Weather Forecasts:\n")
			forecasts := strings.Split(futureWeather, "\n")
			for _, forecast := range forecasts {
				contextBuilder.WriteString(fmt.Sprintf("  * %s\n", forecast))
			}
		}

		if weatherChanges != "" {
			contextBuilder.WriteString("- Weather Forecast Changes:\n")
			changes := strings.Split(weatherChanges, "\n")
			for _, change := range changes {
				contextBuilder.WriteString(fmt.Sprintf("  * %s\n", change))
			}
		}

		if birthdays != "" {
			contextBuilder.WriteString(fmt.Sprintf("- Birthdays Today: %s\n", birthdays))
		}
	}

	// Format memories
	var memoryBuilder strings.Builder
	for _, memory := range memories {
		memoryBuilder.WriteString(fmt.Sprintf("- %s\n", memory))
	}

	// Create the prompt content with context, memories, and language
	promptContent := strings.Join(c.prompts["dailyBrief"], "\n")
	promptContent = strings.ReplaceAll(promptContent, "%CONTEXT%", contextBuilder.String())
	promptContent = strings.ReplaceAll(promptContent, "%NOTES%", memoryBuilder.String())
	promptContent = strings.ReplaceAll(promptContent, "%LANG%", outputLanguage)

	// Generate the brief
	return c.Generate(ctx, "dailyBrief", outputLanguage, promptContent)
}

// GenerateResponse generates a response to a user query
func (c *Client) GenerateResponse(ctx context.Context, query string, memories []string, outputLanguage string) (string, error) {
	// Format memories
	var memoryBuilder strings.Builder
	for _, memory := range memories {
		memoryBuilder.WriteString(fmt.Sprintf("- %s\n", memory))
	}

	// Create the prompt content with query, memories, and language
	promptContent := strings.Join(c.prompts["userQuery"], "\n")
	promptContent = strings.ReplaceAll(promptContent, "%QUERY%", query)
	promptContent = strings.ReplaceAll(promptContent, "%NOTES%", memoryBuilder.String())
	promptContent = strings.ReplaceAll(promptContent, "%LANG%", outputLanguage)

	// Generate the response
	return c.Generate(ctx, "userQuery", outputLanguage, promptContent)
}
