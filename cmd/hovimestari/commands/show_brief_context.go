package commands

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shrike/hovimestari/internal/brief"
	"github.com/shrike/hovimestari/internal/config"
	"github.com/shrike/hovimestari/internal/llm"
	"github.com/shrike/hovimestari/internal/store"
	"github.com/spf13/cobra"
)

// ShowBriefContextCmd returns the show brief context command
func ShowBriefContextCmd() *cobra.Command {
	var daysAhead int

	cmd := &cobra.Command{
		Use:   "show-brief-context",
		Short: "Show the context given to the LLM for brief generation",
		Long:  `Show the full context that would be given to the LLM when generating a brief, without actually generating the brief.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShowBriefContext(cmd.Context(), daysAhead)
		},
	}

	// Add days-ahead flag specifically for brief context
	cmd.Flags().IntVar(&daysAhead, "days-ahead", 2, "Number of days ahead to include in the brief context")

	return cmd
}

// runShowBriefContext runs the show brief context command, building the same context
// that would be used for brief generation but displaying it to the user instead of
// sending it to the LLM. This is useful for debugging and understanding what information
// is included in the brief.
func runShowBriefContext(ctx context.Context, daysAhead int) error {
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

	// Load the prompts
	prompts, err := config.LoadPrompts(cfg.PromptFilePath)
	if err != nil {
		return fmt.Errorf("failed to load prompts: %w", err)
	}

	// Create the LLM client
	llmClient, err := llm.NewClient(cfg.GeminiAPIKey, cfg.GeminiModel, prompts)
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}
	defer func() {
		if err := llmClient.Close(); err != nil {
			slog.Error("Failed to close LLM client", "error", err)
		}
	}()

	// Create the brief generator
	generator := brief.NewGenerator(store, llmClient, cfg)

	// Build the brief context
	memoryStrings, userInfo, outputLanguage, err := generator.BuildBriefContext(ctx, daysAhead)
	if err != nil {
		return fmt.Errorf("failed to build brief context: %w", err)
	}

	// Build the prompt content
	promptContent := llmClient.BuildBriefPrompt(memoryStrings, userInfo, outputLanguage)

	// Print the prompt content
	fmt.Println("=== CONTEXT GIVEN TO LLM ===")
	fmt.Println(promptContent)
	fmt.Println("===========================")

	return nil
}
