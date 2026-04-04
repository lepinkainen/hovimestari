package schoollunch

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

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
	timezone   *time.Location
}

// NewImporter creates a new school lunch importer
func NewImporter(store *store.Store, url, schoolName string, tz *time.Location) *Importer {
	if tz == nil {
		tz = time.UTC
	}
	return &Importer{
		store:      store,
		url:        url,
		schoolName: schoolName,
		timezone:   tz,
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

	source := fmt.Sprintf("%s:%s", SourcePrefix, i.schoolName)

	// Process each day in the current week
	for _, day := range currentWeek.Days {
		// Normalize to midnight in the configured timezone for consistent DB comparisons
		d := day.Date.In(i.timezone)
		relevanceDate := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, i.timezone)

		content := formatMealContent(&day)

		// Remove any existing entry for this source+date before inserting
		if err := i.store.DeleteMemoriesBySourceAndDate(source, relevanceDate); err != nil {
			slog.Error("Failed to delete existing school lunch memory", "date", relevanceDate, "error", err)
			continue
		}

		if _, err := i.store.AddMemory(content, &relevanceDate, source, nil); err != nil {
			slog.Error("Failed to add school lunch menu to database", "date", relevanceDate, "error", err)
			continue
		}

		slog.Debug("Added school lunch menu", "date", relevanceDate.Format("2006-01-02"), "school", i.schoolName)
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
