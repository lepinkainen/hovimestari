package calendar

import (
	"testing"

	"github.com/lepinkainen/hovimestari/internal/store"
)

// TestNewImporterURLConversion tests the URL conversion in NewImporter
func TestNewImporterURLConversion(t *testing.T) {
	tests := []struct {
		name         string
		inputURL     string
		calendarName string
		updateMode   string
		expectedURL  string
	}{
		{
			name:         "Convert webcal to https",
			inputURL:     "webcal://example.com/calendar.ics",
			calendarName: "TestCal",
			updateMode:   "full_refresh",
			expectedURL:  "https://example.com/calendar.ics",
		},
		{
			name:         "Keep https URL unchanged",
			inputURL:     "https://example.com/calendar.ics",
			calendarName: "TestCal",
			updateMode:   "smart",
			expectedURL:  "https://example.com/calendar.ics",
		},
		{
			name:         "Keep http URL unchanged",
			inputURL:     "http://example.com/calendar.ics",
			calendarName: "TestCal",
			updateMode:   "",
			expectedURL:  "http://example.com/calendar.ics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can use nil for store since we're only testing URL conversion
			importer := NewImporter(nil, tt.inputURL, tt.calendarName, tt.updateMode)
			if importer.webCalURL != tt.expectedURL {
				t.Errorf("Expected URL %q, got %q", tt.expectedURL, importer.webCalURL)
			}
		})
	}
}

// TestNewImporter tests the NewImporter function
func TestNewImporter(t *testing.T) {
	// Create a mock store
	mockStore := &store.Store{}

	// Test with a regular URL
	url := "https://example.com/calendar.ics"
	calName := "Test Calendar"
	updateMode := "full_refresh"
	importer := NewImporter(mockStore, url, calName, updateMode)

	if importer.store != mockStore {
		t.Error("Store not properly set in importer")
	}

	if importer.webCalURL != url {
		t.Errorf("Expected URL %q, got %q", url, importer.webCalURL)
	}

	if importer.calendarName != calName {
		t.Errorf("Expected calendar name %q, got %q", calName, importer.calendarName)
	}

	if importer.updateStrategy != UpdateStrategyReplaceAll {
		t.Errorf("Expected update strategy %q, got %q", UpdateStrategyReplaceAll, importer.updateStrategy)
	}

	// Test with smart update mode
	smartImporter := NewImporter(mockStore, url, calName, "smart")
	if smartImporter.updateStrategy != UpdateStrategyUpsert {
		t.Errorf("Expected update strategy %q, got %q", UpdateStrategyUpsert, smartImporter.updateStrategy)
	}
}
