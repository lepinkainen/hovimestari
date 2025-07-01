package commands

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lepinkainen/hovimestari/internal/config"
	"github.com/lepinkainen/hovimestari/internal/llm"
)

// ListModelsCmd defines the list models command for Kong
type ListModelsCmd struct{}

// Run executes the list models command
func (cmd *ListModelsCmd) Run() error {
	return runListModels(context.Background())
}

// runListModels runs the list models command, querying the Gemini API for available
// models and displaying them to the user. It also shows the currently configured model
// from the configuration file.
func runListModels(ctx context.Context) error {
	// Get the configuration
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get configuration: %w", err)
	}

	// List the models
	slog.Info("Listing available Gemini models")
	models, err := llm.ListModels(ctx, cfg.GeminiAPIKey)
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	// Print the models
	fmt.Println("Available models:")
	for _, model := range models {
		fmt.Printf("- %s\n", model)
	}

	// Print the current model
	fmt.Printf("\nCurrent model configured: %s\n", cfg.GeminiModel)
	slog.Info("To change the model, edit the config.json file or set the HOVIMESTARI_GEMINI_MODEL environment variable")

	return nil
}
