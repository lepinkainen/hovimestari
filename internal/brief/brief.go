package brief

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shrike/hovimestari/internal/config"
	"github.com/shrike/hovimestari/internal/importer/calendar"
	weatherimporter "github.com/shrike/hovimestari/internal/importer/weather"
	"github.com/shrike/hovimestari/internal/llm"
	"github.com/shrike/hovimestari/internal/store"
	"github.com/shrike/hovimestari/internal/weather"
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

// getRelevantMemoryStrings fetches relevant memories and formats them as strings
func (g *Generator) getRelevantMemoryStrings(startDate, endDate time.Time) ([]string, []store.Memory, error) {
	// Get relevant memories
	memories, err := g.store.GetRelevantMemories(startDate, endDate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get relevant memories: %w", err)
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

	return memoryStrings, memories, nil
}

// findBirthdaysToday checks for family members' birthdays on the given date
func (g *Generator) findBirthdaysToday(now time.Time) []string {
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
	return birthdaysToday
}

// getFamilyNames returns a list of family member names
func (g *Generator) getFamilyNames() []string {
	var familyNames []string
	for _, member := range g.cfg.Family {
		familyNames = append(familyNames, member.Name)
	}
	return familyNames
}

// findOngoingCalendarEvents identifies calendar events that are currently ongoing
func (g *Generator) findOngoingCalendarEvents(memories []store.Memory, loc *time.Location, now time.Time) []string {
	var ongoingEvents []string
	for _, memory := range memories {
		// Check if this is a calendar event
		if strings.HasPrefix(memory.Source, calendar.CalendarSourcePrefix+":") && strings.HasPrefix(memory.Content, calendar.EventFormatPrefix) {
			// Parse the event content to extract start and end times
			content := memory.Content

			// Check if the event has a time range
			if strings.Contains(content, calendar.EventFormatFromSeparator) && strings.Contains(content, calendar.EventFormatToSeparator) {
				// Extract the start and end times
				fromIndex := strings.Index(content, calendar.EventFormatFromSeparator)
				toIndex := strings.Index(content, calendar.EventFormatToSeparator)

				if fromIndex > 0 && toIndex > fromIndex {
					// Extract the date-time strings
					timeStr := content[fromIndex+6 : toIndex]
					endTimeStr := content[toIndex+4:]

					// If end time contains " at ", truncate it
					if atIndex := strings.Index(endTimeStr, calendar.EventFormatAtSeparator); atIndex > 0 {
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
							summary := content[len(calendar.EventFormatPrefix):fromIndex]

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
	return ongoingEvents
}

// getWeatherData fetches weather forecasts and changes
func (g *Generator) getWeatherData(now, endDate time.Time, daysAhead int) (map[string]string, map[string]string, string, error) {
	// Get weather forecasts from memories
	weatherForecasts, err := weatherimporter.GetLatestForecasts(g.store, now, endDate, g.cfg.LocationName)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to get weather forecasts: %w", err)
	}

	// Check for forecast changes
	forecastChanges, err := weatherimporter.DetectForecastChanges(g.store, now, endDate, g.cfg.LocationName)
	if err != nil {
		return weatherForecasts, nil, "", fmt.Errorf("failed to detect forecast changes: %w", err)
	}

	// Get hourly forecast for today
	hourlyForecast, err := weather.GetCurrentDayHourlyForecast(g.cfg.Latitude, g.cfg.Longitude)
	if err != nil {
		return weatherForecasts, forecastChanges, "", fmt.Errorf("failed to get hourly forecast: %w", err)
	}

	return weatherForecasts, forecastChanges, hourlyForecast, nil
}

// assembleUserInfo creates the userInfo map with all relevant information
func (g *Generator) assembleUserInfo(
	now time.Time,
	daysAhead int,
	familyNames []string,
	ongoingEvents []string,
	birthdaysToday []string,
	weatherForecasts map[string]string,
	forecastChanges map[string]string,
	hourlyForecast string,
) map[string]string {
	// Format the current date and time in standard format (LLM will handle translation)
	formattedDate := now.Format("Monday, 2 January 2006")
	formattedTime := now.Format("15:04")

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

	// Add hourly forecast for today if available
	if hourlyForecast != "" {
		userInfo["HourlyForecastToday"] = hourlyForecast
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

	return userInfo
}

// BuildBriefContext builds the context for a daily brief without generating it
func (g *Generator) BuildBriefContext(ctx context.Context, daysAhead int) ([]string, map[string]string, string, error) {
	// Get the date range for relevant memories
	loc, err := time.LoadLocation(g.cfg.Timezone)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to load timezone: %w", err)
	}

	now := time.Now().In(loc)
	startDate := now
	endDate := startDate.AddDate(0, 0, daysAhead)

	// Get relevant memories and convert to strings
	memoryStrings, memories, err := g.getRelevantMemoryStrings(startDate, endDate)
	if err != nil {
		return nil, nil, "", err
	}

	// Find birthdays for today
	birthdaysToday := g.findBirthdaysToday(now)

	// Get family names
	familyNames := g.getFamilyNames()

	// Find ongoing calendar events
	ongoingEvents := g.findOngoingCalendarEvents(memories, loc, now)

	// Get weather data
	weatherForecasts, forecastChanges, hourlyForecast, err := g.getWeatherData(now, endDate, daysAhead)
	if err != nil {
		// Log the error but continue - weather data is non-critical
		fmt.Printf("Warning: %v\n", err)
	}

	// Assemble the user info map
	userInfo := g.assembleUserInfo(
		now,
		daysAhead,
		familyNames,
		ongoingEvents,
		birthdaysToday,
		weatherForecasts,
		forecastChanges,
		hourlyForecast,
	)

	// Get output language from config, default to Finnish if not specified
	outputLanguage := g.cfg.OutputLanguage
	if outputLanguage == "" {
		outputLanguage = "Finnish"
	}

	return memoryStrings, userInfo, outputLanguage, nil
}

// GenerateDailyBrief generates a daily brief based on memories
func (g *Generator) GenerateDailyBrief(ctx context.Context, daysAhead int) (string, error) {
	// Build the context
	memoryStrings, userInfo, outputLanguage, err := g.BuildBriefContext(ctx, daysAhead)
	if err != nil {
		return "", err
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
