package electricityprice

import (
	"strings"
	"testing"
	"time"
)

func TestIsoResolutionToDuration(t *testing.T) {
	tests := []struct {
		name       string
		resolution string
		want       time.Duration
		wantErr    bool
	}{
		{name: "hours", resolution: "PT1H", want: time.Hour},
		{name: "minutes", resolution: "PT15M", want: 15 * time.Minute},
		{name: "seconds", resolution: "PT30S", want: 30 * time.Second},
		{name: "empty", resolution: "", wantErr: true},
		{name: "unsupported prefix", resolution: "P1D", wantErr: true},
		{name: "invalid number", resolution: "PTxH", wantErr: true},
		{name: "unsupported unit", resolution: "PT1Q", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isoResolutionToDuration(tt.resolution)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("isoResolutionToDuration(%q) error = nil, want error", tt.resolution)
				}
				return
			}
			if err != nil {
				t.Fatalf("isoResolutionToDuration(%q) error = %v", tt.resolution, err)
			}
			if got != tt.want {
				t.Fatalf("isoResolutionToDuration(%q) = %v, want %v", tt.resolution, got, tt.want)
			}
		})
	}
}

func TestFormatPriceMemory(t *testing.T) {
	tz := time.FixedZone("UTC+2", 2*60*60)
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, tz)
	baseUTC := time.Date(2026, 3, 31, 22, 0, 0, 0, time.UTC) // 00:00 in UTC+2

	prices := []pricePoint{
		{utcTime: baseUTC.Add(0 * time.Hour), price: 2},
		{utcTime: baseUTC.Add(1 * time.Hour), price: 4},
		{utcTime: baseUTC.Add(2 * time.Hour), price: 8},
		{utcTime: baseUTC.Add(3 * time.Hour), price: 12},
		{utcTime: baseUTC.Add(4 * time.Hour), price: 16},
		{utcTime: baseUTC.Add(5 * time.Hour), price: 30},
		{utcTime: baseUTC.Add(6 * time.Hour), price: 40},
		{utcTime: baseUTC.Add(7 * time.Hour), price: 50},
	}

	got := formatPriceMemory(prices, date, tz)

	checks := []string{
		"Electricity prices on 2026-04-01 (EXPENSIVE DAY - avg 20.2 c/kWh). Times in UTC+2.",
		"Min: 2.0 c/kWh at 00:00, Max: 50.0 c/kWh at 07:00",
		"Notable spikes (>1.5× avg): 07:00=50.0 06:00=40.0",
		"Hourly prices (c/kWh): 00=2.0 01=4.0 02=8.0 03=12.0 04=16.0 05=30.0 06=40.0 07=50.0",
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("formatPriceMemory() missing %q in %q", check, got)
		}
	}
}

func TestFormatPriceMemoryClassificationsAndSpikeLimit(t *testing.T) {
	tz := time.UTC
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, tz)
	base := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	cheap := formatPriceMemory([]pricePoint{{base, 1}, {base.Add(time.Hour), 2}, {base.Add(2 * time.Hour), 3}}, date, tz)
	if !strings.Contains(cheap, "CHEAP DAY") {
		t.Fatalf("cheap classification missing in %q", cheap)
	}

	expensive := formatPriceMemory([]pricePoint{{base, 21}, {base.Add(time.Hour), 22}, {base.Add(2 * time.Hour), 30}}, date, tz)
	if !strings.Contains(expensive, "EXPENSIVE DAY") {
		t.Fatalf("expensive classification missing in %q", expensive)
	}

	manySpikes := []pricePoint{
		{base.Add(0 * time.Hour), 0},
		{base.Add(1 * time.Hour), 0},
		{base.Add(2 * time.Hour), 0},
		{base.Add(3 * time.Hour), 0},
		{base.Add(4 * time.Hour), 0},
		{base.Add(5 * time.Hour), 0},
		{base.Add(6 * time.Hour), 0},
		{base.Add(7 * time.Hour), 0},
		{base.Add(8 * time.Hour), 0},
		{base.Add(9 * time.Hour), 0},
		{base.Add(10 * time.Hour), 30},
		{base.Add(11 * time.Hour), 40},
		{base.Add(12 * time.Hour), 50},
		{base.Add(13 * time.Hour), 60},
		{base.Add(14 * time.Hour), 70},
		{base.Add(15 * time.Hour), 80},
		{base.Add(16 * time.Hour), 90},
	}
	got := formatPriceMemory(manySpikes, date, tz)
	if strings.Contains(got, "10:00=30.0") || strings.Contains(got, "11:00=40.0") {
		t.Fatalf("formatPriceMemory() kept more than top 5 spikes: %q", got)
	}
	for _, expected := range []string{"16:00=90.0", "15:00=80.0", "14:00=70.0", "13:00=60.0", "12:00=50.0"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("formatPriceMemory() missing top spike %q in %q", expected, got)
		}
	}
}
