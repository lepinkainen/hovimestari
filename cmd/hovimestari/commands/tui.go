package commands

import (
	"context"
	"fmt"

	"github.com/lepinkainen/hovimestari/internal/config"
	"github.com/lepinkainen/hovimestari/internal/store"
	"github.com/lepinkainen/hovimestari/internal/tui"
)

// TUICmd defines the TUI command for Kong
type TUICmd struct{}

// Run executes the TUI command
func (cmd *TUICmd) Run() error {
	return runTUI(context.Background())
}

// runTUI starts the interactive terminal UI
func runTUI(ctx context.Context) error {
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
		if closeErr := store.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close store: %v\n", closeErr)
		}
	}()

	// Initialize the store
	if err := store.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}

	// Start the TUI application
	app := tui.NewApp(store, cfg)
	return app.Run()
}