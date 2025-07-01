package commands

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/lepinkainen/hovimestari/internal/config"
	"github.com/lepinkainen/hovimestari/internal/xdg"
)

// InitConfigCmd defines the init config command for Kong
type InitConfigCmd struct {
	GeminiAPIKey string `kong:"help='Google Gemini API key',required"`
	OutputFormat string `kong:"help='Output format (cli, telegram)',default='cli'"`
	ConfigPath   string `kong:"help='Path to configuration file (default: $XDG_CONFIG_HOME/hovimestari/config.json)',name='config-path'"`
}

// Run executes the init config command
func (cmd *InitConfigCmd) Run() error {
	return runInitConfig(cmd.ConfigPath, cmd.GeminiAPIKey, cmd.OutputFormat)
}

// runInitConfig runs the init config command, creating a new configuration file with
// default values and the provided API key and output format. It sets up a basic configuration
// with example calendar and family member entries that the user can edit manually.
// The function prevents overwriting an existing configuration.
func runInitConfig(configPath, geminiAPIKey, outputFormat string) error {
	// Determine the target config path
	targetConfigPath := configPath
	if targetConfigPath == "" {
		// If no config path is provided, use the XDG config directory
		configDir, err := xdg.GetConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get XDG config directory: %w", err)
		}
		targetConfigPath = filepath.Join(configDir, "config.json")
	}

	// Check if the configuration file already exists
	if _, err := os.Stat(targetConfigPath); err == nil {
		return fmt.Errorf("configuration file '%s' already exists - please remove it first if you want to re-initialize", targetConfigPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check if configuration file exists: %w", err)
	}

	// Get the XDG config and data directories
	configDir, err := xdg.GetConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get XDG config directory: %w", err)
	}

	// Create a basic configuration
	cfg := &config.Config{
		// Leave DBPath and PromptFilePath empty to use the defaults
		DBPath:         "",
		GeminiAPIKey:   geminiAPIKey,
		GeminiModel:    "gemini-2.0-flash", // Default model
		OutputFormat:   outputFormat,
		OutputLanguage: "Finnish",
		PromptFilePath: "",
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
	configFileDir := filepath.Dir(targetConfigPath)
	if configFileDir != "" && configFileDir != "." {
		if err := os.MkdirAll(configFileDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Save the configuration
	file, err := os.Create(targetConfigPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Error("Failed to close config file", "error", err)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	// Create a default prompts.json file if it doesn't exist
	promptsPath := filepath.Join(configDir, "prompts.json")
	if _, err := os.Stat(promptsPath); os.IsNotExist(err) {
		// Copy the existing prompts.json file if it exists in the current directory
		if promptsData, err := os.ReadFile("prompts.json"); err == nil {
			if err := os.WriteFile(promptsPath, promptsData, 0644); err != nil {
				return fmt.Errorf("failed to create prompts.json: %w", err)
			}
			slog.Info("Created prompts.json", "path", promptsPath)
		} else {
			// If the file doesn't exist, create a basic one
			basicPrompts := map[string][]string{
				"dailyBrief": {
					"You are Hovimestari, a helpful butler assistant. Your task is to generate a daily brief in %LANG% for your user based on the following information:",
					"",
					"Context Information:",
					"%CONTEXT%",
					"",
					"Relevant Information:",
					"%NOTES%",
					"",
					"Please generate a concise, well-organized daily brief in %LANG%.",
				},
				"userQuery": {
					"You are Hovimestari, a helpful butler assistant. Your task is to respond to the user's query in %LANG% based on the following information:",
					"",
					"User Query: %QUERY%",
					"",
					"Relevant Information:",
					"%NOTES%",
					"",
					"Please respond in %LANG% using a formal, butler-like tone.",
				},
			}

			promptsFile, err := os.Create(promptsPath)
			if err != nil {
				return fmt.Errorf("failed to create prompts.json: %w", err)
			}
			defer func() {
				if err := promptsFile.Close(); err != nil {
					slog.Error("Failed to close prompts file", "error", err)
				}
			}()

			encoder := json.NewEncoder(promptsFile)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(basicPrompts); err != nil {
				return fmt.Errorf("failed to encode prompts: %w", err)
			}
			slog.Info("Created default prompts.json", "path", promptsPath)
		}
	}

	slog.Info("Settings saved to file", "path", targetConfigPath)
	slog.Info("NOTE: Edit the file manually to add the correct calendars, family members, and location information")
	slog.Info("The application will look for configuration files in the following locations:")
	slog.Info("1. The path specified with --config flag")
	slog.Info("2. $XDG_CONFIG_HOME/hovimestari/ (usually ~/.config/hovimestari/)")
	slog.Info("3. The directory containing the executable")
	return nil
}
