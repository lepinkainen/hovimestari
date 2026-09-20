package calendar

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/lepinkainen/hovimestari/internal/store"
)

func newImportTestStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.NewStore(filepath.Join(t.TempDir(), "calendar.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := s.Initialize(); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestImport(t *testing.T) {
	for _, mode := range []string{"full_refresh", "smart"} {
		t.Run(mode, func(t *testing.T) {
			s := newImportTestStore(t)
			// gocal's default parsing window is relative to today.
			start := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
			end := start.Add(time.Hour)
			for _, event := range []struct{ uid, source string }{
				{"meeting", "calendar:Home"},
				{"removed-from-feed", "calendar:Home"},
				{"meeting", "calendar:Work"},
			} {
				if _, err := s.AddCalendarEvent(event.uid, "Old summary", start, &end, nil, nil, event.source); err != nil {
					t.Fatal(err)
				}
			}
			description := strings.Repeat("a", 1100)
			ics := fmt.Sprintf("BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//Hovimestari//Test//EN\r\nBEGIN:VEVENT\r\nDTSTAMP:20260101T000000Z\r\nUID:meeting\r\nDTSTART:%s\r\nDTEND:%s\r\nSUMMARY:Updated meeting\r\nLOCATION:Office\r\nDESCRIPTION:%s\r\nEND:VEVENT\r\nBEGIN:VEVENT\r\nDTSTAMP:20260101T000000Z\r\nUID:new-event\r\nDTSTART:%s\r\nDTEND:%s\r\nSUMMARY:New event\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n",
				start.Format("20060102T150405Z"), end.Format("20060102T150405Z"), description,
				start.Format("20060102T150405Z"), end.Format("20060102T150405Z"))
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", r.Method)
				}
				w.Header().Set("Content-Type", "text/calendar")
				_, _ = w.Write([]byte(ics))
			}))
			defer server.Close()
			importer := NewImporter(s, server.URL, "Home", mode)
			// Reimporting the same feed must not duplicate events in either mode.
			for range 2 {
				if err := importer.Import(context.Background()); err != nil {
					t.Fatal(err)
				}
			}
			events, err := s.GetRelevantCalendarEvents(start.Add(-time.Hour), end.Add(time.Hour))
			if err != nil {
				t.Fatal(err)
			}
			wantCount := 3
			if mode == "smart" {
				wantCount = 4 // Smart imports retain events no longer present in the feed.
			}
			if len(events) != wantCount {
				t.Fatalf("stored %d events, want %d: %+v", len(events), wantCount, events)
			}
			seen := make(map[string]bool)
			for _, event := range events {
				key := event.Source + "/" + event.UID
				if seen[key] {
					t.Fatalf("duplicate event: %s", key)
				}
				seen[key] = true
				if event.Source == "calendar:Work" {
					if event.Summary != "Old summary" {
						t.Error("import changed another calendar's event")
					}
					continue
				}
				if event.UID == "meeting" {
					if event.Summary != "Updated meeting" || !event.StartTime.Equal(start) || event.EndTime == nil || !event.EndTime.Equal(end) {
						t.Errorf("unexpected imported event: %+v", event)
					}
					if event.Location == nil || *event.Location != "Office" {
						t.Errorf("location = %v, want Office", event.Location)
					}
					if event.Description == nil || *event.Description != strings.Repeat("a", 997)+"..." {
						t.Error("description was not truncated to 1000 bytes")
					}
				}
			}
			for _, key := range []string{"calendar:Home/meeting", "calendar:Home/new-event", "calendar:Work/meeting"} {
				if !seen[key] {
					t.Errorf("missing event %s", key)
				}
			}
			if seen["calendar:Home/removed-from-feed"] != (mode == "smart") {
				t.Error("incorrect handling of event removed from feed")
			}
		})
	}
}

func TestImportHTTPErrorPreservesEvents(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusInternalServerError} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			s := newImportTestStore(t)
			start := time.Now().UTC().Truncate(time.Second)
			if _, err := s.AddCalendarEvent("existing", "Keep me", start, nil, nil, nil, "calendar:Home"); err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(status)
			}))
			defer server.Close()
			err := NewImporter(s, server.URL, "Home", "full_refresh").Import(context.Background())
			if err == nil || !strings.Contains(err.Error(), fmt.Sprint(status)) {
				t.Fatalf("error = %v, want HTTP status %d", err, status)
			}
			events, err := s.GetRelevantCalendarEvents(start.Add(-time.Hour), start.Add(time.Hour))
			if err != nil || len(events) != 1 || events[0].UID != "existing" {
				t.Fatalf("failed fetch changed stored events: %+v, error %v", events, err)
			}
		})
	}
}

// Events far beyond gocal's default 90-day window must survive a full_refresh,
// and multi-byte descriptions must not be truncated into invalid UTF-8.
func TestImportFarFutureEventAndUTF8Truncation(t *testing.T) {
	s := newImportTestStore(t)
	start := time.Now().UTC().AddDate(0, 0, 120).Truncate(time.Second)
	end := start.Add(time.Hour)
	description := strings.Repeat("ä", 600) // 1200 bytes, truncation cuts mid-rune
	ics := fmt.Sprintf("BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//Hovimestari//Test//EN\r\nBEGIN:VEVENT\r\nDTSTAMP:20260101T000000Z\r\nUID:far-future\r\nDTSTART:%s\r\nDTEND:%s\r\nSUMMARY:Birthday\r\nDESCRIPTION:%s\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n",
		start.Format("20060102T150405Z"), end.Format("20060102T150405Z"), description)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/calendar")
		_, _ = w.Write([]byte(ics))
	}))
	defer server.Close()
	if err := NewImporter(s, server.URL, "Home", "full_refresh").Import(context.Background()); err != nil {
		t.Fatal(err)
	}
	events, err := s.GetRelevantCalendarEvents(start.Add(-time.Hour), end.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].UID != "far-future" {
		t.Fatalf("far future event was dropped: %+v", events)
	}
	if events[0].Description == nil {
		t.Fatal("description missing")
	}
	got := *events[0].Description
	if !utf8.ValidString(got) {
		t.Errorf("truncated description is not valid UTF-8: %q", got)
	}
	if want := strings.Repeat("ä", 498) + "..."; got != want {
		t.Errorf("description = %d bytes, want %d", len(got), len(want))
	}
}
