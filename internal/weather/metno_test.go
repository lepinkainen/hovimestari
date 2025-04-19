package weather

import (
	"testing"
	"time"
)

// TestFormatDailyForecast tests the FormatDailyForecast function
func TestFormatDailyForecast(t *testing.T) {
	// Helper function to parse date strings
	parseDate := func(dateStr string) time.Time {
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			panic(err)
		}
		return t
	}

	tests := []struct {
		name     string
		forecast DailyForecast
		expected string
	}{
		{
			name: "Basic forecast with temperature only",
			forecast: DailyForecast{
				Date:        parseDate("2025-04-21"),
				MinTemp:     5.0,
				MaxTemp:     15.0,
				Description: "partly_cloudy",
				WindSpeed:   3.0, // Below threshold for inclusion
				UVIndex:     2.0, // Below threshold for inclusion
			},
			expected: "Weather 2025-04-21: partly_cloudy, temperature 5-15°C",
		},
		{
			name: "Forecast with high wind speed",
			forecast: DailyForecast{
				Date:        parseDate("2025-04-22"),
				MinTemp:     8.0,
				MaxTemp:     18.0,
				Description: "cloudy",
				WindSpeed:   6.5, // Above threshold for inclusion
				UVIndex:     2.5, // Below threshold for inclusion
			},
			expected: "Weather 2025-04-22: cloudy, temperature 8-18°C, wind speed 6.5 m/s",
		},
		{
			name: "Forecast with high UV index",
			forecast: DailyForecast{
				Date:        parseDate("2025-04-23"),
				MinTemp:     10.0,
				MaxTemp:     20.0,
				Description: "clear_sky",
				WindSpeed:   3.5, // Below threshold for inclusion
				UVIndex:     4.0, // Above threshold for inclusion
			},
			expected: "Weather 2025-04-23: clear_sky, temperature 10-20°C, Max UV Index: 4.0",
		},
		{
			name: "Forecast with both high wind speed and UV index",
			forecast: DailyForecast{
				Date:        parseDate("2025-04-24"),
				MinTemp:     12.0,
				MaxTemp:     22.0,
				Description: "rain",
				WindSpeed:   7.0, // Above threshold for inclusion
				UVIndex:     3.5, // Above threshold for inclusion
			},
			expected: "Weather 2025-04-24: rain, temperature 12-22°C, wind speed 7.0 m/s, Max UV Index: 3.5",
		},
		{
			name: "Forecast with negative temperatures",
			forecast: DailyForecast{
				Date:        parseDate("2025-04-25"),
				MinTemp:     -5.0,
				MaxTemp:     2.0,
				Description: "snow",
				WindSpeed:   4.0, // Below threshold for inclusion
				UVIndex:     1.0, // Below threshold for inclusion
			},
			expected: "Weather 2025-04-25: snow, temperature -5-2°C",
		},
		{
			name: "Forecast with same min and max temperature",
			forecast: DailyForecast{
				Date:        parseDate("2025-04-26"),
				MinTemp:     15.0,
				MaxTemp:     15.0,
				Description: "fog",
				WindSpeed:   2.0, // Below threshold for inclusion
				UVIndex:     1.5, // Below threshold for inclusion
			},
			expected: "Weather 2025-04-26: fog, temperature 15-15°C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDailyForecast(tt.forecast)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
