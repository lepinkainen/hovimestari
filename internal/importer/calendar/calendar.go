package calendar

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/apognu/gocal"
	"github.com/lepinkainen/hovimestari/internal/store"
)

const (
	// CalendarSourcePrefix is the prefix used for calendar memory sources.
	CalendarSourcePrefix = "calendar"
)

// UpdateStrategy defines how calendar events should be updated
type UpdateStrategy string

const (
	// UpdateStrategyUpsert updates existing events and inserts new ones
	UpdateStrategyUpsert UpdateStrategy = "upsert"

	// UpdateStrategyReplaceAll deletes all events and reinserts them
	UpdateStrategyReplaceAll UpdateStrategy = "replace_all"
)

// mapUpdateModeToStrategy converts user-friendly update mode to internal strategy
func mapUpdateModeToStrategy(mode string) UpdateStrategy {
	switch mode {
	case "smart":
		return UpdateStrategyUpsert
	default:
		// Default to full_refresh for any other value or empty string
		return UpdateStrategyReplaceAll
	}
}

// Importer handles importing calendar events from a WebCal URL
type Importer struct {
	store          *store.Store
	webCalURL      string
	calendarName   string
	updateStrategy UpdateStrategy
}

// NewImporter creates a new calendar importer
func NewImporter(store *store.Store, webCalURL string, calendarName string, updateMode string) *Importer {
	// Convert webcal:// to https:// if needed
	url := webCalURL
	if strings.HasPrefix(url, "webcal://") {
		url = "https://" + url[9:]
	}

	return &Importer{
		store:          store,
		webCalURL:      url,
		calendarName:   calendarName,
		updateStrategy: mapUpdateModeToStrategy(updateMode),
	}
}

// Import fetches calendar events and stores them in the database
func (i *Importer) Import(ctx context.Context) error {
	// Fetch the iCalendar data
	resp, err := http.Get(i.webCalURL)
	if err != nil {
		return fmt.Errorf("failed to fetch calendar data: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("Failed to close response body", "error", err)
		}
	}()

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
		slog.Warn("Warning: Some events may have been skipped due to parsing errors: %v", "error", err)
	}

	// Log the number of events successfully parsed
	slog.Info("Successfully parsed calendar events", "count", len(parser.Events))

	// Create the source string with the calendar name
	source := fmt.Sprintf("%s:%s", CalendarSourcePrefix, i.calendarName)

	// If using replace_all strategy, delete all existing events for this calendar
	if i.updateStrategy == UpdateStrategyReplaceAll {
		err = i.store.DeleteCalendarEventsBySource(source)
		if err != nil {
			return fmt.Errorf("failed to delete existing calendar events: %w", err)
		}
		slog.Info("Deleted all existing events", "calendarname", i.calendarName)
	}

	// Process the events
	for _, event := range parser.Events {
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

		if i.updateStrategy == UpdateStrategyUpsert {
			// Check if this specific event instance already exists in the database
			exists, err := i.store.CalendarEventExists(source, event.Uid, *event.Start)
			if err != nil {
				return fmt.Errorf("failed to check if calendar event exists: %w", err)
			}

			if exists {
				// Update the existing event
				err = i.store.UpdateCalendarEvent(
					event.Uid,
					event.Summary,
					*event.Start,
					event.End,
					location,
					description,
					source,
				)
				if err != nil {
					return fmt.Errorf("failed to update calendar event in database: %w", err)
				}
				slog.Info("Updated calendar event: %s at %s", event.Summary, event.Start.Format("2006-01-02 15:04"))
				continue
			}
		}

		// Add the event to the database (for both strategies: new events in upsert mode, all events in replace_all mode)
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
