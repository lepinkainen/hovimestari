package brief

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shrike/hovimestari/internal/config"
	"github.com/shrike/hovimestari/internal/llm"
	"github.com/shrike/hovimestari/internal/store"
	"github.com/shrike/hovimestari/internal/weather"
)

// Generator handles generating briefs based on memories
type Generator struct {
	store *store.Store
	llm   *llm.GeminiClient
	cfg   *config.Config
}

// NewGenerator creates a new brief generator
func NewGenerator(store *store.Store, llm *llm.GeminiClient, cfg *config.Config) *Generator {
	return &Generator{
		store: store,
		llm:   llm,
		cfg:   cfg,
	}
}

// GenerateDailyBrief generates a daily brief based on memories
func (g *Generator) GenerateDailyBrief(ctx context.Context, daysAhead int) (string, error) {
	// Get the date range for relevant memories
	loc, err := time.LoadLocation(g.cfg.Timezone)
	if err != nil {
		return "", fmt.Errorf("failed to load timezone: %w", err)
	}

	now := time.Now().In(loc)
	startDate := now
	endDate := startDate.AddDate(0, 0, daysAhead)

	// Get relevant memories
	memories, err := g.store.GetRelevantMemories(startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to get relevant memories: %w", err)
	}

	// Convert memories to strings
	var memoryStrings []string
	for _, memory := range memories {
		var dateInfo string
		if memory.RelevanceDate != nil {
			dateInfo = fmt.Sprintf(" (relevant on %s)", memory.RelevanceDate.Format("2006-01-02"))
		}
		memoryStrings = append(memoryStrings, fmt.Sprintf("%s%s [Source: %s]", memory.Content, dateInfo, memory.Source))
	}

	// Format the current date in Finnish
	weekdays := []string{"sunnuntai", "maanantai", "tiistai", "keskiviikko", "torstai", "perjantai", "lauantai"}
	months := []string{"tammikuuta", "helmikuuta", "maaliskuuta", "huhtikuuta", "toukokuuta", "kesäkuuta", "heinäkuuta", "elokuuta", "syyskuuta", "lokakuuta", "marraskuuta", "joulukuuta"}

	weekday := weekdays[now.Weekday()]
	day := now.Day()
	month := months[now.Month()-1]
	year := now.Year()

	formattedDate := fmt.Sprintf("%s, %d. %s %d", weekday, day, month, year)

	// Check for birthdays today
	var birthdaysToday []string
	for _, member := range g.cfg.Family {
		if member.Birthday != "" {
			birthday, err := time.Parse("2006-01-02", member.Birthday)
			if err != nil {
				continue // Skip invalid birthdays
			}

			// Check if today is their birthday (ignore year)
			if birthday.Month() == now.Month() && birthday.Day() == now.Day() {
				age := now.Year() - birthday.Year()
				birthdaysToday = append(birthdaysToday, fmt.Sprintf("%s (%d vuotta)", member.Name, age))
			}
		}
	}

	// Prepare family names
	var familyNames []string
	for _, member := range g.cfg.Family {
		familyNames = append(familyNames, member.Name)
	}

	// Get weather forecast
	weatherForecast, err := weather.GetForecast(g.cfg.Latitude, g.cfg.Longitude)
	if err != nil {
		// If weather fetching fails, just log the error and continue without weather
		weatherForecast = "Säätietoja ei saatavilla"
		fmt.Printf("Warning: Failed to fetch weather: %v\n", err)
	}

	// Add user information
	userInfo := map[string]string{
		"Date":     formattedDate,
		"Location": g.cfg.LocationName,
		"Family":   strings.Join(familyNames, ", "),
		"Weather":  weatherForecast,
	}

	// Add birthdays if any
	if len(birthdaysToday) > 0 {
		userInfo["Birthdays"] = strings.Join(birthdaysToday, ", ")
	}

	// Generate the brief
	brief, err := g.llm.GenerateBrief(ctx, memoryStrings, userInfo)
	if err != nil {
		return "", fmt.Errorf("failed to generate brief: %w", err)
	}

	return brief, nil
}

// GenerateResponse generates a response to a user query
func (g *Generator) GenerateResponse(ctx context.Context, query string) (string, error) {
	// Get all memories (we could be more selective here)
	startDate := time.Now().AddDate(-1, 0, 0) // Look back 1 year
	endDate := time.Now().AddDate(0, 1, 0)    // Look ahead 1 month

	// Get relevant memories
	memories, err := g.store.GetRelevantMemories(startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to get memories: %w", err)
	}

	// Convert memories to strings
	var memoryStrings []string
	for _, memory := range memories {
		var dateInfo string
		if memory.RelevanceDate != nil {
			dateInfo = fmt.Sprintf(" (relevant on %s)", memory.RelevanceDate.Format("2006-01-02"))
		}
		memoryStrings = append(memoryStrings, fmt.Sprintf("%s%s [Source: %s]", memory.Content, dateInfo, memory.Source))
	}

	// Generate the response
	response, err := g.llm.GenerateResponse(ctx, query, memoryStrings)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	return response, nil
}
