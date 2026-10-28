package llm

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

const (
	// PromptContextPlaceholder is the placeholder for context in prompts.
	PromptContextPlaceholder = "%CONTEXT%"
	// PromptNotesPlaceholder is the placeholder for notes/memories in prompts.
	PromptNotesPlaceholder = "%NOTES%"
	// PromptLanguagePlaceholder is the placeholder for the output language in prompts.
	PromptLanguagePlaceholder = "%LANG%"
	// PromptQueryPlaceholder is the placeholder for user queries in prompts.
	PromptQueryPlaceholder = "%QUERY%"
)

// cleanMarkdownWrapper removes markdown code block wrapping from LLM responses
func cleanMarkdownWrapper(content string) string {
	// Remove leading and trailing whitespace
	content = strings.TrimSpace(content)

	// Check if content starts with ```markdown and ends with ```
	if strings.HasPrefix(content, "```markdown") && strings.HasSuffix(content, "```") {
		// Remove the markdown code block wrapper
		content = strings.TrimPrefix(content, "```markdown")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	return content
}

// Client is a client for the Google Gemini API
type Client struct {
	client  *genai.Client
	model   string
	prompts map[string][]string
}

// newGenaiClient creates a Gemini API client authenticated with the given API key
func newGenaiClient(ctx context.Context, apiKey string) (*genai.Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	return client, nil
}

// NewClient creates a new Gemini client with the given API key, model name, and prompts
func NewClient(apiKey string, modelName string, prompts map[string][]string) (*Client, error) {
	client, err := newGenaiClient(context.Background(), apiKey)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:  client,
		model:   modelName,
		prompts: prompts,
	}, nil
}

// Close releases resources held by the client. The underlying SDK client holds
// no persistent connections, so this is a no-op kept for API compatibility.
func (c *Client) Close() error {
	return nil
}

// Generate generates content using the Gemini API with the specified prompt content and output language
func (c *Client) Generate(ctx context.Context, promptKey string, outputLanguage string, promptContent string) (string, error) {
	// For debugging purposes, you can print the prompt
	// fmt.Println("Prompt for Gemini:", promptContent)

	// Generate the response
	resp, err := c.client.Models.GenerateContent(ctx, c.model, genai.Text(promptContent), nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	// Extract the text from the response
	text := resp.Text()
	if text == "" {
		return "", fmt.Errorf("unexpected response format")
	}

	// Clean any markdown wrapper from the response
	cleanedText := cleanMarkdownWrapper(text)

	return cleanedText, nil
}

// BuildBriefPrompt builds the prompt content for a brief without sending it to the LLM
func (c *Client) BuildBriefPrompt(memories []string, userInfo map[string]string, outputLanguage string) string {
	// Build the context information
	var contextBuilder strings.Builder

	// Add user information if available
	if len(userInfo) > 0 {
		// Extract specific information for special handling
		date := userInfo["Date"]
		currentTime := userInfo["CurrentTime"]
		timezone := userInfo["Timezone"]
		location := userInfo["Location"]
		family := userInfo["Family"]
		weather := userInfo["Weather"]
		futureWeather := userInfo["FutureWeather"]
		weatherChanges := userInfo["WeatherChanges"]
		birthdays := userInfo["Birthdays"]
		ongoingEvents := userInfo["OngoingEvents"]

		if date != "" {
			fmt.Fprintf(&contextBuilder, "- Current Date: %s\n", date)
		}

		if currentTime != "" {
			fmt.Fprintf(&contextBuilder, "- Current Time: %s\n", currentTime)
		}

		if timezone != "" {
			fmt.Fprintf(&contextBuilder, "- Timezone: %s\n", timezone)
		}

		if location != "" {
			fmt.Fprintf(&contextBuilder, "- Location: %s\n", location)
		}

		if family != "" {
			fmt.Fprintf(&contextBuilder, "- Family Members: %s\n", family)
		}

		if weather != "" {
			fmt.Fprintf(&contextBuilder, "- Today's Weather: %s\n", weather)
		}

		if futureWeather != "" {
			contextBuilder.WriteString("- Upcoming Weather Forecasts:\n")
			for forecast := range strings.SplitSeq(futureWeather, "\n") {
				fmt.Fprintf(&contextBuilder, "  * %s\n", forecast)
			}
		}

		if weatherChanges != "" {
			contextBuilder.WriteString("- Weather Forecast Changes:\n")
			for change := range strings.SplitSeq(weatherChanges, "\n") {
				fmt.Fprintf(&contextBuilder, "  * %s\n", change)
			}
		}

		if birthdays != "" {
			fmt.Fprintf(&contextBuilder, "- Birthdays Today: %s\n", birthdays)
		}

		if ongoingEvents != "" {
			contextBuilder.WriteString("- Currently Ongoing:\n")
			for event := range strings.SplitSeq(ongoingEvents, "\n") {
				fmt.Fprintf(&contextBuilder, "  * %s\n", event)
			}
		}
	}

	// Format memories
	var memoryBuilder strings.Builder
	for _, memory := range memories {
		fmt.Fprintf(&memoryBuilder, "- %s\n", memory)
	}

	// Create the prompt content with context, memories, and language
	promptContent := strings.Join(c.prompts["dailyBrief"], "\n")
	promptContent = strings.ReplaceAll(promptContent, PromptContextPlaceholder, contextBuilder.String())
	promptContent = strings.ReplaceAll(promptContent, PromptNotesPlaceholder, memoryBuilder.String())
	promptContent = strings.ReplaceAll(promptContent, PromptLanguagePlaceholder, outputLanguage)

	return promptContent
}

// GenerateBrief generates a brief based on the provided memories
func (c *Client) GenerateBrief(ctx context.Context, memories []string, userInfo map[string]string, outputLanguage string) (string, error) {
	// Build the prompt content
	promptContent := c.BuildBriefPrompt(memories, userInfo, outputLanguage)

	// Generate the brief
	return c.Generate(ctx, "dailyBrief", outputLanguage, promptContent)
}

// GenerateResponse generates a response to a user query
func (c *Client) GenerateResponse(ctx context.Context, query string, memories []string, outputLanguage string) (string, error) {
	// Format memories
	var memoryBuilder strings.Builder
	for _, memory := range memories {
		fmt.Fprintf(&memoryBuilder, "- %s\n", memory)
	}

	// Create the prompt content with query, memories, and language
	promptContent := strings.Join(c.prompts["userQuery"], "\n")
	promptContent = strings.ReplaceAll(promptContent, PromptQueryPlaceholder, query)
	promptContent = strings.ReplaceAll(promptContent, PromptNotesPlaceholder, memoryBuilder.String())
	promptContent = strings.ReplaceAll(promptContent, PromptLanguagePlaceholder, outputLanguage)

	// Generate the response
	return c.Generate(ctx, "userQuery", outputLanguage, promptContent)
}

// ListModels lists the available Gemini models
func ListModels(ctx context.Context, apiKey string) ([]string, error) {
	// Create a temporary client
	client, err := newGenaiClient(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	// List the models, collecting names across all pages
	var modelNames []string
	for model, err := range client.Models.All(ctx) {
		if err != nil {
			return nil, fmt.Errorf("failed to list models: %w", err)
		}
		modelNames = append(modelNames, model.Name)
	}

	// If no models were found, return a clear error
	if len(modelNames) == 0 {
		return nil, fmt.Errorf("no models returned by the API - this may be due to API limitations, " +
			"permissions issues, or regional restrictions")
	}

	return modelNames, nil
}
