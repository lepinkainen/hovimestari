package commands

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lepinkainen/hovimestari/internal/config"
	schoollunchimporter "github.com/lepinkainen/hovimestari/internal/importer/schoollunch"
	"github.com/lepinkainen/hovimestari/internal/store"
)

// ImportSchoolLunchCmd defines the import school lunch command for Kong
type ImportSchoolLunchCmd struct{}

// Run executes the import school lunch command
func (cmd *ImportSchoolLunchCmd) Run() error {
	return runImportSchoolLunch(context.Background())
}

// runImportSchoolLunch runs the import school lunch command, fetching school lunch menus
// for the current week and storing them as memories in the database. Each day's menu is
// stored with its relevance date set to that day's date.
func runImportSchoolLunch(ctx context.Context) error {
	// Get the configuration
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get configuration: %w", err)
	}

	// Check if school lunch is configured
	if cfg.SchoolLunchName == "" {
		slog.Debug("School lunch not configured, skipping import")
		return nil
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

	slog.Info("Importing school lunch menus", "school", cfg.SchoolLunchName)

	// Create the school lunch importer
	importer := schoollunchimporter.NewImporter(store, cfg.SchoolLunchURL, cfg.SchoolLunchName)

	// Import the school lunch menus
	if err := importer.Import(ctx); err != nil {
		return fmt.Errorf("failed to import school lunch menus: %w", err)
	}

	slog.Info("School lunch menus imported successfully")
	return nil
}
