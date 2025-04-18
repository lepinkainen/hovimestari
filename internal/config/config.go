package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CalendarConfig holds configuration for a calendar source
type CalendarConfig struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// FamilyMember represents a family member with optional birthday and Telegram ID
type FamilyMember struct {
	Name       string `json:"name"`
	Birthday   string `json:"birthday,omitempty"` // Format: YYYY-MM-DD
	TelegramID string `json:"telegram_id,omitempty"`
}

// Config holds the application configuration
type Config struct {
	// Database configuration
	DBPath string `json:"db_path"`

	// LLM configuration
	GeminiAPIKey string `json:"gemini_api_key"`

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
	if dbPath := os.Getenv("HOVIMESTARI_DB_PATH"); dbPath != "" {
		config.DBPath = dbPath
	}

	if apiKey := os.Getenv("HOVIMESTARI_GEMINI_API_KEY"); apiKey != "" {
		config.GeminiAPIKey = apiKey
	}

	if outputFormat := os.Getenv("HOVIMESTARI_OUTPUT_FORMAT"); outputFormat != "" {
		config.OutputFormat = outputFormat
	}

	// Validate required configuration
	if config.GeminiAPIKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}

	// Validate location configuration
	if config.LocationName == "" {
		return nil, fmt.Errorf("location_name is required")
	}

	if config.Latitude < -90 || config.Latitude > 90 {
		return nil, fmt.Errorf("latitude must be between -90 and 90")
	}

	if config.Longitude < -180 || config.Longitude > 180 {
		return nil, fmt.Errorf("longitude must be between -180 and 180")
	}

	// Validate timezone
	if config.Timezone == "" {
		return nil, fmt.Errorf("timezone is required")
	}

	// Try to load the timezone to validate it
	_, err := time.LoadLocation(config.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %w", err)
	}

	// Validate calendars
	if len(config.Calendars) == 0 {
		return nil, fmt.Errorf("at least one calendar is required")
	}

	for i, cal := range config.Calendars {
		if cal.Name == "" {
			return nil, fmt.Errorf("calendar %d is missing a name", i+1)
		}
		if cal.URL == "" {
			return nil, fmt.Errorf("calendar %d (%s) is missing a URL", i+1, cal.Name)
		}
	}

	// Validate family members
	if len(config.Family) == 0 {
		return nil, fmt.Errorf("at least one family member is required")
	}

	for i, member := range config.Family {
		if member.Name == "" {
			return nil, fmt.Errorf("family member %d is missing a name", i+1)
		}

		// Validate birthday format if provided
		if member.Birthday != "" {
			_, err := time.Parse("2006-01-02", member.Birthday)
			if err != nil {
				return nil, fmt.Errorf("invalid birthday format for %s: %w", member.Name, err)
			}
		}
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
