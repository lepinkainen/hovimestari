package brief

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lepinkainen/hovimestari/internal/config"
	"github.com/lepinkainen/hovimestari/internal/store"
)

func newBriefGenerator(t *testing.T) (*Generator, *store.Store) {
	t.Helper()

	s, err := store.NewStore(filepath.Join(t.TempDir(), "brief.db"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	})
	if err := s.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	g := NewGenerator(s, nil, &config.Config{
		Timezone:     "Europe/Helsinki",
		LocationName: "Helsinki",
		Family: []config.FamilyMember{
			{Name: "Alice", Birthday: "2000-04-01"},
			{Name: "Bob", Birthday: "not-a-date"},
			{Name: "Charlie"},
		},
	})

	return g, s
}

func TestFindBirthdaysTodayAndGetFamilyNames(t *testing.T) {
	g, _ := newBriefGenerator(t)

	now := time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC)
	birthdays := g.findBirthdaysToday(now)
	if len(birthdays) != 1 || birthdays[0] != "Alice (26 years)" {
		t.Fatalf("findBirthdaysToday() = %#v, want Alice birthday", birthdays)
	}

	names := g.getFamilyNames()
	if got := strings.Join(names, ","); got != "Alice,Bob,Charlie" {
		t.Fatalf("getFamilyNames() = %q, want %q", got, "Alice,Bob,Charlie")
	}
}

func TestAssembleUserInfo(t *testing.T) {
	g, _ := newBriefGenerator(t)
	now := time.Date(2026, 4, 1, 7, 30, 0, 0, time.UTC)

	userInfo := g.assembleUserInfo(
		now,
		[]string{"Alice", "Bob"},
		[]string{"Breakfast (until 08:00)"},
		[]string{"Alice (26 years)"},
		map[string]string{"2026-04-01": "Sunny"},
		"08 clear, 12 warm",
	)

	if userInfo["Weather"] != "Sunny" {
		t.Fatalf("Weather = %q, want Sunny", userInfo["Weather"])
	}
	if userInfo["HourlyForecastToday"] != "08 clear, 12 warm" {
		t.Fatalf("HourlyForecastToday = %q", userInfo["HourlyForecastToday"])
	}
	if userInfo["Birthdays"] != "Alice (26 years)" {
		t.Fatalf("Birthdays = %q", userInfo["Birthdays"])
	}
	if userInfo["OngoingEvents"] != "Breakfast (until 08:00)" {
		t.Fatalf("OngoingEvents = %q", userInfo["OngoingEvents"])
	}
	if userInfo["Family"] != "Alice, Bob" {
		t.Fatalf("Family = %q", userInfo["Family"])
	}

	missingWeather := g.assembleUserInfo(now, []string{"Alice"}, nil, nil, map[string]string{}, "")
	if missingWeather["Weather"] != "Weather information not available" {
		t.Fatalf("missing weather fallback = %q", missingWeather["Weather"])
	}
	if _, ok := missingWeather["HourlyForecastToday"]; ok {
		t.Fatal("HourlyForecastToday present for empty hourly forecast")
	}
	if _, ok := missingWeather["Birthdays"]; ok {
		t.Fatal("Birthdays present when no birthdays provided")
	}
	if _, ok := missingWeather["OngoingEvents"]; ok {
		t.Fatal("OngoingEvents present when no events provided")
	}
}

func TestFormatCalendarEventString(t *testing.T) {
	location := "Office"
	description := "Quarterly planning"
	start := time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC)
	sameDayEnd := time.Date(2026, 4, 1, 10, 30, 0, 0, time.UTC)
	multiDayEnd := time.Date(2026, 4, 2, 8, 15, 0, 0, time.UTC)

	tests := []struct {
		name  string
		event store.CalendarEvent
		want  string
	}{
		{
			name:  "no end time",
			event: store.CalendarEvent{Summary: "Coffee", StartTime: start},
			want:  "Calendar Event: Coffee at 2026-04-01 09:00",
		},
		{
			name:  "same day with metadata",
			event: store.CalendarEvent{Summary: "Planning", StartTime: start, EndTime: &sameDayEnd, Location: &location, Description: &description},
			want:  "Calendar Event: Planning from 2026-04-01 09:00 to 10:30 at Office. Description: Quarterly planning",
		},
		{
			name:  "multi day end time includes date",
			event: store.CalendarEvent{Summary: "Trip", StartTime: start, EndTime: &multiDayEnd},
			want:  "Calendar Event: Trip from 2026-04-01 09:00 to 2026-04-02 08:15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatCalendarEventString(tt.event); got != tt.want {
				t.Fatalf("formatCalendarEventString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetRelevantMemoryStringsSkipsWeatherMemories(t *testing.T) {
	g, s := newBriefGenerator(t)
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	if _, err := s.AddMemory("Manual note", &date, "manual", nil); err != nil {
		t.Fatalf("AddMemory() manual error = %v", err)
	}
	if _, err := s.AddMemory("Weather note", &date, "weather-metno:Helsinki", nil); err != nil {
		t.Fatalf("AddMemory() weather error = %v", err)
	}
	if _, err := s.AddMemory("General note", nil, "manual", nil); err != nil {
		t.Fatalf("AddMemory() general error = %v", err)
	}

	formatted, memories, err := g.getRelevantMemoryStrings(date.Add(-time.Hour), date.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("getRelevantMemoryStrings() error = %v", err)
	}

	if len(memories) != 3 {
		t.Fatalf("len(memories) = %d, want 3", len(memories))
	}
	if len(formatted) != 2 {
		t.Fatalf("len(formatted) = %d, want 2", len(formatted))
	}
	if strings.Contains(strings.Join(formatted, "\n"), "Weather note") {
		t.Fatalf("formatted memories unexpectedly included weather memory: %#v", formatted)
	}
	if !strings.Contains(formatted[0], "Manual note (relevant on 2026-04-01) [Source: manual]") {
		t.Fatalf("first formatted memory = %q", formatted[0])
	}
}
