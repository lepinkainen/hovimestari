package commands

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shrike/hovimestari/internal/config"
	"github.com/shrike/hovimestari/internal/importer/calendar"
	"github.com/shrike/hovimestari/internal/store"
	"github.com/spf13/cobra"
)

// ImportCalendarCmd returns the import calendar command
func ImportCalendarCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-calendar",
		Short: "Import calendar events",
		Long:  `Import all calendar events from the configured WebCal URLs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImportCalendar(cmd.Context())
		},
	}

	return cmd
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
	defer store.Close()

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
