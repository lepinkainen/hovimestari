package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/shrike/hovimestari/internal/config"
	"github.com/shrike/hovimestari/internal/store"
	"github.com/spf13/cobra"
)

// AddMemoryCmd returns the add memory command
func AddMemoryCmd() *cobra.Command {
	var (
		content       string
		relevanceDate string
		source        string
	)

	cmd := &cobra.Command{
		Use:   "add-memory",
		Short: "Add a memory",
		Long:  `Add a memory to the database.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddMemory(cmd.Context(), content, relevanceDate, source)
		},
	}

	cmd.Flags().StringVar(&content, "content", "", "Memory content")
	cmd.Flags().StringVar(&relevanceDate, "relevance-date", "", "Relevance date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&source, "source", "manual", "Memory source")

	cmd.MarkFlagRequired("content")

	return cmd
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
	defer store.Close()

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

	fmt.Printf("Memory added successfully with ID %d.\n", id)
	return nil
}
