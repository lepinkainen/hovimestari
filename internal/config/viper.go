package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shrike/hovimestari/internal/xdg"
	"github.com/spf13/viper"
)

// InitViper initializes the Viper configuration system
// It sets up the search paths for configuration files and loads the configuration
// If configFileFlag is not empty, it will be used as the configuration file path
// Otherwise, it will search for config.json in the XDG config directory and executable directory
func InitViper(configFileFlag string) error {
	// Set default values for fields not expected to be in the config file initially
	viper.SetDefault("gemini_model", "gemini-2.0-flash")
	viper.SetDefault("output_language", "Finnish")
	viper.SetDefault("output_format", "cli")

	// Configure environment variable handling
	viper.SetEnvPrefix("HOVIMESTARI")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set up key aliases for inconsistent casing in the config file
	viper.RegisterAlias("gemini_api_key", "geminiAPIKey")
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
	}

	return nil
}

// GetConfig returns the configuration from Viper
// It unmarshals the Viper configuration into a Config struct and resolves file paths
func GetConfig() (*Config, error) {
	// Create an empty Config struct
	cfg := &Config{}

	// Try to read values directly from the file
	if configFile := viper.ConfigFileUsed(); configFile != "" {
		data, err := os.ReadFile(configFile)
		if err == nil {
			content := string(data)

			// Extract API key
			apiKeyPrefix := `"gemini_api_key": "`
			if idx := strings.Index(content, apiKeyPrefix); idx >= 0 {
				start := idx + len(apiKeyPrefix)
				end := strings.Index(content[start:], `"`)
				if end > 0 {
					apiKey := content[start : start+end]
					cfg.GeminiAPIKey = apiKey
				}
			}

			// Extract location name
			locationPrefix := `"location_name": "`
			if idx := strings.Index(content, locationPrefix); idx >= 0 {
				start := idx + len(locationPrefix)
				end := strings.Index(content[start:], `"`)
				if end > 0 {
					location := content[start : start+end]
					cfg.LocationName = location
				}
			}

			// Extract latitude
			latitudePrefix := `"latitude": `
			if idx := strings.Index(content, latitudePrefix); idx >= 0 {
				start := idx + len(latitudePrefix)
				end := strings.IndexAny(content[start:], ",\n}")
				if end > 0 {
					latitudeStr := content[start : start+end]
					if latitude, err := strconv.ParseFloat(latitudeStr, 64); err == nil {
						cfg.Latitude = latitude
					}
				}
			}

			// Extract longitude
			longitudePrefix := `"longitude": `
			if idx := strings.Index(content, longitudePrefix); idx >= 0 {
				start := idx + len(longitudePrefix)
				end := strings.IndexAny(content[start:], ",\n}")
				if end > 0 {
					longitudeStr := content[start : start+end]
					if longitude, err := strconv.ParseFloat(longitudeStr, 64); err == nil {
						cfg.Longitude = longitude
					}
				}
			}

			// Extract timezone
			timezonePrefix := `"timezone": "`
			if idx := strings.Index(content, timezonePrefix); idx >= 0 {
				start := idx + len(timezonePrefix)
				end := strings.Index(content[start:], `"`)
				if end > 0 {
					timezone := content[start : start+end]
					cfg.Timezone = timezone
				}
			}

			// Extract Gemini model
			modelPrefix := `"gemini_model": "`
			if idx := strings.Index(content, modelPrefix); idx >= 0 {
				start := idx + len(modelPrefix)
				end := strings.Index(content[start:], `"`)
				if end > 0 {
					model := content[start : start+end]
					cfg.GeminiModel = model
				}
			}
		}
	}

	// Unmarshal the Viper configuration into the struct
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

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
		return nil, err
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
	if cfg.Outputs.EnableCLI == false && len(cfg.Outputs.DiscordWebhookURLs) == 0 && len(cfg.Outputs.TelegramBots) == 0 {
		// If no outputs are configured, use the legacy OutputFormat field
		if cfg.OutputFormat == "cli" || cfg.OutputFormat == "" {
			cfg.Outputs.EnableCLI = true
		}
	}

	return cfg, nil
}
