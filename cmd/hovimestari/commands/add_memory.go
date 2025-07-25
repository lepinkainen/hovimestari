package commands

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lepinkainen/hovimestari/internal/config"
	"github.com/lepinkainen/hovimestari/internal/store"
)

// AddMemoryCmd defines the add memory command for Kong
type AddMemoryCmd struct {
	Content       string `kong:"help='Memory content',required"`
	RelevanceDate string `kong:"help='Relevance date (YYYY-MM-DD)'"`
	Source        string `kong:"help='Memory source',default='manual'"`
}

// Run executes the add memory command
func (cmd *AddMemoryCmd) Run() error {
	return runAddMemory(context.Background(), cmd.Content, cmd.RelevanceDate, cmd.Source)
}

// runAddMemory runs the add memory command, adding a new memory to the database with
// the specified content, relevance date, and source. The relevance date is optional
// and can be provided in YYYY-MM-DD format. If not provided, the memory will be
// considered relevant for all dates.
func runAddMemory(ctx context.Context, content, relevanceDateStr, source string) error {
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

	// Parse the relevance date if provided
	var relevanceDate *time.Time
	if relevanceDateStr != "" {
		date, err := time.Parse("2006-01-02", relevanceDateStr)
		if err != nil {
			return fmt.Errorf("failed to parse relevance date: %w", err)
		}
		relevanceDate = &date
	}

	// Add the memory
	id, err := store.AddMemory(content, relevanceDate, source, nil)
	if err != nil {
		return fmt.Errorf("failed to add memory: %w", err)
	}

	slog.Info("Memory added successfully", "id", id)
	return nil
}
