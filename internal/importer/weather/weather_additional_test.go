package weather

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/lepinkainen/hovimestari/internal/store"
)

func newWeatherTestStore(t *testing.T) *store.Store {
	t.Helper()

	s, err := store.NewStore(filepath.Join(t.TempDir(), "weather.db"))
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

	return s
}

func TestGetLatestForecasts(t *testing.T) {
	s := newWeatherTestStore(t)

	start := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	end := start.Add(48 * time.Hour)
	dayOne := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	dayTwo := time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC)

	if _, err := s.AddMemory("old forecast", &dayOne, SourcePrefix+":Helsinki", nil); err != nil {
		t.Fatalf("AddMemory() old forecast error = %v", err)
	}
	// SQLite CURRENT_TIMESTAMP is second-resolution; sleep to ensure a newer CreatedAt.
	time.Sleep(1100 * time.Millisecond)
	if _, err := s.AddMemory("new forecast", &dayOne, SourcePrefix+":Helsinki", nil); err != nil {
		t.Fatalf("AddMemory() new forecast error = %v", err)
	}
	if _, err := s.AddMemory("other day", &dayTwo, SourcePrefix+":Helsinki", nil); err != nil {
		t.Fatalf("AddMemory() day two error = %v", err)
	}
	if _, err := s.AddMemory("other location", &dayOne, SourcePrefix+":Espoo", nil); err != nil {
		t.Fatalf("AddMemory() other location error = %v", err)
	}
	if _, err := s.AddMemory("non weather", &dayOne, "manual", nil); err != nil {
		t.Fatalf("AddMemory() non-weather error = %v", err)
	}
	if _, err := s.AddMemory("no relevance date", nil, SourcePrefix+":Helsinki", nil); err != nil {
		t.Fatalf("AddMemory() nil relevance error = %v", err)
	}

	forecasts, err := GetLatestForecasts(s, start, end, "Helsinki")
	if err != nil {
		t.Fatalf("GetLatestForecasts() error = %v", err)
	}

	if len(forecasts) != 2 {
		t.Fatalf("len(GetLatestForecasts()) = %d, want 2", len(forecasts))
	}
	if forecasts["2026-04-01"] != "new forecast" {
		t.Fatalf("forecast for 2026-04-01 = %q, want %q", forecasts["2026-04-01"], "new forecast")
	}
	if forecasts["2026-04-02"] != "other day" {
		t.Fatalf("forecast for 2026-04-02 = %q, want %q", forecasts["2026-04-02"], "other day")
	}
}
