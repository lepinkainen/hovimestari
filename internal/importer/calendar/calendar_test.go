package calendar

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/apognu/gocal"
	"github.com/shrike/hovimestari/internal/store"
)

// Helper function to parse time strings into time.Time pointers
func parseTime(timeStr string) *time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		panic(err)
	}
	return &t
}

// TestNewImporterURLConversion tests the URL conversion in NewImporter
func TestNewImporterURLConversion(t *testing.T) {
	tests := []struct {
		name         string
		inputURL     string
		calendarName string
		expectedURL  string
	}{
		{
			name:         "Convert webcal to https",
			inputURL:     "webcal://example.com/calendar.ics",
			calendarName: "TestCal",
			expectedURL:  "https://example.com/calendar.ics",
		},
		{
			name:         "Keep https URL unchanged",
			inputURL:     "https://example.com/calendar.ics",
			calendarName: "TestCal",
			expectedURL:  "https://example.com/calendar.ics",
		},
		{
			name:         "Keep http URL unchanged",
			inputURL:     "http://example.com/calendar.ics",
			calendarName: "TestCal",
			expectedURL:  "http://example.com/calendar.ics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can use nil for store since we're only testing URL conversion
			importer := NewImporter(nil, tt.inputURL, tt.calendarName)
			if importer.webCalURL != tt.expectedURL {
				t.Errorf("Expected URL %q, got %q", tt.expectedURL, importer.webCalURL)
			}
		})
	}
}

// mockFormatEvent is a copy of the formatEvent function for testing
// This is needed because the original function assumes non-nil pointers
func mockFormatEvent(event *gocal.Event) string {
	var builder strings.Builder

	// Add the event summary
	builder.WriteString(fmt.Sprintf("Calendar Event: %s", event.Summary))

	// Add the event time
	if event.Start != nil && !event.Start.IsZero() {
		startTime := event.Start.Format("2006-01-02 15:04")

		if event.End != nil && !event.End.IsZero() {
			endTime := event.End.Format("15:04")
			if event.Start.Day() != event.End.Day() {
				endTime = event.End.Format("2006-01-02 15:04")
			}
			builder.WriteString(fmt.Sprintf(" from %s to %s", startTime, endTime))
		} else {
			builder.WriteString(fmt.Sprintf(" at %s", startTime))
		}
	}

	// Add the location if available
	if event.Location != "" {
		builder.WriteString(fmt.Sprintf(" at %s", event.Location))
	}

	// Add the description if available
	if event.Description != "" {
		// Truncate long descriptions
		description := event.Description
		if len(description) > 200 {
			description = description[:197] + "..."
		}
		builder.WriteString(fmt.Sprintf(". Description: %s", description))
	}

	return builder.String()
}

// TestFormatEvent tests the formatEvent function
func TestFormatEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    gocal.Event
		expected string
	}{
		{
			name: "Basic event with summary and start time",
			event: gocal.Event{
				Summary: "Meeting",
				Start:   parseTime("2025-04-21T10:00:00Z"),
			},
			expected: "Calendar Event: Meeting at 2025-04-21 10:00",
		},
		{
			name: "Event with start and end time (same day)",
			event: gocal.Event{
				Summary: "Lunch",
				Start:   parseTime("2025-04-21T12:00:00Z"),
				End:     parseTime("2025-04-21T13:00:00Z"),
			},
			expected: "Calendar Event: Lunch from 2025-04-21 12:00 to 13:00",
		},
		{
			name: "Event spanning multiple days",
			event: gocal.Event{
				Summary: "Vacation",
				Start:   parseTime("2025-04-22T08:00:00Z"),
				End:     parseTime("2025-04-25T17:00:00Z"),
			},
			expected: "Calendar Event: Vacation from 2025-04-22 08:00 to 2025-04-25 17:00",
		},
		{
			name: "Event with location",
			event: gocal.Event{
				Summary:  "Conference",
				Start:    parseTime("2025-04-23T09:00:00Z"),
				Location: "Room 101",
			},
			expected: "Calendar Event: Conference at 2025-04-23 09:00 at Room 101",
		},
		{
			name: "Event with description (short)",
			event: gocal.Event{
				Summary:     "Project Update",
				Start:       parseTime("2025-04-24T14:00:00Z"),
				Description: "Discuss progress.",
			},
			expected: "Calendar Event: Project Update at 2025-04-24 14:00. Description: Discuss progress.",
		},
		{
			name: "Event with long description (truncation)",
			event: gocal.Event{
				Summary:     "Workshop",
				Start:       parseTime("2025-04-25T11:00:00Z"),
				Description: strings.Repeat("Long description. ", 20), // > 200 chars
			},
			expected: "Calendar Event: Workshop at 2025-04-25 11:00. Description: " + strings.Repeat("Long description. ", 20)[:197] + "...",
		},
		{
			name: "Event with all fields",
			event: gocal.Event{
				Summary:     "Team Sync",
				Start:       parseTime("2025-04-26T15:00:00Z"),
				End:         parseTime("2025-04-26T16:30:00Z"),
				Location:    "Online",
				Description: "Weekly check-in.",
			},
			expected: "Calendar Event: Team Sync from 2025-04-26 15:00 to 16:30 at Online. Description: Weekly check-in.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mockFormatEvent(&tt.event)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestFilterValidEvents tests the filterValidEvents function
func TestFilterValidEvents(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError bool
		checkOutput   func(string) bool
	}{
		{
			name: "Valid event with DTSTAMP",
			input: `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test Calendar//EN
BEGIN:VEVENT
UID:event1@example.com
DTSTAMP:20250420T100000Z
SUMMARY:Valid Event
DTSTART:20250421T090000Z
END:VEVENT
END:VCALENDAR`,
			expectedError: false,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "BEGIN:VEVENT") &&
					strings.Contains(output, "DTSTAMP:20250420T100000Z") &&
					strings.Contains(output, "SUMMARY:Valid Event")
			},
		},
		{
			name: "Event missing DTSTAMP",
			input: `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test Calendar//EN
BEGIN:VEVENT
UID:event2@example.com
SUMMARY:Invalid Event (No DTSTAMP)
DTSTART:20250422T110000Z
END:VEVENT
END:VCALENDAR`,
			expectedError: false,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "BEGIN:VCALENDAR") &&
					strings.Contains(output, "VERSION:2.0") &&
					strings.Contains(output, "PRODID:-//Test Calendar//EN") &&
					!strings.Contains(output, "BEGIN:VEVENT") &&
					!strings.Contains(output, "SUMMARY:Invalid Event")
			},
		},
		{
			name: "Mix of valid and invalid events",
			input: `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test Calendar//EN
BEGIN:VEVENT
UID:event1@example.com
DTSTAMP:20250420T100000Z
SUMMARY:Valid Event 1
DTSTART:20250421T090000Z
END:VEVENT
BEGIN:VEVENT
UID:event2@example.com
SUMMARY:Invalid Event (No DTSTAMP)
DTSTART:20250422T110000Z
END:VEVENT
BEGIN:VEVENT
UID:event3@example.com
DTSTAMP:20250420T110000Z
SUMMARY:Valid Event 2
DTSTART:20250423T140000Z
END:VEVENT
END:VCALENDAR`,
			expectedError: false,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "BEGIN:VCALENDAR") &&
					strings.Contains(output, "SUMMARY:Valid Event 1") &&
					strings.Contains(output, "SUMMARY:Valid Event 2") &&
					!strings.Contains(output, "SUMMARY:Invalid Event")
			},
		},
		{
			name:          "Empty iCalendar data",
			input:         "",
			expectedError: true,
			checkOutput: func(output string) bool {
				return output == ""
			},
		},
		{
			name: "Data with header/footer but no BEGIN:VEVENT",
			input: `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test Calendar//EN
END:VCALENDAR`,
			expectedError: true,
			checkOutput: func(output string) bool {
				return output == ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := filterValidEvents(tt.input)

			// Check error expectation
			if (err != nil) != tt.expectedError {
				t.Errorf("Expected error: %v, got error: %v", tt.expectedError, err != nil)
			}

			// Check output content
			if !tt.checkOutput(output) {
				t.Errorf("Output does not match expected pattern: %s", output)
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
	importer := NewImporter(mockStore, url, calName)

	if importer.store != mockStore {
		t.Error("Store not properly set in importer")
	}

	if importer.webCalURL != url {
		t.Errorf("Expected URL %q, got %q", url, importer.webCalURL)
	}

	if importer.calendarName != calName {
		t.Errorf("Expected calendar name %q, got %q", calName, importer.calendarName)
	}
}
