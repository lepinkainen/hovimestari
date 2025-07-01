package commands

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lepinkainen/hovimestari/internal/config"
	weatherimporter "github.com/lepinkainen/hovimestari/internal/importer/weather"
	"github.com/lepinkainen/hovimestari/internal/store"
	"github.com/spf13/cobra"
)

// ImportWeatherCmd returns the import weather command
func ImportWeatherCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-weather",
		Short: "Import weather forecasts",
		Long:  `Import all available weather forecasts for the configured location and store them as memories.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImportWeather(cmd.Context())
		},
	}

	return cmd
}

// runImportWeather runs the import weather command, fetching weather forecasts for the
// configured location and storing them as memories in the database. Each forecast is
// stored with its relevance date set to the forecast date.
func runImportWeather(ctx context.Context) error {
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

	slog.Info("Importing weather forecasts", "location", cfg.LocationName)

	// Create the weather importer
	importer := weatherimporter.NewImporter(store, cfg.Latitude, cfg.Longitude, cfg.LocationName)

	// Import the weather forecasts
	if err := importer.Import(ctx); err != nil {
		return fmt.Errorf("failed to import weather forecasts: %w", err)
	}

	slog.Info("Weather forecasts imported successfully")
	return nil
}
