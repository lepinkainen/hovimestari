package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/shrike/hovimestari/internal/brief"
	"github.com/shrike/hovimestari/internal/config"
	"github.com/shrike/hovimestari/internal/importer/calendar"
	weatherimporter "github.com/shrike/hovimestari/internal/importer/weather"
	"github.com/shrike/hovimestari/internal/llm"
	"github.com/shrike/hovimestari/internal/output"
	"github.com/shrike/hovimestari/internal/store"
	"github.com/spf13/cobra"

	// Import SQLite driver
	_ "modernc.org/sqlite"
)

var (
	configPath string
	daysAhead  int
)

func main() {
	// Define the root command
	rootCmd := &cobra.Command{
		Use:   "hovimestari",
		Short: "Hovimestari - A personal AI butler assistant",
		Long:  `Hovimestari is a personal AI butler assistant that provides daily briefs and responds to queries.`,
	}

	// Add global flags
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "config.json", "Path to the configuration file")

	// Add commands
	rootCmd.AddCommand(importCalendarCmd())
	rootCmd.AddCommand(importWeatherCmd())
	rootCmd.AddCommand(generateBriefCmd())
	rootCmd.AddCommand(showBriefContextCmd())
	rootCmd.AddCommand(addMemoryCmd())
	rootCmd.AddCommand(initConfigCmd())
	rootCmd.AddCommand(listModelsCmd())

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// importCalendarCmd returns the import calendar command
func importCalendarCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-calendar",
		Short: "Import calendar events",
		Long:  `Import all calendar events from the configured WebCal URLs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImportCalendar(cmd.Context())
		},
	}

	return cmd
}

// importWeatherCmd returns the import weather command
func importWeatherCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-weather",
		Short: "Import weather forecasts",
		Long:  `Import all available weather forecasts for the configured location and store them as memories.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImportWeather(cmd.Context())
		},
	}

	return cmd
}

// generateBriefCmd returns the generate brief command
func generateBriefCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-brief",
		Short: "Generate a daily brief",
		Long:  `Generate a daily brief based on the stored memories.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerateBrief(cmd.Context())
		},
	}

	// Add days-ahead flag specifically for brief generation
	cmd.Flags().IntVar(&daysAhead, "days-ahead", 2, "Number of days ahead to include in the brief")

	return cmd
}

// addMemoryCmd returns the add memory command
func addMemoryCmd() *cobra.Command {
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

// initConfigCmd returns the init config command
func initConfigCmd() *cobra.Command {
	var (
		dbPath       string
		geminiAPIKey string
		outputFormat string
	)

	cmd := &cobra.Command{
		Use:   "init-config",
		Short: "Initialize the configuration",
		Long:  `Initialize the configuration file with the provided values. Note that this only sets up the basic configuration. You will need to edit the config.json file manually to add calendars, family members, and location information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInitConfig(dbPath, geminiAPIKey, outputFormat)
		},
	}

	cmd.Flags().StringVar(&dbPath, "db-path", "memories.db", "Path to the database file")
	cmd.Flags().StringVar(&geminiAPIKey, "gemini-api-key", "", "Google Gemini API key")
	cmd.Flags().StringVar(&outputFormat, "output-format", "cli", "Output format (cli, telegram)")

	cmd.MarkFlagRequired("gemini-api-key")

	return cmd
}

// runImportCalendar runs the import calendar command
func runImportCalendar(ctx context.Context) error {
	// Load the configuration
	cfg, err := config.LoadConfig(configPath)
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

	// Import events from each calendar
	for _, cal := range cfg.Calendars {
		fmt.Printf("Importing calendar events from calendar '%s'...\n", cal.Name)

		// Create the calendar importer
		importer := calendar.NewImporter(store, cal.URL, cal.Name)

		// Import the calendar events
		if err := importer.Import(ctx); err != nil {
			return fmt.Errorf("failed to import calendar events from '%s': %w", cal.Name, err)
		}
	}

	fmt.Println("Calendar events imported successfully.")
	return nil
}

// runImportWeather runs the import weather command
func runImportWeather(ctx context.Context) error {
	// Load the configuration
	cfg, err := config.LoadConfig(configPath)
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

	fmt.Printf("Importing weather forecasts for location '%s'...\n", cfg.LocationName)

	// Create the weather importer
	importer := weatherimporter.NewImporter(store, cfg.Latitude, cfg.Longitude, cfg.LocationName)

	// Import the weather forecasts
	if err := importer.Import(ctx); err != nil {
		return fmt.Errorf("failed to import weather forecasts: %w", err)
	}

	fmt.Println("Weather forecasts imported successfully.")
	return nil
}

// runGenerateBrief runs the generate brief command
func runGenerateBrief(ctx context.Context) error {
	// Load the configuration
	cfg, err := config.LoadConfig(configPath)
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

// runAddMemory runs the add memory command
func runAddMemory(ctx context.Context, content, relevanceDateStr, source string) error {
	// Load the configuration
	cfg, err := config.LoadConfig(configPath)
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

// listModelsCmd returns the list models command
func listModelsCmd() *cobra.Command {
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

// runListModels runs the list models command
func runListModels(ctx context.Context) error {
	// Load the configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// List the models
	fmt.Println("Listing available Gemini models...")
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
	fmt.Println("\nTo change the model, edit the config.json file or set the HOVIMESTARI_GEMINI_MODEL environment variable.")

	return nil
}

// showBriefContextCmd returns the show brief context command
func showBriefContextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-brief-context",
		Short: "Show the context given to the LLM for brief generation",
		Long:  `Show the full context that would be given to the LLM when generating a brief, without actually generating the brief.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShowBriefContext(cmd.Context())
		},
	}

	// Add days-ahead flag specifically for brief context
	cmd.Flags().IntVar(&daysAhead, "days-ahead", 2, "Number of days ahead to include in the brief context")

	return cmd
}

// runShowBriefContext runs the show brief context command
func runShowBriefContext(ctx context.Context) error {
	// Load the configuration
	cfg, err := config.LoadConfig(configPath)
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

// runInitConfig runs the init config command
func runInitConfig(dbPath, geminiAPIKey, outputFormat string) error {
	// Create a basic configuration
	cfg := &config.Config{
		DBPath:         dbPath,
		GeminiAPIKey:   geminiAPIKey,
		GeminiModel:    "gemini-2.0-flash", // Default model
		OutputFormat:   outputFormat,
		OutputLanguage: "Finnish",
		PromptFilePath: "prompts.json",
		LocationName:   "Helsinki",
		Latitude:       60.1699,
		Longitude:      24.9384,
		Timezone:       "Europe/Helsinki",
		Calendars: []config.CalendarConfig{
			{
				Name: "Example Calendar",
				URL:  "webcal://example.com/calendar.ics",
			},
		},
		Family: []config.FamilyMember{
			{
				Name:     "Example Person",
				Birthday: "2000-01-01",
			},
		},
		Outputs: config.OutputConfig{
			EnableCLI: outputFormat == "cli" || outputFormat == "",
		},
	}

	// Ensure the config directory exists
	configDir := filepath.Dir(configPath)
	if configDir != "" && configDir != "." {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Save the configuration
	if err := config.SaveConfig(cfg, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Settings saved to file %s.\n", configPath)
	fmt.Println("NOTE: Edit the file manually to add the correct calendars, family members, and location information.")
	return nil
}
