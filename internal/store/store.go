package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Memory represents a single memory entry in the database
type Memory struct {
	ID            int64
	Content       string
	CreatedAt     time.Time
	RelevanceDate *time.Time // Pointer to allow NULL values
	Source        string
}

// Store handles database operations
type Store struct {
	db *sql.DB
}

// NewStore creates a new store with the given database path
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
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
	query := `
	CREATE TABLE IF NOT EXISTS memories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		relevance_date TIMESTAMP,
		source TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_memories_relevance_date ON memories(relevance_date);
	CREATE INDEX IF NOT EXISTS idx_memories_source ON memories(source);
	`

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// AddMemory adds a new memory to the database
func (s *Store) AddMemory(content string, relevanceDate *time.Time, source string) (int64, error) {
	query := `
	INSERT INTO memories (content, relevance_date, source)
	VALUES (?, ?, ?)
	`

	result, err := s.db.Exec(query, content, relevanceDate, source)
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
	SELECT id, content, created_at, relevance_date, source
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

		err := rows.Scan(&memory.ID, &memory.Content, &memory.CreatedAt, &relevanceDate, &memory.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory row: %w", err)
		}

		if relevanceDate.Valid {
			memory.RelevanceDate = &relevanceDate.Time
		}

		memories = append(memories, memory)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating memory rows: %w", err)
	}

	return memories, nil
}

// GetMemoriesBySource retrieves memories from a specific source
func (s *Store) GetMemoriesBySource(source string) ([]Memory, error) {
	query := `
	SELECT id, content, created_at, relevance_date, source
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

		err := rows.Scan(&memory.ID, &memory.Content, &memory.CreatedAt, &relevanceDate, &memory.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory row: %w", err)
		}

		if relevanceDate.Valid {
			memory.RelevanceDate = &relevanceDate.Time
		}

		memories = append(memories, memory)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating memory rows: %w", err)
	}

	return memories, nil
}
