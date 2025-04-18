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
