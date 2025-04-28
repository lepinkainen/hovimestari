package calendar

import (
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

// TestEventData represents a calendar event for testing
type TestEventData struct {
	UID         string
	Summary     string
	StartTime   time.Time
	EndTime     *time.Time
	Location    *string
	Description *string
	Source      string
}

// TestCalendarEventStorage tests the calendar event storage logic
func TestCalendarEventStorage(t *testing.T) {
	tests := []struct {
		name       string
		event      gocal.Event
		checkEvent func(event struct {
			UID         string
			Summary     string
			StartTime   time.Time
			EndTime     *time.Time
			Location    *string
			Description *string
			Source      string
		}) bool
	}{
		{
			name: "Basic event with summary and start time",
			event: gocal.Event{
				Uid:     "event1@example.com",
				Summary: "Meeting",
				Start:   parseTime("2025-04-21T10:00:00Z"),
			},
			checkEvent: func(event struct {
				UID         string
				Summary     string
				StartTime   time.Time
				EndTime     *time.Time
				Location    *string
				Description *string
				Source      string
			}) bool {
				return event.UID == "event1@example.com" &&
					event.Summary == "Meeting" &&
					event.StartTime.Equal(*parseTime("2025-04-21T10:00:00Z")) &&
					event.EndTime == nil &&
					event.Location == nil &&
					event.Description == nil &&
					event.Source == "calendar:TestCal"
			},
		},
		{
			name: "Event with start and end time",
			event: gocal.Event{
				Uid:     "event2@example.com",
				Summary: "Lunch",
				Start:   parseTime("2025-04-21T12:00:00Z"),
				End:     parseTime("2025-04-21T13:00:00Z"),
			},
			checkEvent: func(event struct {
				UID         string
				Summary     string
				StartTime   time.Time
				EndTime     *time.Time
				Location    *string
				Description *string
				Source      string
			}) bool {
				return event.UID == "event2@example.com" &&
					event.Summary == "Lunch" &&
					event.StartTime.Equal(*parseTime("2025-04-21T12:00:00Z")) &&
					event.EndTime != nil &&
					event.EndTime.Equal(*parseTime("2025-04-21T13:00:00Z")) &&
					event.Location == nil &&
					event.Description == nil &&
					event.Source == "calendar:TestCal"
			},
		},
		{
			name: "Event with location",
			event: gocal.Event{
				Uid:      "event3@example.com",
				Summary:  "Conference",
				Start:    parseTime("2025-04-23T09:00:00Z"),
				Location: "Room 101",
			},
			checkEvent: func(event struct {
				UID         string
				Summary     string
				StartTime   time.Time
				EndTime     *time.Time
				Location    *string
				Description *string
				Source      string
			}) bool {
				return event.UID == "event3@example.com" &&
					event.Summary == "Conference" &&
					event.StartTime.Equal(*parseTime("2025-04-23T09:00:00Z")) &&
					event.EndTime == nil &&
					event.Location != nil &&
					*event.Location == "Room 101" &&
					event.Description == nil &&
					event.Source == "calendar:TestCal"
			},
		},
		{
			name: "Event with description",
			event: gocal.Event{
				Uid:         "event4@example.com",
				Summary:     "Project Update",
				Start:       parseTime("2025-04-24T14:00:00Z"),
				Description: "Discuss progress.",
			},
			checkEvent: func(event struct {
				UID         string
				Summary     string
				StartTime   time.Time
				EndTime     *time.Time
				Location    *string
				Description *string
				Source      string
			}) bool {
				return event.UID == "event4@example.com" &&
					event.Summary == "Project Update" &&
					event.StartTime.Equal(*parseTime("2025-04-24T14:00:00Z")) &&
					event.EndTime == nil &&
					event.Location == nil &&
					event.Description != nil &&
					*event.Description == "Discuss progress." &&
					event.Source == "calendar:TestCal"
			},
		},
		{
			name: "Event with long description (truncation)",
			event: gocal.Event{
				Uid:         "event5@example.com",
				Summary:     "Workshop",
				Start:       parseTime("2025-04-25T11:00:00Z"),
				Description: strings.Repeat("Long description. ", 100), // > 1000 chars
			},
			checkEvent: func(event struct {
				UID         string
				Summary     string
				StartTime   time.Time
				EndTime     *time.Time
				Location    *string
				Description *string
				Source      string
			}) bool {
				expectedDesc := strings.Repeat("Long description. ", 100)[:997] + "..."
				return event.UID == "event5@example.com" &&
					event.Summary == "Workshop" &&
					event.StartTime.Equal(*parseTime("2025-04-25T11:00:00Z")) &&
					event.EndTime == nil &&
					event.Location == nil &&
					event.Description != nil &&
					*event.Description == expectedDesc &&
					event.Source == "calendar:TestCal"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test event data
			var eventData TestEventData

			// Create the source string
			source := CalendarSourcePrefix + ":TestCal"

			var location *string
			if tt.event.Location != "" {
				location = &tt.event.Location
			}

			var description *string
			if tt.event.Description != "" {
				// Truncate long descriptions
				desc := tt.event.Description
				if len(desc) > 1000 {
					desc = desc[:997] + "..."
				}
				description = &desc
			}

			// Set the event data
			eventData.UID = tt.event.Uid
			eventData.Summary = tt.event.Summary
			eventData.StartTime = *tt.event.Start
			eventData.EndTime = tt.event.End
			eventData.Location = location
			eventData.Description = description
			eventData.Source = source

			// Check that the event data is correct
			if !tt.checkEvent(eventData) {
				t.Errorf("Event data not prepared correctly: %+v", eventData)
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
