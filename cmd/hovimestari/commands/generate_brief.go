package commands

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lepinkainen/hovimestari/internal/brief"
	"github.com/lepinkainen/hovimestari/internal/config"
	"github.com/lepinkainen/hovimestari/internal/llm"
	"github.com/lepinkainen/hovimestari/internal/output"
	"github.com/lepinkainen/hovimestari/internal/store"
	"github.com/spf13/cobra"
)

// GenerateBriefCmd returns the generate brief command
func GenerateBriefCmd() *cobra.Command {
	var daysAheadFlag int

	cmd := &cobra.Command{
		Use:   "generate-brief",
		Short: "Generate a daily brief",
		Long:  `Generate a daily brief based on the stored memories.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the configuration
			cfg, err := config.GetConfig()
			if err != nil {
				return fmt.Errorf("failed to get configuration: %w", err)
			}

			// Use the flag value if provided, otherwise use the config value
			daysAhead := cfg.DaysAhead

			if cmd.Flags().Changed("days-ahead") {
				daysAhead = daysAheadFlag
			}

			// If neither flag nor config has a value, use the default
			if daysAhead == 0 {
				daysAhead = 2
			}

			return runGenerateBrief(cmd.Context(), daysAhead)
		},
	}

	// Add days-ahead flag as an override for the config value
	cmd.Flags().IntVar(&daysAheadFlag, "days-ahead", 0, "Number of days ahead to include in the brief (overrides config value)")

	return cmd
}

// runGenerateBrief runs the generate brief command, generating a daily brief based on
// memories stored in the database. It retrieves relevant memories for the current date
// and the specified number of days ahead, then uses the LLM to generate a natural language
// brief. The brief is then sent to all configured output channels (CLI, Discord, Telegram).
func runGenerateBrief(ctx context.Context, daysAhead int) error {
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

	// Generate the brief
	briefText, err := generator.GenerateDailyBrief(ctx, daysAhead)
	if err != nil {
		return fmt.Errorf("failed to generate brief: %w", err)
	}

	// Create a list of outputters based on the configuration
	var outputters []output.Outputter

	// Log configuration details
	slog.Debug("Configuration loaded", "output_format", cfg.OutputFormat)

	// Add CLI outputter if enabled
	if cfg.Outputs.EnableCLI || (cfg.OutputFormat == "cli" && len(cfg.Outputs.DiscordWebhookURLs) == 0 && len(cfg.Outputs.TelegramBots) == 0) {
		slog.Debug("Adding CLI outputter")
		outputters = append(outputters, output.NewCLIOutputter())
	}

	// Add Discord outputters
	for _, webhookURL := range cfg.Outputs.DiscordWebhookURLs {
		if webhookURL != "" {
			slog.Debug("Adding Discord outputter", "webhook_url_length", len(webhookURL))
			outputters = append(outputters, output.NewDiscordOutputter(webhookURL))
		}
	}

	// Add Telegram outputters
	for _, telegramCfg := range cfg.Outputs.TelegramBots {
		if telegramCfg.BotToken != "" && telegramCfg.ChatID != "" {
			slog.Debug("Adding Telegram outputter", "chat_id", telegramCfg.ChatID)
			outputters = append(outputters, output.NewTelegramOutputter(telegramCfg.BotToken, telegramCfg.ChatID))
		}
	}

	// If no outputters were configured, default to CLI
	if len(outputters) == 0 {
		slog.Debug("No outputters configured, defaulting to CLI")
		outputters = append(outputters, output.NewCLIOutputter())
	}

	slog.Debug("Total outputters configured", "count", len(outputters))

	// Send the brief to all configured outputters
	var outputErrors []error
	for _, outputter := range outputters {
		if err := outputter.Send(ctx, briefText); err != nil {
			outputErrors = append(outputErrors, err)
			slog.Error("Error sending brief", "error", err)
		}
	}

	// If all outputs failed, return an error
	if len(outputErrors) > 0 && len(outputErrors) == len(outputters) {
		return fmt.Errorf("all outputs failed: %v", outputErrors[0])
	}

	return nil
}
