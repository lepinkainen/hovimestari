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
	Name       string `json:"name"`
	URL        string `json:"url"`
	UpdateMode string `json:"update_mode,omitempty"` // "smart" or "full_refresh"
}

// FamilyMember represents a family member with optional birthday and Telegram ID
type FamilyMember struct {
	Name       string `json:"name"`
	Birthday   string `json:"birthday,omitempty"` // Format: YYYY-MM-DD
	TelegramID string `json:"telegram_id,omitempty"`
}

// TelegramConfig holds configuration for a Telegram bot
type TelegramConfig struct {
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
}

// OutputConfig holds configuration for various output methods
type OutputConfig struct {
	EnableCLI          bool             `json:"enable_cli"`
	DiscordWebhookURLs []string         `json:"discord_webhook_urls,omitempty"`
	TelegramBots       []TelegramConfig `json:"telegram_bots,omitempty"`
}

// Config holds the application configuration
type Config struct {
	// Database configuration
	DBPath string `json:"db_path"`

	// LLM configuration
	GeminiAPIKey   string `json:"gemini_api_key"`
	GeminiModel    string `json:"gemini_model,omitempty"` // Gemini model to use (e.g., "gemini-2.0-flash")
	OutputLanguage string `json:"outputLanguage"`         // Language for LLM responses (e.g., "Finnish", "English")
	PromptFilePath string `json:"promptFilePath"`         // Path to the prompts.json file

	// Brief configuration
	DaysAhead int `json:"days_ahead,omitempty"` // Number of days ahead to include in the brief

	// Location configuration
	LocationName string  `json:"location_name"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Timezone     string  `json:"timezone"`

	// Calendar configuration
	Calendars []CalendarConfig `json:"calendars"`

	// Family configuration
	Family []FamilyMember `json:"family"`

	// Output configuration
	OutputFormat string       `json:"output_format"` // "cli", "telegram", etc. (legacy, use Outputs instead)
	Outputs      OutputConfig `json:"outputs,omitempty"`
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

	// Configure environment variable handling
	viper.SetEnvPrefix("HOVIMESTARI")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Bind environment variables to specific keys
	viper.BindEnv("gemini_api_key", "HOVIMESTARI_GEMINI_API_KEY")
	viper.BindEnv("gemini_model", "HOVIMESTARI_GEMINI_MODEL")
	viper.BindEnv("output_format", "HOVIMESTARI_OUTPUT_FORMAT")
	viper.BindEnv("db_path", "HOVIMESTARI_DB_PATH")

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
	defer file.Close()

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

	// Explicitly map Viper keys to struct fields
	cfg.GeminiAPIKey = viper.GetString("gemini_api_key")
	cfg.GeminiModel = viper.GetString("gemini_model")
	cfg.OutputLanguage = viper.GetString("output_language")
	cfg.PromptFilePath = viper.GetString("prompt_file_path")
	cfg.DBPath = viper.GetString("db_path")
	cfg.DaysAhead = viper.GetInt("days_ahead")
	cfg.LocationName = viper.GetString("location_name")
	cfg.Latitude = viper.GetFloat64("latitude")
	cfg.Longitude = viper.GetFloat64("longitude")
	cfg.Timezone = viper.GetString("timezone")
	cfg.OutputFormat = viper.GetString("output_format")

	// For complex types, we need to use Unmarshal
	if err := viper.UnmarshalKey("calendars", &cfg.Calendars); err != nil {
		return nil, fmt.Errorf("failed to unmarshal calendars: %w", err)
	}

	if err := viper.UnmarshalKey("family", &cfg.Family); err != nil {
		return nil, fmt.Errorf("failed to unmarshal family: %w", err)
	}

	// Unmarshal the outputs configuration
	if err := viper.UnmarshalKey("outputs", &cfg.Outputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal outputs: %w", err)
	}

	// Explicitly unmarshal the Discord webhook URLs and Telegram bots
	// This is a workaround for a potential issue with Viper not correctly unmarshaling nested fields
	if viper.IsSet("outputs.discord_webhook_urls") {
		webhookURLs := viper.Get("outputs.discord_webhook_urls")
		if urls, ok := webhookURLs.([]interface{}); ok {
			cfg.Outputs.DiscordWebhookURLs = make([]string, len(urls))
			for i, url := range urls {
				if strURL, ok := url.(string); ok {
					cfg.Outputs.DiscordWebhookURLs[i] = strURL
				}
			}
			slog.Debug("Explicitly unmarshaled Discord webhook URLs", "count", len(cfg.Outputs.DiscordWebhookURLs))
		}
	}

	if viper.IsSet("outputs.telegram_bots") {
		telegramBots := viper.Get("outputs.telegram_bots")
		if bots, ok := telegramBots.([]interface{}); ok {
			cfg.Outputs.TelegramBots = make([]TelegramConfig, len(bots))
			for i, bot := range bots {
				if botMap, ok := bot.(map[string]interface{}); ok {
					if botToken, ok := botMap["bot_token"].(string); ok {
						cfg.Outputs.TelegramBots[i].BotToken = botToken
					}
					if chatID, ok := botMap["chat_id"].(string); ok {
						cfg.Outputs.TelegramBots[i].ChatID = chatID
					}
				}
			}
			slog.Debug("Explicitly unmarshaled Telegram bots", "count", len(cfg.Outputs.TelegramBots))
		}
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
