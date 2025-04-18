package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	// Database configuration
	DBPath string `json:"db_path"`

	// LLM configuration
	GeminiAPIKey string `json:"gemini_api_key"`

	// Calendar configuration
	WebCalURL string `json:"webcal_url"`

	// Output configuration
	OutputFormat string `json:"output_format"` // "cli", "telegram", etc.
}

// LoadConfig loads the configuration from the specified file
func LoadConfig(configPath string) (*Config, error) {
	// Set default values
	config := &Config{
		DBPath:       "memories.db",
		OutputFormat: "cli",
	}

	// If config file exists, load it
	if configPath != "" {
		file, err := os.Open(configPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to open config file: %w", err)
			}
			// If file doesn't exist, we'll use defaults and environment variables
		} else {
			defer file.Close()
			decoder := json.NewDecoder(file)
			if err := decoder.Decode(config); err != nil {
				return nil, fmt.Errorf("failed to decode config file: %w", err)
			}
		}
	}

	// Override with environment variables if they exist
	if dbPath := os.Getenv("STEVENS_DB_PATH"); dbPath != "" {
		config.DBPath = dbPath
	}

	if apiKey := os.Getenv("STEVENS_GEMINI_API_KEY"); apiKey != "" {
		config.GeminiAPIKey = apiKey
	}

	if webCalURL := os.Getenv("STEVENS_WEBCAL_URL"); webCalURL != "" {
		config.WebCalURL = webCalURL
	}

	if outputFormat := os.Getenv("STEVENS_OUTPUT_FORMAT"); outputFormat != "" {
		config.OutputFormat = outputFormat
	}

	// Validate required configuration
	if config.GeminiAPIKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}

	if config.WebCalURL == "" {
		return nil, fmt.Errorf("WebCal URL is required")
	}

	// Ensure DB path is absolute
	if !filepath.IsAbs(config.DBPath) {
		absPath, err := filepath.Abs(config.DBPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for DB: %w", err)
		}
		config.DBPath = absPath
	}

	return config, nil
}

// SaveConfig saves the configuration to the specified file
func SaveConfig(config *Config, configPath string) error {
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}
