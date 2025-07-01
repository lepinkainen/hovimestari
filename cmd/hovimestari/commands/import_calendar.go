package commands

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lepinkainen/hovimestari/internal/config"
	"github.com/lepinkainen/hovimestari/internal/importer/calendar"
	"github.com/lepinkainen/hovimestari/internal/store"
)

// ImportCalendarCmd defines the import calendar command for Kong
type ImportCalendarCmd struct{}

// Run executes the import calendar command
func (cmd *ImportCalendarCmd) Run() error {
	return runImportCalendar(context.Background())
}

// runImportCalendar runs the import calendar command, fetching events from all configured
// WebCal URLs and storing them as memories in the database. Each event is stored with
// its relevance date set to the event's start time.
func runImportCalendar(ctx context.Context) error {
	// Get the configuration
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get configuration: %w", err)
	}

	// Create the store
	store, err := store.NewStore(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			slog.Error("Failed to close store", "error", err)
		}
	}()

	// Initialize the store
	if err := store.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}

	// Import events from each calendar
	for _, cal := range cfg.Calendars {
		slog.Info("Importing calendar events", "calendar", cal.Name, "update_mode", cal.UpdateMode)

		// Create the calendar importer
		importer := calendar.NewImporter(store, cal.URL, cal.Name, cal.UpdateMode)

		// Import the calendar events
		if err := importer.Import(ctx); err != nil {
			return fmt.Errorf("failed to import calendar events from '%s': %w", cal.Name, err)
		}
	}

	slog.Info("Calendar events imported successfully")
	return nil
}
