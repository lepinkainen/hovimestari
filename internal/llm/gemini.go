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

	promptBuilder.WriteString("Please generate a concise, well-organized daily brief in Finnish. Use a formal, respectful, butler-like tone throughout.\n\n")

	promptBuilder.WriteString("**Instructions:**\n\n")

	promptBuilder.WriteString("1. **Structure:** Structure the brief as follows:\n")
	promptBuilder.WriteString("   * Greeting & Date (e.g., \"Hyvää huomenta, arvoisat asukkaat! Tänään on [Date].\")\n")
	promptBuilder.WriteString("   * Today's Summary: Include today's weather first, followed by today's calendar events chronologically.\n")
	promptBuilder.WriteString("   * Upcoming Days: Group information by day (e.g., \"Sunnuntai, 20. huhtikuuta:\"). For each day, list the weather forecast, followed by any relevant calendar events chronologically.\n")
	promptBuilder.WriteString("   * Closing (e.g., \"Kunnioittavasti, Hovimestarinne.\")\n\n")

	promptBuilder.WriteString("2. **Weather:**\n")
	promptBuilder.WriteString("   * For today's weather, mention the conditions and temperature range. Mention wind speed *only if* it exceeds 5 m/s.\n")
	promptBuilder.WriteString("   * For upcoming days' weather, mention only the conditions and temperature range.\n")
	promptBuilder.WriteString("   * If there are changes to weather forecasts compared to previous forecasts, mention these changes in the relevant day's section.\n\n")

	promptBuilder.WriteString("3. **Events:**\n")
	promptBuilder.WriteString("   * List calendar events chronologically within each day.\n")
	promptBuilder.WriteString("   * For Sanni's school events, list only the time and subject abbreviation (e.g., \"09:30 KS\"). Clearly group these under a heading like \"Sannin lukujärjestys:\" or similar for the relevant days.\n")
	promptBuilder.WriteString("   * Mention who events pertain to when relevant (e.g., \"Riku on Helsingissä.\").\n")
	promptBuilder.WriteString("   * Simplify event details where necessary for clarity, focusing on the essential information (what, when, where if applicable).\n\n")

	promptBuilder.WriteString("4. **Tone & Language:** Maintain a formal, helpful butler persona. Use Finnish. Use emojis sparingly and appropriately (e.g., for weather).\n\n")

	promptBuilder.WriteString("5. **Birthdays:** If any family members have a birthday today, highlight it prominently with congratulations near the beginning of the brief.\n\n")

	// For debugging purposes, you can print the prompt
	//fmt.Println("Prompt for Gemini:", promptBuilder.String())

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
