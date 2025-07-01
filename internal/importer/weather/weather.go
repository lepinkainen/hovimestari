package weather

import (
	"context"
	"fmt"
	"time"

	"github.com/lepinkainen/hovimestari/internal/store"
	"github.com/lepinkainen/hovimestari/internal/weather"
)

const (
	// SourcePrefix is the prefix used for weather memory sources
	SourcePrefix = "weather-metno"
)

// Importer handles importing weather forecasts
type Importer struct {
	store     *store.Store
	latitude  float64
	longitude float64
	location  string
}

// NewImporter creates a new weather importer
func NewImporter(store *store.Store, latitude, longitude float64, location string) *Importer {
	return &Importer{
		store:     store,
		latitude:  latitude,
		longitude: longitude,
		location:  location,
	}
}

// Import fetches weather forecasts and stores them in the database
func (i *Importer) Import(ctx context.Context) error {
	// Fetch all available forecasts
	forecasts, err := weather.GetMultiDayForecast(i.latitude, i.longitude)
	if err != nil {
		return fmt.Errorf("failed to fetch weather forecasts: %w", err)
	}

	// Process each day's forecast
	for _, forecast := range forecasts {
		// Format the forecast as a memory
		content := weather.FormatDailyForecast(forecast)

		// Use the forecast date as the relevance date
		relevanceDate := forecast.Date

		// Add the memory to the database with the weather source
		source := fmt.Sprintf("%s:%s", SourcePrefix, i.location)
		_, err := i.store.AddMemory(content, &relevanceDate, source, nil)
		if err != nil {
			return fmt.Errorf("failed to add weather forecast to database: %w", err)
		}
	}

	return nil
}

// GetLatestForecasts retrieves the latest weather forecasts for a date range
func GetLatestForecasts(s *store.Store, startDate, endDate time.Time, location string) (map[string]string, error) {
	// Get all weather memories for the date range
	// Adjust the start date to include the entire day in UTC
	utcStartDate := startDate.UTC().Truncate(24 * time.Hour)
	utcEndDate := endDate.UTC().Add(24 * time.Hour).Truncate(24 * time.Hour)

	memories, err := s.GetRelevantMemories(utcStartDate, utcEndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get weather memories: %w", err)
	}

	// Group forecasts by date, keeping only the most recent for each date
	latestForecasts := make(map[string]store.Memory)
	source := fmt.Sprintf("%s:%s", SourcePrefix, location)

	for _, memory := range memories {
		// Skip non-weather memories or memories for other locations
		if memory.Source != source || memory.RelevanceDate == nil {
			continue
		}

		// Get the date as a string (YYYY-MM-DD)
		dateStr := memory.RelevanceDate.Format("2006-01-02")

		// Check if we already have a forecast for this date
		existing, exists := latestForecasts[dateStr]
		if !exists || memory.CreatedAt.After(existing.CreatedAt) {
			// This is a newer forecast, replace the existing one
			latestForecasts[dateStr] = memory
		}
	}

	// Convert to a map of date -> forecast content
	result := make(map[string]string)
	for dateStr, memory := range latestForecasts {
		result[dateStr] = memory.Content
	}

	return result, nil
}
