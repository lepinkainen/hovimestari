package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// CalendarEvent represents a calendar event in the database
type CalendarEvent struct {
	ID          int64
	UID         string
	Summary     string
	StartTime   time.Time
	EndTime     *time.Time // Pointer to allow NULL values
	Location    *string    // Pointer to allow NULL values
	Description *string    // Pointer to allow NULL values
	CreatedAt   time.Time
	Source      string // Format: "calendar:calendarName"
}

// Memory represents a single memory entry in the database
type Memory struct {
	ID            int64
	Content       string
	CreatedAt     time.Time
	RelevanceDate *time.Time // Pointer to allow NULL values
	Source        string
	UID           *string // Pointer to allow NULL values, used for unique identification (e.g., calendar event UID)
}

// Store handles database operations
type Store struct {
	db *sql.DB
}

// NewStore creates a new store with the given database path
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// Initialize creates the necessary tables if they don't exist
func (s *Store) Initialize() error {
	// Create memories table
	memoriesQuery := `
	CREATE TABLE IF NOT EXISTS memories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		relevance_date TIMESTAMP,
		source TEXT NOT NULL,
		uid TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_memories_relevance_date ON memories(relevance_date);
	CREATE INDEX IF NOT EXISTS idx_memories_source ON memories(source);
	CREATE INDEX IF NOT EXISTS idx_memories_source_uid ON memories(source, uid);
	`

	_, err := s.db.Exec(memoriesQuery)
	if err != nil {
		return fmt.Errorf("failed to create memories table: %w", err)
	}

	// Create calendar_events table
	calendarEventsQuery := `
	CREATE TABLE IF NOT EXISTS calendar_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uid TEXT NOT NULL,
		summary TEXT NOT NULL,
		start_time TIMESTAMP NOT NULL,
		end_time TIMESTAMP,
		location TEXT,
		description TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		source TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_calendar_events_start_time ON calendar_events(start_time);
	CREATE INDEX IF NOT EXISTS idx_calendar_events_end_time ON calendar_events(end_time);
	CREATE INDEX IF NOT EXISTS idx_calendar_events_source ON calendar_events(source);
	CREATE INDEX IF NOT EXISTS idx_calendar_events_source_uid ON calendar_events(source, uid);
	CREATE INDEX IF NOT EXISTS idx_calendar_events_uid ON calendar_events(uid);
	`

	_, err = s.db.Exec(calendarEventsQuery)
	if err != nil {
		return fmt.Errorf("failed to create calendar_events table: %w", err)
	}

	return nil
}

// AddMemory adds a new memory to the database
func (s *Store) AddMemory(content string, relevanceDate *time.Time, source string, uid *string) (int64, error) {
	query := `
	INSERT INTO memories (content, relevance_date, source, uid)
	VALUES (?, ?, ?, ?)
	`

	result, err := s.db.Exec(query, content, relevanceDate, source, uid)
	if err != nil {
		return 0, fmt.Errorf("failed to add memory: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// GetRelevantMemories retrieves memories relevant for a specific date range
func (s *Store) GetRelevantMemories(startDate, endDate time.Time) ([]Memory, error) {
	query := `
	SELECT id, content, created_at, relevance_date, source, uid
	FROM memories
	WHERE (relevance_date IS NULL OR (relevance_date >= ? AND relevance_date <= ?))
	ORDER BY CASE WHEN relevance_date IS NULL THEN 1 ELSE 0 END, relevance_date ASC
	`

	rows, err := s.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query memories: %w", err)
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var memory Memory
		var relevanceDate sql.NullTime

		var uid sql.NullString
		err := rows.Scan(&memory.ID, &memory.Content, &memory.CreatedAt, &relevanceDate, &memory.Source, &uid)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory row: %w", err)
		}

		if relevanceDate.Valid {
			memory.RelevanceDate = &relevanceDate.Time
		}

		if uid.Valid {
			memory.UID = &uid.String
		}

		memories = append(memories, memory)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating memory rows: %w", err)
	}

	return memories, nil
}

// MemoryExists checks if a memory with the given source, uid, and relevance date already exists
func (s *Store) MemoryExists(source string, uid string, relevanceDate time.Time) (bool, error) {
	query := `
	SELECT COUNT(*)
	FROM memories
	WHERE source = ? AND uid = ? AND relevance_date = ?
	`

	var count int
	err := s.db.QueryRow(query, source, uid, relevanceDate).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if memory exists: %w", err)
	}

	return count > 0, nil
}

// GetMemoriesBySource retrieves memories from a specific source
func (s *Store) GetMemoriesBySource(source string) ([]Memory, error) {
	query := `
	SELECT id, content, created_at, relevance_date, source, uid
	FROM memories
	WHERE source = ?
	ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, source)
	if err != nil {
		return nil, fmt.Errorf("failed to query memories by source: %w", err)
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var memory Memory
		var relevanceDate sql.NullTime

		var uid sql.NullString
		err := rows.Scan(&memory.ID, &memory.Content, &memory.CreatedAt, &relevanceDate, &memory.Source, &uid)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory row: %w", err)
		}

		if relevanceDate.Valid {
			memory.RelevanceDate = &relevanceDate.Time
		}

		if uid.Valid {
			memory.UID = &uid.String
		}

		memories = append(memories, memory)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating memory rows: %w", err)
	}

	return memories, nil
}

// AddCalendarEvent adds a new calendar event to the database
func (s *Store) AddCalendarEvent(uid, summary string, startTime time.Time, endTime *time.Time, location, description *string, source string) (int64, error) {
	query := `
	INSERT INTO calendar_events (uid, summary, start_time, end_time, location, description, source)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(query, uid, summary, startTime, endTime, location, description, source)
	if err != nil {
		return 0, fmt.Errorf("failed to add calendar event: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// CalendarEventExists checks if a calendar event with the given source, uid, and start time already exists
func (s *Store) CalendarEventExists(source string, uid string, startTime time.Time) (bool, error) {
	query := `
	SELECT COUNT(*)
	FROM calendar_events
	WHERE source = ? AND uid = ? AND start_time = ?
	`

	var count int
	err := s.db.QueryRow(query, source, uid, startTime).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if calendar event exists: %w", err)
	}

	return count > 0, nil
}

// UpdateCalendarEvent updates an existing calendar event in the database
func (s *Store) UpdateCalendarEvent(uid, summary string, startTime time.Time, endTime *time.Time, location, description *string, source string) error {
	query := `
	UPDATE calendar_events
	SET summary = ?, end_time = ?, location = ?, description = ?
	WHERE source = ? AND uid = ? AND start_time = ?
	`

	_, err := s.db.Exec(query, summary, endTime, location, description, source, uid, startTime)
	if err != nil {
		return fmt.Errorf("failed to update calendar event: %w", err)
	}

	return nil
}

// DeleteCalendarEventsBySource deletes all calendar events from a specific source
func (s *Store) DeleteCalendarEventsBySource(source string) error {
	query := `DELETE FROM calendar_events WHERE source = ?`

	_, err := s.db.Exec(query, source)
	if err != nil {
		return fmt.Errorf("failed to delete calendar events: %w", err)
	}

	return nil
}

// GetRelevantCalendarEvents retrieves calendar events relevant for a specific date range
func (s *Store) GetRelevantCalendarEvents(startDate, endDate time.Time) ([]CalendarEvent, error) {
	// Get events that:
	// 1. Start within the date range, OR
	// 2. End within the date range, OR
	// 3. Span across the date range (start before and end after)
	query := `
	SELECT id, uid, summary, start_time, end_time, location, description, created_at, source
	FROM calendar_events
	WHERE 
		(start_time >= ? AND start_time <= ?) OR
		(end_time >= ? AND end_time <= ?) OR
		(start_time <= ? AND end_time >= ?)
	ORDER BY start_time ASC
	`

	rows, err := s.db.Query(query, startDate, endDate, startDate, endDate, startDate, startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query calendar events: %w", err)
	}
	defer rows.Close()

	var events []CalendarEvent
	for rows.Next() {
		var event CalendarEvent
		var endTime sql.NullTime
		var location sql.NullString
		var description sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.UID,
			&event.Summary,
			&event.StartTime,
			&endTime,
			&location,
			&description,
			&event.CreatedAt,
			&event.Source,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan calendar event row: %w", err)
		}

		if endTime.Valid {
			event.EndTime = &endTime.Time
		}

		if location.Valid {
			event.Location = &location.String
		}

		if description.Valid {
			event.Description = &description.String
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating calendar event rows: %w", err)
	}

	return events, nil
}

// GetOngoingCalendarEvents retrieves calendar events that are ongoing at the specified time
func (s *Store) GetOngoingCalendarEvents(currentTime time.Time) ([]CalendarEvent, error) {
	query := `
	SELECT id, uid, summary, start_time, end_time, location, description, created_at, source
	FROM calendar_events
	WHERE start_time <= ? AND (end_time IS NULL OR end_time >= ?)
	ORDER BY start_time ASC
	`

	rows, err := s.db.Query(query, currentTime, currentTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query ongoing calendar events: %w", err)
	}
	defer rows.Close()

	var events []CalendarEvent
	for rows.Next() {
		var event CalendarEvent
		var endTime sql.NullTime
		var location sql.NullString
		var description sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.UID,
			&event.Summary,
			&event.StartTime,
			&endTime,
			&location,
			&description,
			&event.CreatedAt,
			&event.Source,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan calendar event row: %w", err)
		}

		if endTime.Valid {
			event.EndTime = &endTime.Time
		}

		if location.Valid {
			event.Location = &location.String
		}

		if description.Valid {
			event.Description = &description.String
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating calendar event rows: %w", err)
	}

	return events, nil
}
