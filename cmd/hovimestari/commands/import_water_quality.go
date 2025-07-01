package commands

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lepinkainen/hovimestari/internal/config"
	"github.com/lepinkainen/hovimestari/internal/store"
	"github.com/spf13/cobra"
)

// ImportWaterQualityCmd returns the import water quality command
func ImportWaterQualityCmd() *cobra.Command {
	var (
		location string
		quality  string
	)

	cmd := &cobra.Command{
		Use:   "import-water-quality",
		Short: "Import water quality data for a specific location",
		Long:  `Imports water quality data for a specific location and stores it as a memory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImportWaterQuality(cmd.Context(), location, quality)
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "The name of the measurement location (required)")
	cmd.Flags().StringVar(&quality, "quality", "", "The water quality status (required)")

	if err := cmd.MarkFlagRequired("location"); err != nil {
		return nil
	}
	if err := cmd.MarkFlagRequired("quality"); err != nil {
		return nil
	}

	return cmd
}

func runImportWaterQuality(ctx context.Context, location, quality string) error {
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get configuration: %w", err)
	}

	s, err := store.NewStore(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			slog.Error("Failed to close store", "error", err)
		}
	}()

	if err := s.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}

	memoryContent := fmt.Sprintf("Water quality at %s is %s.", location, quality)
	source := fmt.Sprintf("waterquality:%s", location)
	relevanceDate := time.Now()

	_, err = s.AddMemory(memoryContent, &relevanceDate, source, nil)
	if err != nil {
		return fmt.Errorf("failed to add memory: %w", err)
	}

	slog.Info("Water quality memory added successfully", "location", location)
	return nil
}