package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/shrike/hovimestari/internal/brief"
	"github.com/shrike/hovimestari/internal/config"
	"github.com/shrike/hovimestari/internal/llm"
	"github.com/shrike/hovimestari/internal/output"
	"github.com/shrike/hovimestari/internal/store"
	"github.com/spf13/cobra"
)

// GenerateBriefCmd returns the generate brief command
func GenerateBriefCmd() *cobra.Command {
	var daysAhead int

	cmd := &cobra.Command{
		Use:   "generate-brief",
		Short: "Generate a daily brief",
		Long:  `Generate a daily brief based on the stored memories.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerateBrief(cmd.Context(), daysAhead)
		},
	}

	// Add days-ahead flag specifically for brief generation
	cmd.Flags().IntVar(&daysAhead, "days-ahead", 2, "Number of days ahead to include in the brief")

	return cmd
}

// runGenerateBrief runs the generate brief command, generating a daily brief based on
// memories stored in the database. It retrieves relevant memories for the current date
// and the specified number of days ahead, then uses the LLM to generate a natural language
// brief. The brief is then sent to all configured output channels (CLI, Discord, Telegram).
func runGenerateBrief(ctx context.Context, daysAhead int) error {
	// Load the configuration
	cfg, err := config.LoadConfig(ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
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

	// Determine the prompt file path
	promptFilePath := cfg.PromptFilePath
	if promptFilePath == "" {
		promptFilePath = "prompts.json"
	}

	// Load the prompts
	prompts, err := config.LoadPrompts(promptFilePath)
	if err != nil {
		return fmt.Errorf("failed to load prompts: %w", err)
	}

	// Create the LLM client
	llmClient, err := llm.NewClient(cfg.GeminiAPIKey, cfg.GeminiModel, prompts)
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}
	defer llmClient.Close()

	// Create the brief generator
	generator := brief.NewGenerator(store, llmClient, cfg)

	// Generate the brief
	briefText, err := generator.GenerateDailyBrief(ctx, daysAhead)
	if err != nil {
		return fmt.Errorf("failed to generate brief: %w", err)
	}

	// Create a list of outputters based on the configuration
	var outputters []output.Outputter

	// Add CLI outputter if enabled
	if cfg.Outputs.EnableCLI || (cfg.OutputFormat == "cli" && len(cfg.Outputs.DiscordWebhookURLs) == 0 && len(cfg.Outputs.TelegramBots) == 0) {
		outputters = append(outputters, output.NewCLIOutputter())
	}

	// Add Discord outputters
	for _, webhookURL := range cfg.Outputs.DiscordWebhookURLs {
		if webhookURL != "" {
			outputters = append(outputters, output.NewDiscordOutputter(webhookURL))
		}
	}

	// Add Telegram outputters
	for _, telegramCfg := range cfg.Outputs.TelegramBots {
		if telegramCfg.BotToken != "" && telegramCfg.ChatID != "" {
			outputters = append(outputters, output.NewTelegramOutputter(telegramCfg.BotToken, telegramCfg.ChatID))
		}
	}

	// If no outputters were configured, default to CLI
	if len(outputters) == 0 {
		outputters = append(outputters, output.NewCLIOutputter())
	}

	// Send the brief to all configured outputters
	var outputErrors []error
	for _, outputter := range outputters {
		if err := outputter.Send(ctx, briefText); err != nil {
			outputErrors = append(outputErrors, err)
			fmt.Fprintf(os.Stderr, "Error sending brief: %v\n", err)
		}
	}

	// If all outputs failed, return an error
	if len(outputErrors) > 0 && len(outputErrors) == len(outputters) {
		return fmt.Errorf("all outputs failed: %v", outputErrors[0])
	}

	return nil
}
