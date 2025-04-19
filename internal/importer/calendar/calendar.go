package calendar

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time" // Required for time.Time handling

	"github.com/apognu/gocal"
	"github.com/shrike/hovimestari/internal/store"
)

// Importer handles importing calendar events from a WebCal URL
type Importer struct {
	store        *store.Store
	webCalURL    string
	calendarName string
}

// NewImporter creates a new calendar importer
func NewImporter(store *store.Store, webCalURL string, calendarName string) *Importer {
	// Convert webcal:// to https:// if needed
	url := webCalURL
	if strings.HasPrefix(url, "webcal://") {
		url = "https://" + url[9:]
	}

	return &Importer{
		store:        store,
		webCalURL:    url,
		calendarName: calendarName,
	}
}

// Import fetches calendar events and stores them in the database
func (i *Importer) Import(ctx context.Context) error {
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

	// Pre-filter the iCalendar data to remove events without DTSTAMP
	filteredData, err := filterValidEvents(string(body))
	if err != nil {
		return fmt.Errorf("failed to filter calendar data: %w", err)
	}

	// Log a summary instead of the entire filtered data to avoid log flooding
	log.Printf("Calendar data filtered successfully. Ready to parse.")

	// Parse the filtered iCalendar data
	// No date filtering - import all events
	parser := gocal.NewParser(strings.NewReader(filteredData))
	// Use time package explicitly to ensure import is kept
	_ = time.Now() // This line ensures the time package is used
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

		// Add the memory to the database with the calendar name in the source
		source := fmt.Sprintf("calendar:%s", i.calendarName)

		// Check if this specific event instance already exists in the database
		exists, err := i.store.MemoryExists(source, event.Uid, *event.Start)
		if err != nil {
			return fmt.Errorf("failed to check if calendar event exists: %w", err)
		}

		if exists {
			// Skip this event instance as it's already in the database
			log.Printf("Skipping duplicate calendar event: %s at %s", event.Summary, event.Start.Format("2006-01-02 15:04"))
			continue
		}

		// Add the event to the database
		_, err = i.store.AddMemory(content, relevanceDate, source, &event.Uid)
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

// filterValidEvents filters out iCalendar events that don't have a DTSTAMP field
func filterValidEvents(icsData string) (string, error) {
	// Split the iCalendar data into components
	// First, ensure we keep the VCALENDAR begin and end tags
	parts := strings.Split(icsData, "BEGIN:VEVENT")

	if len(parts) <= 1 {
		// No events found or invalid format
		return "", fmt.Errorf("no events found in iCalendar data or invalid format")
	}

	// The first part contains the header (BEGIN:VCALENDAR and other properties)
	header := parts[0]

	var filteredBuilder strings.Builder
	filteredBuilder.WriteString(header)

	// Count of valid and invalid events
	validCount := 0
	invalidCount := 0

	// Process each event part (skip the first part which is the header)
	for i := 1; i < len(parts); i++ {
		eventPart := parts[i]

		// Check if this event has a DTSTAMP field
		if strings.Contains(eventPart, "DTSTAMP:") {
			// This event has a DTSTAMP, include it in the filtered data
			filteredBuilder.WriteString("BEGIN:VEVENT")
			filteredBuilder.WriteString(eventPart)
			validCount++
		} else {
			// This event is missing DTSTAMP, log and skip it
			// Extract some identifying information if possible
			summary := "unknown"
			if summaryIdx := strings.Index(eventPart, "SUMMARY:"); summaryIdx != -1 {
				endIdx := strings.Index(eventPart[summaryIdx:], "\n")
				if endIdx != -1 {
					summary = eventPart[summaryIdx+8 : summaryIdx+endIdx]
				}
			}

			log.Printf("Skipping calendar event due to missing DTSTAMP. Event summary: %s", summary)
			invalidCount++
		}
	}

	// Log the filtering results
	log.Printf("Calendar filtering results: %d valid events, %d invalid events skipped", validCount, invalidCount)

	return filteredBuilder.String(), nil
}
