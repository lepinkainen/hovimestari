package commands

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lepinkainen/hovimestari/internal/config"
	electricityimporter "github.com/lepinkainen/hovimestari/internal/importer/electricityprice"
	"github.com/lepinkainen/hovimestari/internal/store"
)

// ImportElectricityPriceCmd defines the import electricity price command for Kong
type ImportElectricityPriceCmd struct{}

// Run executes the import electricity price command
func (cmd *ImportElectricityPriceCmd) Run() error {
	return runImportElectricityPrice(context.Background())
}

func runImportElectricityPrice(ctx context.Context) error {
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

	zone := cfg.EntsoeZone
	if zone == "" {
		zone = "10YFI-1--------U" // default to Finland
	}

	tz, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		slog.Warn("Failed to load timezone, using UTC", "timezone", cfg.Timezone, "error", err)
		tz = time.UTC
	}

	slog.Info("Importing electricity prices", "zone", zone)

	importer := electricityimporter.NewImporter(s, cfg.EntsoeAPIKey, zone, tz)
	if err := importer.Import(ctx); err != nil {
		return fmt.Errorf("failed to import electricity prices: %w", err)
	}

	slog.Info("Electricity prices imported successfully")
	return nil
}
