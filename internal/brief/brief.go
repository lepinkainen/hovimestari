package brief

import (
	"context"
	"fmt"
	"time"

	"github.com/shrike/hovimestari/internal/llm"
	"github.com/shrike/hovimestari/internal/store"
)

// Generator handles generating briefs based on memories
type Generator struct {
	store *store.Store
	llm   *llm.GeminiClient
}

// NewGenerator creates a new brief generator
func NewGenerator(store *store.Store, llm *llm.GeminiClient) *Generator {
	return &Generator{
		store: store,
		llm:   llm,
	}
}

// GenerateDailyBrief generates a daily brief based on memories
func (g *Generator) GenerateDailyBrief(ctx context.Context, daysAhead int) (string, error) {
	// Get the date range for relevant memories
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, daysAhead)

	// Get relevant memories
	memories, err := g.store.GetRelevantMemories(startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to get relevant memories: %w", err)
	}

	// Convert memories to strings
	var memoryStrings []string
	for _, memory := range memories {
		var dateInfo string
		if memory.RelevanceDate != nil {
			dateInfo = fmt.Sprintf(" (relevant on %s)", memory.RelevanceDate.Format("2006-01-02"))
		}
		memoryStrings = append(memoryStrings, fmt.Sprintf("%s%s [Source: %s]", memory.Content, dateInfo, memory.Source))
	}

	// Add user information
	userInfo := map[string]string{
		"Name": "User", // This could be configurable
	}

	// Generate the brief
	brief, err := g.llm.GenerateBrief(ctx, memoryStrings, userInfo)
	if err != nil {
		return "", fmt.Errorf("failed to generate brief: %w", err)
	}

	return brief, nil
}

// GenerateResponse generates a response to a user query
func (g *Generator) GenerateResponse(ctx context.Context, query string) (string, error) {
	// Get all memories (we could be more selective here)
	startDate := time.Now().AddDate(-1, 0, 0) // Look back 1 year
	endDate := time.Now().AddDate(0, 1, 0)    // Look ahead 1 month

	// Get relevant memories
	memories, err := g.store.GetRelevantMemories(startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to get memories: %w", err)
	}

	// Convert memories to strings
	var memoryStrings []string
	for _, memory := range memories {
		var dateInfo string
		if memory.RelevanceDate != nil {
			dateInfo = fmt.Sprintf(" (relevant on %s)", memory.RelevanceDate.Format("2006-01-02"))
		}
		memoryStrings = append(memoryStrings, fmt.Sprintf("%s%s [Source: %s]", memory.Content, dateInfo, memory.Source))
	}

	// Generate the response
	response, err := g.llm.GenerateResponse(ctx, query, memoryStrings)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	return response, nil
}
