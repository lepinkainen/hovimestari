package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const (
	// MetNoAPIURL is the base URL for the MET Norway Locationforecast API
	MetNoAPIURL = "https://api.met.no/weatherapi/locationforecast/2.0/compact"
	// UserAgent is required by the MET Norway API
	UserAgent = "Hovimestari/1.0 github.com/lepinkainen/hovimestari"
)

// MetNoForecast represents the response from the MET Norway API
type MetNoForecast struct {
	Properties struct {
		Timeseries []struct {
			Time time.Time `json:"time"`
			Data struct {
				Instant struct {
					Details struct {
						AirTemperature           float64 `json:"air_temperature"`
						RelativeHumidity         float64 `json:"relative_humidity"`
						WindSpeed                float64 `json:"wind_speed"`
						WindFromDirection        float64 `json:"wind_from_direction"`
						UltravioletIndexClearSky float64 `json:"ultraviolet_index_clear_sky"`
					} `json:"details"`
				} `json:"instant"`
				Next1Hours *struct {
					Summary struct {
						SymbolCode string `json:"symbol_code"`
					} `json:"summary"`
				} `json:"next_1_hours,omitempty"`
				Next6Hours *struct {
					Summary struct {
						SymbolCode string `json:"symbol_code"`
					} `json:"summary"`
				} `json:"next_6_hours,omitempty"`
				Next12Hours *struct {
					Summary struct {
						SymbolCode string `json:"symbol_code"`
					} `json:"summary"`
				} `json:"next_12_hours,omitempty"`
			} `json:"data"`
		} `json:"timeseries"`
	} `json:"properties"`
}

// DailyForecast represents a summarized forecast for a single day
type DailyForecast struct {
	Date        time.Time
	MinTemp     float64
	MaxTemp     float64
	SymbolCode  string
	Description string
	WindSpeed   float64
	UVIndex     float64
}

// GetForecast fetches the weather forecast for the given location
func GetForecast(latitude, longitude float64) (string, error) {
	// Construct the API URL
	url := fmt.Sprintf("%s?lat=%.6f&lon=%.6f", MetNoAPIURL, latitude, longitude)

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("User-Agent", UserAgent)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("Failed to close response body", "error", err)
		}
	}()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned non-OK status: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the JSON response
	var forecast MetNoForecast
	if err := json.Unmarshal(body, &forecast); err != nil {
		return "", fmt.Errorf("failed to parse weather data: %w", err)
	}

	// Extract relevant forecast data
	if len(forecast.Properties.Timeseries) == 0 {
		return "", fmt.Errorf("no forecast data available")
	}

	// Get the current weather
	current := forecast.Properties.Timeseries[0]
	currentTemp := current.Data.Instant.Details.AirTemperature

	// Find the min and max temperatures for the day
	var minTemp, maxTemp = currentTemp, currentTemp
	var symbolCode string

	// Look for the next 24 hours
	now := time.Now()
	endTime := now.Add(24 * time.Hour)

	for _, ts := range forecast.Properties.Timeseries {
		if ts.Time.After(endTime) {
			break
		}

		temp := ts.Data.Instant.Details.AirTemperature
		if temp < minTemp {
			minTemp = temp
		}
		if temp > maxTemp {
			maxTemp = temp
		}

		// Get the weather symbol for the next period
		if symbolCode == "" {
			if ts.Data.Next1Hours != nil {
				symbolCode = ts.Data.Next1Hours.Summary.SymbolCode
			} else if ts.Data.Next6Hours != nil {
				symbolCode = ts.Data.Next6Hours.Summary.SymbolCode
			} else if ts.Data.Next12Hours != nil {
				symbolCode = ts.Data.Next12Hours.Summary.SymbolCode
			}
		}
	}

	// Use the symbol code directly (no translation)
	weatherDesc := "variable"
	if symbolCode != "" {
		weatherDesc = symbolCode
	}

	// Format the forecast
	forecastText := fmt.Sprintf("Weather today: %s, temperature %.0f-%.0f°C", weatherDesc, minTemp, maxTemp)
	return forecastText, nil
}

// GetMultiDayForecast fetches weather forecasts for multiple days
func GetMultiDayForecast(latitude, longitude float64) ([]DailyForecast, error) {
	// Construct the API URL
	url := fmt.Sprintf("%s?lat=%.6f&lon=%.6f", MetNoAPIURL, latitude, longitude)

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("User-Agent", UserAgent)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("Failed to close response body", "error", err)
		}
	}()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-OK status: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the JSON response
	var forecast MetNoForecast
	if err := json.Unmarshal(body, &forecast); err != nil {
		return nil, fmt.Errorf("failed to parse weather data: %w", err)
	}

	// Extract relevant forecast data
	if len(forecast.Properties.Timeseries) == 0 {
		return nil, fmt.Errorf("no forecast data available")
	}

	// Group forecasts by day
	dailyForecasts := make(map[string]*DailyForecast)

	// Get the timezone from the local system
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return nil, fmt.Errorf("failed to get local timezone: %w", err)
	}

	// Process each timeseries entry
	for _, ts := range forecast.Properties.Timeseries {
		// Convert to local time
		localTime := ts.Time.In(loc)

		// Format date as YYYY-MM-DD for grouping
		dateKey := localTime.Format("2006-01-02")

		// Process all available days

		// Get temperature and other details
		temp := ts.Data.Instant.Details.AirTemperature
		windSpeed := ts.Data.Instant.Details.WindSpeed
		uvIndex := ts.Data.Instant.Details.UltravioletIndexClearSky

		// Initialize daily forecast if not exists
		if dailyForecasts[dateKey] == nil {
			date, _ := time.Parse("2006-01-02", dateKey)
			dailyForecasts[dateKey] = &DailyForecast{
				Date:      date,
				MinTemp:   temp,
				MaxTemp:   temp,
				WindSpeed: windSpeed,
				UVIndex:   uvIndex,
			}
		}

		// Update min/max temperature
		if temp < dailyForecasts[dateKey].MinTemp {
			dailyForecasts[dateKey].MinTemp = temp
		}
		if temp > dailyForecasts[dateKey].MaxTemp {
			dailyForecasts[dateKey].MaxTemp = temp
		}

		// Update wind speed (use average or max as needed)
		dailyForecasts[dateKey].WindSpeed = (dailyForecasts[dateKey].WindSpeed + windSpeed) / 2

		// Update UV index (use maximum value for the day)
		if uvIndex > dailyForecasts[dateKey].UVIndex {
			dailyForecasts[dateKey].UVIndex = uvIndex
		}

		// Get the weather symbol for the day
		// Prefer symbols from daytime hours (8:00 - 20:00)
		if dailyForecasts[dateKey].SymbolCode == "" || (localTime.Hour() >= 8 && localTime.Hour() <= 20) {
			var symbolCode string
			if ts.Data.Next6Hours != nil {
				symbolCode = ts.Data.Next6Hours.Summary.SymbolCode
			} else if ts.Data.Next1Hours != nil {
				symbolCode = ts.Data.Next1Hours.Summary.SymbolCode
			} else if ts.Data.Next12Hours != nil {
				symbolCode = ts.Data.Next12Hours.Summary.SymbolCode
			}

			if symbolCode != "" {
				dailyForecasts[dateKey].SymbolCode = symbolCode
				// Store the symbol code directly in the Description field
				dailyForecasts[dateKey].Description = symbolCode
			}
		}
	}

	// Convert map to slice and sort by date
	var result []DailyForecast
	for _, forecast := range dailyForecasts {
		result = append(result, *forecast)
	}

	// Sort by date
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Date.After(result[j].Date) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	// Return all available days
	return result, nil
}

// GetCurrentDayHourlyForecast fetches hourly weather forecasts for the current day
func GetCurrentDayHourlyForecast(latitude, longitude float64) (string, error) {
	// Construct the API URL
	url := fmt.Sprintf("%s?lat=%.6f&lon=%.6f", MetNoAPIURL, latitude, longitude)

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("User-Agent", UserAgent)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("Failed to close response body", "error", err)
		}
	}()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned non-OK status: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the JSON response
	var forecast MetNoForecast
	if err := json.Unmarshal(body, &forecast); err != nil {
		return "", fmt.Errorf("failed to parse weather data: %w", err)
	}

	// Extract relevant forecast data
	if len(forecast.Properties.Timeseries) == 0 {
		return "", fmt.Errorf("no forecast data available")
	}

	// Get the timezone from the local system
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return "", fmt.Errorf("failed to get local timezone: %w", err)
	}

	// Get current time in local timezone
	now := time.Now().In(loc)

	// Calculate the end of the current day
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, loc)

	// Collect hourly forecasts for the current day
	var hourlyForecasts []string

	for _, ts := range forecast.Properties.Timeseries {
		// Convert to local time
		localTime := ts.Time.In(loc)

		// Skip entries from previous hours or after today
		if localTime.Before(now) || localTime.After(endOfDay) {
			continue
		}

		// Get temperature
		temp := ts.Data.Instant.Details.AirTemperature

		// Get weather symbol
		var symbolCode string
		if ts.Data.Next1Hours != nil {
			symbolCode = ts.Data.Next1Hours.Summary.SymbolCode
		} else if ts.Data.Next6Hours != nil {
			symbolCode = ts.Data.Next6Hours.Summary.SymbolCode
		} else if ts.Data.Next12Hours != nil {
			symbolCode = ts.Data.Next12Hours.Summary.SymbolCode
		} else {
			symbolCode = "unknown"
		}

		// Format the hourly forecast
		hourlyForecast := fmt.Sprintf("%s: %.0f°C (%s)",
			localTime.Format("15:04"),
			temp,
			symbolCode)

		hourlyForecasts = append(hourlyForecasts, hourlyForecast)

		// Limit to the next 12 hours to keep it concise
		if len(hourlyForecasts) >= 12 {
			break
		}
	}

	// If no hourly forecasts were found
	if len(hourlyForecasts) == 0 {
		return "No hourly forecast data available for today", nil
	}

	// Join the hourly forecasts with commas
	result := fmt.Sprintf("Hourly forecast for today: %s", strings.Join(hourlyForecasts, ", "))

	return result, nil
}

// FormatDailyForecast formats a daily forecast as a string
func FormatDailyForecast(forecast DailyForecast) string {
	var result string

	// Base format with temperature
	baseFormat := "Weather %s: %s, temperature %.0f-%.0f°C"

	// Add wind speed if it's over 5 m/s
	if forecast.WindSpeed > 5.0 {
		result = fmt.Sprintf(baseFormat+", wind speed %.1f m/s",
			forecast.Date.Format("2006-01-02"),
			forecast.Description,
			forecast.MinTemp,
			forecast.MaxTemp,
			forecast.WindSpeed)
	} else {
		result = fmt.Sprintf(baseFormat,
			forecast.Date.Format("2006-01-02"),
			forecast.Description,
			forecast.MinTemp,
			forecast.MaxTemp)
	}

	// Add UV index if it's 3.0 or higher
	if forecast.UVIndex >= 3.0 {
		result = fmt.Sprintf("%s, Max UV Index: %.1f", result, forecast.UVIndex)
	}

	return result
}
