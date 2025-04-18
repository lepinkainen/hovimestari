package calendar

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/apognu/gocal"
	"github.com/shrike/hovimestari/internal/store"
)

// Importer handles importing calendar events from a WebCal URL
type Importer struct {
	store     *store.Store
	webCalURL string
}

// NewImporter creates a new calendar importer
func NewImporter(store *store.Store, webCalURL string) *Importer {
	// Convert webcal:// to https:// if needed
	url := webCalURL
	if strings.HasPrefix(url, "webcal://") {
		url = "https://" + url[9:]
	}

	return &Importer{
		store:     store,
		webCalURL: url,
	}
}

// Import fetches calendar events and stores them in the database
func (i *Importer) Import(ctx context.Context, daysAhead int) error {
	// Fetch the iCalendar data
	resp, err := http.Get(i.webCalURL)
	if err != nil {
		return fmt.Errorf("failed to fetch calendar data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch calendar data: status code %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read calendar data: %w", err)
	}

	// Parse the iCalendar data
	start := time.Now()
	end := start.AddDate(0, 0, daysAhead)

	parser := gocal.NewParser(strings.NewReader(string(body)))
	parser.Start = &start
	parser.End = &end
	err = parser.Parse()
	if err != nil {
		return fmt.Errorf("failed to parse calendar data: %w", err)
	}

	// Process the events
	for _, event := range parser.Events {
		// Format the event as a memory
		content := formatEvent(&event)

		// Use the event start time as the relevance date
		relevanceDate := event.Start

		// Add the memory to the database
		_, err := i.store.AddMemory(content, relevanceDate, "calendar")
		if err != nil {
			return fmt.Errorf("failed to add calendar event to database: %w", err)
		}
	}

	return nil
}

// formatEvent formats a calendar event as a memory string
func formatEvent(event *gocal.Event) string {
	var builder strings.Builder

	// Add the event summary
	builder.WriteString(fmt.Sprintf("Calendar Event: %s", event.Summary))

	// Add the event time
	if !event.Start.IsZero() {
		startTime := event.Start.Format("2006-01-02 15:04")

		if !event.End.IsZero() {
			endTime := event.End.Format("15:04")
			if event.Start.Day() != event.End.Day() {
				endTime = event.End.Format("2006-01-02 15:04")
			}
			builder.WriteString(fmt.Sprintf(" from %s to %s", startTime, endTime))
		} else {
			builder.WriteString(fmt.Sprintf(" at %s", startTime))
		}
	}

	// Add the location if available
	if event.Location != "" {
		builder.WriteString(fmt.Sprintf(" at %s", event.Location))
	}

	// Add the description if available
	if event.Description != "" {
		// Truncate long descriptions
		description := event.Description
		if len(description) > 200 {
			description = description[:197] + "..."
		}
		builder.WriteString(fmt.Sprintf(". Description: %s", description))
	}

	return builder.String()
}
