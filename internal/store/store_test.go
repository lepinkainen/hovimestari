package store

import (
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()

	s, err := NewStore(filepath.Join(t.TempDir(), "test.db"))
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

func strPtr(s string) *string { return &s }

func TestMemoryLifecycle(t *testing.T) {
	s := newTestStore(t)

	inRange := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	outOfRange := inRange.AddDate(0, 0, 5)
	uid := "uid-1"

	if _, err := s.AddMemory("dated memory", &inRange, "calendar:home", &uid); err != nil {
		t.Fatalf("AddMemory() dated error = %v", err)
	}
	if _, err := s.AddMemory("always relevant", nil, "manual", nil); err != nil {
		t.Fatalf("AddMemory() nil relevance error = %v", err)
	}
	if _, err := s.AddMemory("future memory", &outOfRange, "calendar:home", nil); err != nil {
		t.Fatalf("AddMemory() future error = %v", err)
	}

	exists, err := s.MemoryExists("calendar:home", uid, inRange)
	if err != nil {
		t.Fatalf("MemoryExists() error = %v", err)
	}
	if !exists {
		t.Fatal("MemoryExists() = false, want true")
	}

	missing, err := s.MemoryExists("calendar:home", "missing", inRange)
	if err != nil {
		t.Fatalf("MemoryExists() missing error = %v", err)
	}
	if missing {
		t.Fatal("MemoryExists() = true for missing uid, want false")
	}

	memories, err := s.GetRelevantMemories(inRange.Add(-time.Hour), inRange.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("GetRelevantMemories() error = %v", err)
	}
	if len(memories) != 2 {
		t.Fatalf("len(GetRelevantMemories()) = %d, want 2", len(memories))
	}
	if memories[0].Content != "dated memory" {
		t.Fatalf("first relevant memory = %q, want dated memory", memories[0].Content)
	}
	if memories[0].UID == nil || *memories[0].UID != uid {
		t.Fatalf("first memory UID = %v, want %q", memories[0].UID, uid)
	}
	if memories[1].Content != "always relevant" {
		t.Fatalf("second relevant memory = %q, want always relevant", memories[1].Content)
	}

	bySource, err := s.GetMemoriesBySource("calendar:home")
	if err != nil {
		t.Fatalf("GetMemoriesBySource() error = %v", err)
	}
	if len(bySource) != 2 {
		t.Fatalf("len(GetMemoriesBySource()) = %d, want 2", len(bySource))
	}
}

func TestDeleteMemoriesBySourceAndDate(t *testing.T) {
	s := newTestStore(t)

	date := time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC)
	otherDate := date.AddDate(0, 0, 1)

	if _, err := s.AddMemory("delete me", &date, "electricity:fi", nil); err != nil {
		t.Fatalf("AddMemory() error = %v", err)
	}
	if _, err := s.AddMemory("keep different date", &otherDate, "electricity:fi", nil); err != nil {
		t.Fatalf("AddMemory() error = %v", err)
	}
	if _, err := s.AddMemory("keep different source", &date, "manual", nil); err != nil {
		t.Fatalf("AddMemory() error = %v", err)
	}

	if err := s.DeleteMemoriesBySourceAndDate("electricity:fi", date); err != nil {
		t.Fatalf("DeleteMemoriesBySourceAndDate() error = %v", err)
	}

	memories, err := s.GetRelevantMemories(date.Add(-time.Hour), otherDate.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("GetRelevantMemories() error = %v", err)
	}
	if len(memories) != 2 {
		t.Fatalf("len(GetRelevantMemories()) after delete = %d, want 2", len(memories))
	}
	for _, memory := range memories {
		if memory.Content == "delete me" {
			t.Fatal("deleted memory still present")
		}
	}
}

func TestCalendarEventLifecycleAndQueries(t *testing.T) {
	s := newTestStore(t)

	rangeStart := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)
	rangeEnd := rangeStart.Add(24 * time.Hour)

	spanningStart := rangeStart.Add(-2 * time.Hour)
	spanningEnd := rangeStart.Add(2 * time.Hour)
	withinStart := rangeStart.Add(10 * time.Hour)
	withinEnd := rangeStart.Add(11 * time.Hour)
	outsideStart := rangeEnd.Add(24 * time.Hour)
	outsideEnd := outsideStart.Add(time.Hour)
	ongoingNoEndStart := rangeStart.Add(30 * time.Minute)

	if _, err := s.AddCalendarEvent("uid-span", "Overnight event", spanningStart, &spanningEnd, strPtr("Home"), strPtr("Spans range"), "calendar:home"); err != nil {
		t.Fatalf("AddCalendarEvent() spanning error = %v", err)
	}
	if _, err := s.AddCalendarEvent("uid-within", "Meeting", withinStart, &withinEnd, nil, nil, "calendar:home"); err != nil {
		t.Fatalf("AddCalendarEvent() within error = %v", err)
	}
	if _, err := s.AddCalendarEvent("uid-open", "Open ended", ongoingNoEndStart, nil, nil, nil, "calendar:work"); err != nil {
		t.Fatalf("AddCalendarEvent() open-ended error = %v", err)
	}
	if _, err := s.AddCalendarEvent("uid-outside", "Outside", outsideStart, &outsideEnd, nil, nil, "calendar:home"); err != nil {
		t.Fatalf("AddCalendarEvent() outside error = %v", err)
	}

	exists, err := s.CalendarEventExists("calendar:home", "uid-within", withinStart)
	if err != nil {
		t.Fatalf("CalendarEventExists() error = %v", err)
	}
	if !exists {
		t.Fatal("CalendarEventExists() = false, want true")
	}

	updatedEnd := withinEnd.Add(time.Hour)
	if err := s.UpdateCalendarEvent("uid-within", "Updated meeting", withinStart, &updatedEnd, strPtr("Office"), strPtr("Updated description"), "calendar:home"); err != nil {
		t.Fatalf("UpdateCalendarEvent() error = %v", err)
	}

	relevant, err := s.GetRelevantCalendarEvents(rangeStart, rangeEnd)
	if err != nil {
		t.Fatalf("GetRelevantCalendarEvents() error = %v", err)
	}
	if len(relevant) != 3 {
		t.Fatalf("len(GetRelevantCalendarEvents()) = %d, want 3", len(relevant))
	}
	if relevant[0].UID != "uid-span" || relevant[1].UID != "uid-open" || relevant[2].UID != "uid-within" {
		t.Fatalf("unexpected relevant event order: %#v", []string{relevant[0].UID, relevant[1].UID, relevant[2].UID})
	}
	if relevant[2].Summary != "Updated meeting" {
		t.Fatalf("updated summary = %q, want %q", relevant[2].Summary, "Updated meeting")
	}
	if relevant[2].Location == nil || *relevant[2].Location != "Office" {
		t.Fatalf("updated location = %v, want Office", relevant[2].Location)
	}
	if relevant[2].Description == nil || *relevant[2].Description != "Updated description" {
		t.Fatalf("updated description = %v, want Updated description", relevant[2].Description)
	}
	if relevant[2].EndTime == nil || !relevant[2].EndTime.Equal(updatedEnd) {
		t.Fatalf("updated end time = %v, want %v", relevant[2].EndTime, updatedEnd)
	}

	ongoing, err := s.GetOngoingCalendarEvents(rangeStart.Add(30 * time.Minute))
	if err != nil {
		t.Fatalf("GetOngoingCalendarEvents() error = %v", err)
	}
	if len(ongoing) != 2 {
		t.Fatalf("len(GetOngoingCalendarEvents()) = %d, want 2", len(ongoing))
	}
	if ongoing[0].UID != "uid-span" || ongoing[1].UID != "uid-open" {
		t.Fatalf("unexpected ongoing events: %#v", []string{ongoing[0].UID, ongoing[1].UID})
	}

	if err := s.DeleteCalendarEventsBySource("calendar:home"); err != nil {
		t.Fatalf("DeleteCalendarEventsBySource() error = %v", err)
	}

	remaining, err := s.GetRelevantCalendarEvents(rangeStart.Add(-24*time.Hour), outsideEnd.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("GetRelevantCalendarEvents() after delete error = %v", err)
	}
	if len(remaining) != 1 || remaining[0].UID != "uid-open" {
		t.Fatalf("remaining events = %#v, want only uid-open", remaining)
	}
}
