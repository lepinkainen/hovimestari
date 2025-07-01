package commands

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lepinkainen/hovimestari/internal/config"
	"github.com/lepinkainen/hovimestari/internal/llm"
	"github.com/spf13/cobra"
)

// ListModelsCmd returns the list models command
func ListModelsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-models",
		Short: "List available Gemini models",
		Long:  `List all available Gemini models that can be used with the API.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListModels(cmd.Context())
		},
	}

	return cmd
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
