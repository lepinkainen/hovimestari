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
		// Extract specific information for special handling
		date := userInfo["Date"]
		location := userInfo["Location"]
		family := userInfo["Family"]
		weather := userInfo["Weather"]
		futureWeather := userInfo["FutureWeather"]
		weatherChanges := userInfo["WeatherChanges"]
		birthdays := userInfo["Birthdays"]

		promptBuilder.WriteString("Context Information:\n")

		if date != "" {
			promptBuilder.WriteString(fmt.Sprintf("- Current Date: %s\n", date))
		}

		if location != "" {
			promptBuilder.WriteString(fmt.Sprintf("- Location: %s\n", location))
		}

		if family != "" {
			promptBuilder.WriteString(fmt.Sprintf("- Family Members: %s\n", family))
		}

		if weather != "" {
			promptBuilder.WriteString(fmt.Sprintf("- Today's Weather: %s\n", weather))
		}

		if futureWeather != "" {
			promptBuilder.WriteString("- Upcoming Weather Forecasts:\n")
			forecasts := strings.Split(futureWeather, "\n")
			for _, forecast := range forecasts {
				promptBuilder.WriteString(fmt.Sprintf("  * %s\n", forecast))
			}
		}

		if weatherChanges != "" {
			promptBuilder.WriteString("- Weather Forecast Changes:\n")
			changes := strings.Split(weatherChanges, "\n")
			for _, change := range changes {
				promptBuilder.WriteString(fmt.Sprintf("  * %s\n", change))
			}
		}

		if birthdays != "" {
			promptBuilder.WriteString(fmt.Sprintf("- Birthdays Today: %s\n", birthdays))
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

	promptBuilder.WriteString("Please generate a concise, well-organized daily brief in Finnish. Use a formal, butler-like tone as if you were a professional butler addressing the family. Begin with a proper greeting that includes the current date. Include today's weather forecast near the beginning of the brief. If there are upcoming weather forecasts, include them in a separate section. If there are changes to weather forecasts compared to previous forecasts, mention these changes. If there are birthdays today, make sure to highlight them prominently with congratulations.\n\n")

	promptBuilder.WriteString("Organize the information in a clear, readable format with appropriate sections. If there are calendar events, list them chronologically. If there are tasks or reminders, prioritize them appropriately.\n\n")

	promptBuilder.WriteString("End the brief with a respectful closing remark, as a butler would.\n")

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
