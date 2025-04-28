package calendar

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/apognu/gocal"
	"github.com/shrike/hovimestari/internal/store"
)

const (
	// CalendarSourcePrefix is the prefix used for calendar memory sources.
	CalendarSourcePrefix = "calendar"
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

	// Parse the iCalendar data directly without filtering
	// No date filtering - import all events
	parser := gocal.NewParser(strings.NewReader(string(body)))
	// Set strict mode to fail only events with errors, not the entire feed
	parser.Strict.Mode = gocal.StrictModeFailEvent
	err = parser.Parse()
	if err != nil {
		log.Printf("Warning: Some events may have been skipped due to parsing errors: %v", err)
	}

	// Log the number of events successfully parsed
	log.Printf("Successfully parsed %d calendar events", len(parser.Events))

	// Process the events
	for _, event := range parser.Events {
		// Create the source string with the calendar name
		source := fmt.Sprintf("%s:%s", CalendarSourcePrefix, i.calendarName)

		// Check if this specific event instance already exists in the database
		exists, err := i.store.CalendarEventExists(source, event.Uid, *event.Start)
		if err != nil {
			return fmt.Errorf("failed to check if calendar event exists: %w", err)
		}

		if exists {
			// Skip this event instance as it's already in the database
			log.Printf("Skipping duplicate calendar event: %s at %s", event.Summary, event.Start.Format("2006-01-02 15:04"))
			continue
		}

		// Prepare event data for storage
		var location *string
		if event.Location != "" {
			location = &event.Location
		}

		var description *string
		if event.Description != "" {
			// Truncate long descriptions
			desc := event.Description
			if len(desc) > 1000 {
				desc = desc[:997] + "..."
			}
			description = &desc
		}

		// Add the event to the database
		_, err = i.store.AddCalendarEvent(
			event.Uid,
			event.Summary,
			*event.Start,
			event.End,
			location,
			description,
			source,
		)
		if err != nil {
			return fmt.Errorf("failed to add calendar event to database: %w", err)
		}
	}

	return nil
}
