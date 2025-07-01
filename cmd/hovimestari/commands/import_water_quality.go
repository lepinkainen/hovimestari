package commands

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lepinkainen/hovimestari/internal/config"
	"github.com/lepinkainen/hovimestari/internal/store"
)

// ImportWaterQualityCmd defines the import water quality command for Kong
type ImportWaterQualityCmd struct {
	Location string `kong:"help='The name of the measurement location',required"`
	Quality  string `kong:"help='The water quality status',required"`
}

// Run executes the import water quality command
func (cmd *ImportWaterQualityCmd) Run() error {
	return runImportWaterQuality(context.Background(), cmd.Location, cmd.Quality)
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