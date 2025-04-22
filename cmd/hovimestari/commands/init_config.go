package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shrike/hovimestari/internal/config"
	"github.com/spf13/cobra"
)

// InitConfigCmd returns the init config command
func InitConfigCmd() *cobra.Command {
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

// runInitConfig runs the init config command, creating a new configuration file with
// default values and the provided API key, database path, and output format. It sets
// up a basic configuration with example calendar and family member entries that the
// user can edit manually. The function prevents overwriting an existing configuration.
func runInitConfig(dbPath, geminiAPIKey, outputFormat string) error {
	// Check if the configuration file already exists
	if _, err := os.Stat(ConfigPath); err == nil {
		return fmt.Errorf("configuration file '%s' already exists - please remove it first if you want to re-initialize", ConfigPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check if configuration file exists: %w", err)
	}

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
	configDir := filepath.Dir(ConfigPath)
	if configDir != "" && configDir != "." {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Save the configuration
	if err := config.SaveConfig(cfg, ConfigPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Settings saved to file %s.\n", ConfigPath)
	fmt.Println("NOTE: Edit the file manually to add the correct calendars, family members, and location information.")
	return nil
}
