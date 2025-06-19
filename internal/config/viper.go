package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shrike/hovimestari/internal/xdg"
	"github.com/spf13/viper"
)

// CalendarConfig holds configuration for a calendar source
type CalendarConfig struct {
	Name       string `json:"name" mapstructure:"name"`
	URL        string `json:"url" mapstructure:"url"`
	UpdateMode string `json:"update_mode,omitempty" mapstructure:"update_mode"` // "smart" or "full_refresh"
}

// FamilyMember represents a family member with optional birthday and Telegram ID
type FamilyMember struct {
	Name       string `json:"name" mapstructure:"name"`
	Birthday   string `json:"birthday,omitempty" mapstructure:"birthday"` // Format: YYYY-MM-DD
	TelegramID string `json:"telegram_id,omitempty" mapstructure:"telegram_id"`
}

// TelegramConfig holds configuration for a Telegram bot
type TelegramConfig struct {
	BotToken string `json:"bot_token" mapstructure:"bot_token"`
	ChatID   string `json:"chat_id" mapstructure:"chat_id"`
}

// OutputConfig holds configuration for various output methods
type OutputConfig struct {
	EnableCLI          bool             `json:"enable_cli" mapstructure:"enable_cli"`
	DiscordWebhookURLs []string         `json:"discord_webhook_urls,omitempty" mapstructure:"discord_webhook_urls"`
	TelegramBots       []TelegramConfig `json:"telegram_bots,omitempty" mapstructure:"telegram_bots"`
}

// Config holds the application configuration
type Config struct {
	// Database configuration
	DBPath string `json:"db_path" mapstructure:"db_path"`

	// Logging configuration
	LogLevel string `json:"log_level,omitempty" mapstructure:"log_level"` // Log level (debug, info, warn, error)

	// LLM configuration
	GeminiAPIKey   string `json:"gemini_api_key" mapstructure:"gemini_api_key"`
	GeminiModel    string `json:"gemini_model,omitempty" mapstructure:"gemini_model"` // Gemini model to use (e.g., "gemini-2.0-flash")
	OutputLanguage string `json:"outputLanguage" mapstructure:"outputLanguage"`       // Language for LLM responses (e.g., "Finnish", "English")
	PromptFilePath string `json:"promptFilePath" mapstructure:"promptFilePath"`       // Path to the prompts.json file

	// Brief configuration
	DaysAhead int `json:"days_ahead,omitempty" mapstructure:"days_ahead"` // Number of days ahead to include in the brief

	// Location configuration
	LocationName string  `json:"location_name" mapstructure:"location_name"`
	Latitude     float64 `json:"latitude" mapstructure:"latitude"`
	Longitude    float64 `json:"longitude" mapstructure:"longitude"`
	Timezone     string  `json:"timezone" mapstructure:"timezone"`

	// Calendar configuration
	Calendars []CalendarConfig `json:"calendars" mapstructure:"calendars"`

	// Family configuration
	Family []FamilyMember `json:"family" mapstructure:"family"`

	// Output configuration
	OutputFormat string       `json:"output_format" mapstructure:"output_format"` // "cli", "telegram", etc. (legacy, use Outputs instead)
	Outputs      OutputConfig `json:"outputs,omitempty" mapstructure:"outputs"`
}

// validateRequiredFields validates that required configuration fields are present
func validateRequiredFields(config *Config) error {
	if config.GeminiAPIKey == "" {
		return fmt.Errorf("gemini API key is required")
	}
	return nil
}

// validateLocation validates the location configuration
func validateLocation(config *Config) error {
	if config.LocationName == "" {
		return fmt.Errorf("location_name is required")
	}

	if config.Latitude < -90 || config.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}

	if config.Longitude < -180 || config.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}

	// Validate timezone
	if config.Timezone == "" {
		return fmt.Errorf("timezone is required")
	}

	// Try to load the timezone to validate it
	_, err := time.LoadLocation(config.Timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone: %w", err)
	}

	return nil
}

// validateCalendars validates the calendar configurations
func validateCalendars(config *Config) error {
	if len(config.Calendars) == 0 {
		return fmt.Errorf("at least one calendar is required")
	}

	for i, cal := range config.Calendars {
		if cal.Name == "" {
			return fmt.Errorf("calendar %d is missing a name", i+1)
		}
		if cal.URL == "" {
			return fmt.Errorf("calendar %d (%s) is missing a URL", i+1, cal.Name)
		}
	}

	return nil
}

// validateFamily validates the family member configurations
func validateFamily(config *Config) error {
	if len(config.Family) == 0 {
		return fmt.Errorf("at least one family member is required")
	}

	for i, member := range config.Family {
		if member.Name == "" {
			return fmt.Errorf("family member %d is missing a name", i+1)
		}

		// Validate birthday format if provided
		if member.Birthday != "" {
			_, err := time.Parse("2006-01-02", member.Birthday)
			if err != nil {
				return fmt.Errorf("invalid birthday format for %s: %w", member.Name, err)
			}
		}
	}

	return nil
}

// InitViper initializes the Viper configuration system
// It sets up the search paths for configuration files and loads the configuration
// If configFileFlag is not empty, it will be used as the configuration file path
// Otherwise, it will search for config.json in the XDG config directory and executable directory
func InitViper(configFileFlag string) error {
	// Set default values for fields not expected to be in the config file initially
	viper.SetDefault("gemini_model", "gemini-2.0-flash")
	viper.SetDefault("output_language", "Finnish")
	viper.SetDefault("output_format", "cli")
	viper.SetDefault("days_ahead", 2)
	viper.SetDefault("log_level", "info")

	// Configure environment variable handling
	viper.SetEnvPrefix("HOVIMESTARI")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Bind environment variables to specific keys
	if err := viper.BindEnv("gemini_api_key", "HOVIMESTARI_GEMINI_API_KEY"); err != nil {
		slog.Warn("Failed to bind gemini_api_key environment variable", "error", err)
	}
	if err := viper.BindEnv("gemini_model", "HOVIMESTARI_GEMINI_MODEL"); err != nil {
		slog.Warn("Failed to bind gemini_model environment variable", "error", err)
	}
	if err := viper.BindEnv("output_format", "HOVIMESTARI_OUTPUT_FORMAT"); err != nil {
		slog.Warn("Failed to bind output_format environment variable", "error", err)
	}
	if err := viper.BindEnv("db_path", "HOVIMESTARI_DB_PATH"); err != nil {
		slog.Warn("Failed to bind db_path environment variable", "error", err)
	}
	if err := viper.BindEnv("log_level", "HOVIMESTARI_LOG_LEVEL"); err != nil {
		slog.Warn("Failed to bind log_level environment variable", "error", err)
	}

	// Set up key mappings for inconsistent casing in the config file
	// This maps the JSON keys to the struct field names
	viper.SetDefault("gemini_api_key", "")
	viper.SetDefault("output_language", "Finnish")
	viper.SetDefault("prompt_file_path", "")

	// Handle inconsistent key names in the config file
	viper.RegisterAlias("outputLanguage", "output_language")
	viper.RegisterAlias("promptFilePath", "prompt_file_path")

	// If configFileFlag is provided, use that specific file
	if configFileFlag != "" {
		viper.SetConfigFile(configFileFlag)
	} else {
		// Otherwise, set up the search paths
		viper.SetConfigName("config")
		viper.SetConfigType("json")

		// Add the XDG config directory as the highest priority search path
		configDir, err := xdg.GetConfigDir()
		if err == nil {
			viper.AddConfigPath(configDir)
		}

		// Add the executable directory as a fallback search path
		exeDir, err := xdg.GetExecutableDir()
		if err == nil {
			viper.AddConfigPath(exeDir)
		}
	}

	// Attempt to read the configuration
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, but this might be expected if using defaults/env vars
			// Log it informatively but don't treat it as a fatal error yet
			fmt.Fprintf(os.Stderr, "Warning: No configuration file found. Using defaults and environment variables.\n")
			fmt.Fprintf(os.Stderr, "Expected locations: $XDG_CONFIG_HOME/hovimestari/config.json or executable directory\n")
		} else {
			// Some other error occurred while reading the config file
			return fmt.Errorf("failed to read configuration file: %w", err)
		}
	} else {
		// Debug output - log the config file that was used
		slog.Debug("Using config file", "path", viper.ConfigFileUsed())

		// Debug output - log all keys in the config file
		slog.Debug("Available keys in config")
		for _, key := range viper.AllKeys() {
			slog.Debug("Config key", "key", key, "value", viper.Get(key))
		}
	}

	return nil
}

// LoadPrompts loads the prompts from the specified file
func LoadPrompts(filePath string) (map[string][]string, error) {
	// If no path is provided, use the default
	if filePath == "" {
		filePath = "prompts.json"
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open prompts file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Error("Failed to close prompts file", "error", err)
		}
	}()

	// Decode the JSON
	var prompts map[string][]string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&prompts); err != nil {
		return nil, fmt.Errorf("failed to decode prompts file: %w", err)
	}

	return prompts, nil
}

// GetConfig returns the configuration from Viper
// It unmarshals the Viper configuration into a Config struct and resolves file paths
func GetConfig() (*Config, error) {
	// Create an empty Config struct
	cfg := &Config{}

	// Configure Viper to use the JSON tags when unmarshaling
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Unmarshal the entire configuration at once
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Debug output
	slog.Debug("Loaded configuration", "config", cfg)

	// Get the XDG config directory
	configDir, _ := xdg.GetConfigDir()

	// Get the path of the config file Viper actually used
	foundConfigFile := viper.ConfigFileUsed()
	configFileDir := ""
	if foundConfigFile != "" {
		configFileDir = filepath.Dir(foundConfigFile)
	}

	// Determine the final DBPath
	if cfg.DBPath == "" {
		// If DBPath is empty, set it to the default
		cfg.DBPath = filepath.Join(configDir, "memories.db")
	} else if !filepath.IsAbs(cfg.DBPath) && configFileDir != "" {
		// If DBPath is relative and config was loaded from a file, resolve it relative to the config file
		cfg.DBPath = filepath.Join(configFileDir, cfg.DBPath)
	}

	// Determine the final PromptFilePath
	if cfg.PromptFilePath == "" {
		// If PromptFilePath is empty, set it to the default
		cfg.PromptFilePath = filepath.Join(configDir, "prompts.json")
	} else if !filepath.IsAbs(cfg.PromptFilePath) && configFileDir != "" {
		// If PromptFilePath is relative and config was loaded from a file, resolve it relative to the config file
		cfg.PromptFilePath = filepath.Join(configFileDir, cfg.PromptFilePath)
	}

	// Check if prompts.json exists
	if _, err := os.Stat(cfg.PromptFilePath); os.IsNotExist(err) {
		// Try to find prompts.json in the standard locations
		promptsPath, err := xdg.FindConfigFile("prompts.json", "")
		if err == nil {
			// Found prompts.json in a standard location, use that
			cfg.PromptFilePath = promptsPath
		} else {
			// prompts.json not found, return an error
			return nil, fmt.Errorf("prompts file not found at %s or in standard locations ($XDG_CONFIG_HOME/hovimestari/ or executable directory)", cfg.PromptFilePath)
		}
	}

	// Validate the configuration
	if err := validateRequiredFields(cfg); err != nil {
		configSource := "environment variables"
		if foundConfigFile != "" {
			configSource = fmt.Sprintf("configuration file '%s'", foundConfigFile)
		}
		return nil, fmt.Errorf("%w (from %s)", err, configSource)
	}

	if err := validateLocation(cfg); err != nil {
		return nil, err
	}

	if err := validateCalendars(cfg); err != nil {
		return nil, err
	}

	if err := validateFamily(cfg); err != nil {
		return nil, err
	}

	// Set default values for Outputs if not specified
	if !cfg.Outputs.EnableCLI && len(cfg.Outputs.DiscordWebhookURLs) == 0 && len(cfg.Outputs.TelegramBots) == 0 {
		// If no outputs are configured, use the legacy OutputFormat field
		if cfg.OutputFormat == "cli" || cfg.OutputFormat == "" {
			cfg.Outputs.EnableCLI = true
		}
	}

	return cfg, nil
}
