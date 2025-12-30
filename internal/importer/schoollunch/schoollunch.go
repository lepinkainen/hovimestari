package schoollunch

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lepinkainen/hovimestari/internal/store"
	lunch "github.com/lepinkainen/palmia-lunch/lunch"
)

const (
	// SourcePrefix is the prefix used for school lunch memory sources
	SourcePrefix = "schoollunch"
)

// Importer handles importing school lunch menus
type Importer struct {
	store      *store.Store
	url        string
	schoolName string
}

// NewImporter creates a new school lunch importer
func NewImporter(store *store.Store, url, schoolName string) *Importer {
	return &Importer{
		store:      store,
		url:        url,
		schoolName: schoolName,
	}
}

// Import fetches school lunch menus and stores them in the database
func (i *Importer) Import(ctx context.Context) error {
	// Fetch menu from the configured URL or use default
	var menu *lunch.Menu
	var err error

	if i.url != "" {
		menu, err = lunch.FetchFromURL(i.url)
	} else {
		menu, err = lunch.Fetch()
	}

	if err != nil {
		return fmt.Errorf("failed to fetch lunch menu: %w", err)
	}

	// Get current week's menu
	currentWeek := menu.GetCurrentWeek()
	if currentWeek == nil {
		slog.Warn("No current week menu found")
		return nil
	}

	// Process each day in the current week
	for _, day := range currentWeek.Days {
		// Format the day's menu as a memory
		content := formatMealContent(&day)

		// Use the day's date as the relevance date
		relevanceDate := day.Date

		// Add the memory to the database with the school lunch source
		source := fmt.Sprintf("%s:%s", SourcePrefix, i.schoolName)
		_, err := i.store.AddMemory(content, &relevanceDate, source, nil)
		if err != nil {
			slog.Error("Failed to add school lunch menu to database", "date", day.Date, "error", err)
			continue
		}

		slog.Debug("Added school lunch menu", "date", day.Date.Format("2006-01-02"), "school", i.schoolName)
	}

	return nil
}

// formatMealContent formats a day's lunch menu as a string
func formatMealContent(day *lunch.Day) string {
	var sb strings.Builder

	// Format lunch
	if day.Lunch.Name != "" {
		sb.WriteString("Lounas: ")
		sb.WriteString(day.Lunch.Name)
		if len(day.Lunch.Allergens) > 0 {
			sb.WriteString(" (")
			sb.WriteString(strings.Join(day.Lunch.Allergens, ","))
			sb.WriteString(")")
		}
		sb.WriteString("\n")

		if len(day.Lunch.Components) > 0 {
			sb.WriteString("Osat: ")
			sb.WriteString(strings.Join(day.Lunch.Components, ", "))
			sb.WriteString("\n")
		}
	}

	// Format vegetarian lunch
	if day.Vegetarian.Name != "" {
		sb.WriteString("Kasvislounas: ")
		sb.WriteString(day.Vegetarian.Name)
		if len(day.Vegetarian.Allergens) > 0 {
			sb.WriteString(" (")
			sb.WriteString(strings.Join(day.Vegetarian.Allergens, ","))
			sb.WriteString(")")
		}
		sb.WriteString("\n")

		if len(day.Vegetarian.Components) > 0 {
			sb.WriteString("Osat: ")
			sb.WriteString(strings.Join(day.Vegetarian.Components, ", "))
			sb.WriteString("\n")
		}
	}

	// Trim trailing newline
	result := strings.TrimSuffix(sb.String(), "\n")
	return result
}
