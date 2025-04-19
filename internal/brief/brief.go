package brief

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shrike/hovimestari/internal/config"
	weatherimporter "github.com/shrike/hovimestari/internal/importer/weather"
	"github.com/shrike/hovimestari/internal/llm"
	"github.com/shrike/hovimestari/internal/store"
)

// Generator handles generating briefs based on memories
type Generator struct {
	store *store.Store
	llm   *llm.Client
	cfg   *config.Config
}

// NewGenerator creates a new brief generator
func NewGenerator(store *store.Store, llm *llm.Client, cfg *config.Config) *Generator {
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

	// Format the current date and time in standard format (LLM will handle translation)
	formattedDate := now.Format("Monday, 2 January 2006")
	formattedTime := now.Format("15:04")

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
				birthdaysToday = append(birthdaysToday, fmt.Sprintf("%s (%d years)", member.Name, age))
			}
		}
	}

	// Prepare family names
	var familyNames []string
	for _, member := range g.cfg.Family {
		familyNames = append(familyNames, member.Name)
	}

	// Get weather forecasts from memories
	// Use the already defined now and endDate variables
	weatherForecasts, err := weatherimporter.GetLatestForecasts(g.store, now, endDate, g.cfg.LocationName)
	if err != nil {
		fmt.Printf("Warning: Failed to get weather forecasts: %v\n", err)
	}

	// Check for forecast changes
	forecastChanges, err := weatherimporter.DetectForecastChanges(g.store, now, endDate, g.cfg.LocationName)
	if err != nil {
		fmt.Printf("Warning: Failed to detect forecast changes: %v\n", err)
	}

	// Find ongoing calendar events
	var ongoingEvents []string
	for _, memory := range memories {
		// Check if this is a calendar event
		if strings.HasPrefix(memory.Source, "calendar:") && strings.HasPrefix(memory.Content, "Calendar Event:") {
			// Parse the event content to extract start and end times
			content := memory.Content

			// Check if the event has a time range
			if strings.Contains(content, " from ") && strings.Contains(content, " to ") {
				// Extract the start and end times
				fromIndex := strings.Index(content, " from ")
				toIndex := strings.Index(content, " to ")

				if fromIndex > 0 && toIndex > fromIndex {
					// Extract the date-time strings
					timeStr := content[fromIndex+6 : toIndex]
					endTimeStr := content[toIndex+4:]

					// If end time contains " at ", truncate it
					if atIndex := strings.Index(endTimeStr, " at "); atIndex > 0 {
						endTimeStr = endTimeStr[:atIndex]
					} else if dotIndex := strings.Index(endTimeStr, "."); dotIndex > 0 {
						// If end time contains ".", truncate it
						endTimeStr = endTimeStr[:dotIndex]
					}

					// Parse the start time
					var startTime time.Time
					var endTime time.Time
					var err error

					// Try to parse with different formats
					startTime, err = time.ParseInLocation("2006-01-02 15:04", timeStr, loc)
					if err == nil {
						// Check if end time is just a time (not a full date)
						if !strings.Contains(endTimeStr, "-") {
							// End time is just HH:MM, use the same date as start
							endTime, err = time.ParseInLocation("15:04", endTimeStr, loc)
							if err == nil {
								// Combine the start date with the end time
								endTime = time.Date(
									startTime.Year(), startTime.Month(), startTime.Day(),
									endTime.Hour(), endTime.Minute(), 0, 0, loc,
								)
							}
						} else {
							// End time includes a date
							endTime, err = time.ParseInLocation("2006-01-02 15:04", endTimeStr, loc)
						}

						// Check if the event is ongoing
						if err == nil && now.After(startTime) && now.Before(endTime) {
							// Extract the event summary
							summary := content[len("Calendar Event: "):fromIndex]

							// Format the ongoing event
							ongoingEvent := fmt.Sprintf("%s (until %s)",
								summary,
								endTime.Format("15:04"),
							)
							ongoingEvents = append(ongoingEvents, ongoingEvent)
						}
					}
				}
			}
		}
	}

	// Add user information
	userInfo := map[string]string{
		"Date":        formattedDate,
		"CurrentTime": formattedTime,
		"Timezone":    g.cfg.Timezone,
		"Location":    g.cfg.LocationName,
		"Family":      strings.Join(familyNames, ", "),
	}

	// Add ongoing events if any
	if len(ongoingEvents) > 0 {
		userInfo["OngoingEvents"] = strings.Join(ongoingEvents, "\n")
	}

	// Add today's weather if available
	todayStr := now.Format("2006-01-02")
	if forecast, ok := weatherForecasts[todayStr]; ok {
		userInfo["Weather"] = forecast
	} else {
		userInfo["Weather"] = "Weather information not available"
	}

	// Add future weather forecasts if available
	var futureWeather []string
	for i := 1; i <= daysAhead; i++ {
		futureDate := now.AddDate(0, 0, i)
		dateStr := futureDate.Format("2006-01-02")
		if forecast, ok := weatherForecasts[dateStr]; ok {
			futureWeather = append(futureWeather, forecast)
		}
	}

	if len(futureWeather) > 0 {
		userInfo["FutureWeather"] = strings.Join(futureWeather, "\n")
	}

	// Add forecast changes if any
	if len(forecastChanges) > 0 {
		var changes []string
		for _, change := range forecastChanges {
			changes = append(changes, change)
		}
		userInfo["WeatherChanges"] = strings.Join(changes, "\n")
	}

	// Add birthdays if any
	if len(birthdaysToday) > 0 {
		userInfo["Birthdays"] = strings.Join(birthdaysToday, ", ")
	}

	// Get output language from config, default to Finnish if not specified
	outputLanguage := g.cfg.OutputLanguage
	if outputLanguage == "" {
		outputLanguage = "Finnish"
	}

	// Generate the brief
	brief, err := g.llm.GenerateBrief(ctx, memoryStrings, userInfo, outputLanguage)
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

	// Get output language from config, default to Finnish if not specified
	outputLanguage := g.cfg.OutputLanguage
	if outputLanguage == "" {
		outputLanguage = "Finnish"
	}

	// Generate the response
	response, err := g.llm.GenerateResponse(ctx, query, memoryStrings, outputLanguage)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	return response, nil
}
