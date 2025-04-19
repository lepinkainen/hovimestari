package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// MetNoAPIURL is the base URL for the MET Norway Locationforecast API
	MetNoAPIURL = "https://api.met.no/weatherapi/locationforecast/2.0/compact"
	// UserAgent is required by the MET Norway API
	UserAgent = "Hovimestari/1.0 github.com/shrike/hovimestari"
)

// MetNoForecast represents the response from the MET Norway API
type MetNoForecast struct {
	Properties struct {
		Timeseries []struct {
			Time time.Time `json:"time"`
			Data struct {
				Instant struct {
					Details struct {
						AirTemperature    float64 `json:"air_temperature"`
						RelativeHumidity  float64 `json:"relative_humidity"`
						WindSpeed         float64 `json:"wind_speed"`
						WindFromDirection float64 `json:"wind_from_direction"`
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
}

// WeatherSymbols maps MET Norway symbol codes to Finnish descriptions
var WeatherSymbols = map[string]string{
	"clearsky":                    "selkeää",
	"fair":                        "poutaa",
	"partlycloudy":                "puolipilvistä",
	"cloudy":                      "pilvistä",
	"rainshowers":                 "sadekuuroja",
	"rainshowersandthunder":       "sadekuuroja ja ukkosta",
	"sleetshowers":                "räntäkuuroja",
	"snowshowers":                 "lumikuuroja",
	"rain":                        "sadetta",
	"heavyrain":                   "rankkasadetta",
	"heavyrainandthunder":         "rankkasadetta ja ukkosta",
	"sleet":                       "räntää",
	"snow":                        "lumisadetta",
	"snowandthunder":              "lumisadetta ja ukkosta",
	"fog":                         "sumua",
	"sleetshowersandthunder":      "räntäkuuroja ja ukkosta",
	"snowshowersandthunder":       "lumikuuroja ja ukkosta",
	"rainandthunder":              "sadetta ja ukkosta",
	"sleetandthunder":             "räntää ja ukkosta",
	"lightrainshowersandthunder":  "kevyitä sadekuuroja ja ukkosta",
	"heavyrainshowersandthunder":  "voimakkaita sadekuuroja ja ukkosta",
	"lightsleetshowersandthunder": "kevyitä räntäkuuroja ja ukkosta",
	"heavysleetshowersandthunder": "voimakkaita räntäkuuroja ja ukkosta",
	"lightsnowshowersandthunder":  "kevyitä lumikuuroja ja ukkosta",
	"heavysnowshowersandthunder":  "voimakkaita lumikuuroja ja ukkosta",
	"lightrainandthunder":         "kevyttä sadetta ja ukkosta",
	"lightsleetandthunder":        "kevyttä räntää ja ukkosta",
	"heavysleetandthunder":        "voimakasta räntää ja ukkosta",
	"lightsnowandthunder":         "kevyttä lumisadetta ja ukkosta",
	"heavysnowandthunder":         "voimakasta lumisadetta ja ukkosta",
	"lightrainshowers":            "kevyitä sadekuuroja",
	"heavyrainshowers":            "voimakkaita sadekuuroja",
	"lightsleetshowers":           "kevyitä räntäkuuroja",
	"heavysleetshowers":           "voimakkaita räntäkuuroja",
	"lightsnowshowers":            "kevyitä lumikuuroja",
	"heavysnowshowers":            "voimakkaita lumikuuroja",
	"lightrain":                   "kevyttä sadetta",
	"lightsleet":                  "kevyttä räntää",
	"heavysleet":                  "voimakasta räntää",
	"lightsnow":                   "kevyttä lumisadetta",
	"heavysnow":                   "voimakasta lumisadetta",
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
	defer resp.Body.Close()

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

	// Translate the symbol code to Finnish
	weatherDesc := "vaihtelevaa"
	if desc, ok := WeatherSymbols[symbolCode]; ok {
		weatherDesc = desc
	}

	// Format the forecast
	forecastText := fmt.Sprintf("Sää tänään: %s, lämpötila %.0f-%.0f°C", weatherDesc, minTemp, maxTemp)
	return forecastText, nil
}

// GetMultiDayForecast fetches weather forecasts for multiple days
func GetMultiDayForecast(latitude, longitude float64, days int) ([]DailyForecast, error) {
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
	defer resp.Body.Close()

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

		// Skip if we already have enough days
		if len(dailyForecasts) >= days && dailyForecasts[dateKey] == nil {
			continue
		}

		// Get temperature and other details
		temp := ts.Data.Instant.Details.AirTemperature
		windSpeed := ts.Data.Instant.Details.WindSpeed

		// Initialize daily forecast if not exists
		if dailyForecasts[dateKey] == nil {
			date, _ := time.Parse("2006-01-02", dateKey)
			dailyForecasts[dateKey] = &DailyForecast{
				Date:      date,
				MinTemp:   temp,
				MaxTemp:   temp,
				WindSpeed: windSpeed,
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
				if desc, ok := WeatherSymbols[symbolCode]; ok {
					dailyForecasts[dateKey].Description = desc
				} else {
					dailyForecasts[dateKey].Description = "vaihtelevaa"
				}
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

	// Limit to requested number of days
	if len(result) > days {
		result = result[:days]
	}

	return result, nil
}

// FormatDailyForecast formats a daily forecast as a string
func FormatDailyForecast(forecast DailyForecast) string {
	// Only include wind speed if it's over 5 m/s
	if forecast.WindSpeed > 5.0 {
		return fmt.Sprintf("Sää %s: %s, lämpötila %.0f-%.0f°C, tuulen nopeus %.1f m/s",
			forecast.Date.Format("2006-01-02"),
			forecast.Description,
			forecast.MinTemp,
			forecast.MaxTemp,
			forecast.WindSpeed)
	}

	// Otherwise, just include temperature and conditions
	return fmt.Sprintf("Sää %s: %s, lämpötila %.0f-%.0f°C",
		forecast.Date.Format("2006-01-02"),
		forecast.Description,
		forecast.MinTemp,
		forecast.MaxTemp)
}
