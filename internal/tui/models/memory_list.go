package models

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lepinkainen/hovimestari/internal/store"
)

// MemoryItem represents a memory item for the list
type MemoryItem struct {
	ID            int64
	Content       string
	Source        string
	RelevanceDate *time.Time
	CreatedAt     time.Time
}

// FilterValue returns the value used for filtering
func (i MemoryItem) FilterValue() string {
	return i.Content
}

// Title returns the title for the list item
func (i MemoryItem) Title() string {
	return fmt.Sprintf("[%s] %s", i.Source, truncateString(i.Content, 60))
}

// Description returns the description for the list item
func (i MemoryItem) Description() string {
	dateStr := "No date"
	if i.RelevanceDate != nil {
		dateStr = i.RelevanceDate.Format("2006-01-02")
	}
	return fmt.Sprintf("Created: %s | Relevant: %s", 
		i.CreatedAt.Format("2006-01-02 15:04"), dateStr)
}

// DateFilter represents different date filter options
type DateFilter int

const (
	AllDates DateFilter = iota
	Today
	ThisWeek
	ThisMonth
	ThisYear
)

// MemoryList represents the memory list model
type MemoryList struct {
	store        *store.Store
	list         list.Model
	loading      bool
	width        int
	height       int
	sourceFilter string
	dateFilter   DateFilter
	allMemories  []store.Memory
}

// MemoriesMsg represents a message containing memories
type MemoriesMsg struct {
	Memories []store.Memory
	Err      error
}

// ShowMemoryDetailCmd represents a command to show memory detail
type ShowMemoryDetailCmd struct {
	MemoryID int64
}

// EditMemoryCmd represents a command to edit a memory
type EditMemoryCmd struct {
	MemoryID int64
}

// DeleteMemoryCmd represents a command to delete a memory
type DeleteMemoryCmd struct {
	MemoryID int64
}

// NewMemoryList creates a new memory list model
func NewMemoryList(store *store.Store) *MemoryList {
	// Create the list with default delegate
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Memories"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle()
	l.Styles.PaginationStyle = paginationStyle()
	l.Styles.HelpStyle = helpStyle()
	
	return &MemoryList{
		store:   store,
		list:    l,
		loading: true,
	}
}

// Init initializes the memory list
func (m *MemoryList) Init() tea.Cmd {
	return m.fetchMemories()
}

// Update handles messages for the memory list
func (m *MemoryList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4) // Account for navigation and status bars
		
	case MemoriesMsg:
		m.loading = false
		if msg.Err != nil {
			// Handle error
			return m, nil
		}
		
		// Store all memories for filtering
		m.allMemories = msg.Memories
		
		// Apply current filter and update list
		return m, m.applyFilter()
		
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			m.loading = true
			return m, m.fetchMemories()
		case "f":
			// Cycle through source filters
			return m, m.cycleSourceFilter()
		case "d":
			// Cycle through date filters
			return m, m.cycleDateFilter()
		case "c":
			// Clear all filters
			m.sourceFilter = ""
			m.dateFilter = AllDates
			return m, m.applyFilter()
		case "enter":
			// Show memory detail if item is selected
			if selectedItem := m.list.SelectedItem(); selectedItem != nil {
				if memoryItem, ok := selectedItem.(MemoryItem); ok {
					return m, func() tea.Msg {
						return ShowMemoryDetailCmd{MemoryID: memoryItem.ID}
					}
				}
			}
		case "e":
			// Edit memory if item is selected
			if selectedItem := m.list.SelectedItem(); selectedItem != nil {
				if memoryItem, ok := selectedItem.(MemoryItem); ok {
					return m, func() tea.Msg {
						return EditMemoryCmd{MemoryID: memoryItem.ID}
					}
				}
			}
		case "delete", "x":
			// Delete memory if item is selected
			if selectedItem := m.list.SelectedItem(); selectedItem != nil {
				if memoryItem, ok := selectedItem.(MemoryItem); ok {
					return m, func() tea.Msg {
						return DeleteMemoryCmd{MemoryID: memoryItem.ID}
					}
				}
			}
		}
	}
	
	// Update the list
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the memory list
func (m *MemoryList) View() string {
	if m.loading {
		return "Loading memories..."
	}
	
	return m.list.View()
}

// fetchMemories fetches memories from the store
func (m *MemoryList) fetchMemories() tea.Cmd {
	return func() tea.Msg {
		// Get memories for the current date (as a fallback for recent memories)
		// We'll use a large date range to get all memories
		endDate := time.Now()
		startDate := time.Now().AddDate(-1, 0, 0) // Last year
		memories, err := m.store.GetRelevantMemories(startDate, endDate)
		return MemoriesMsg{Memories: memories, Err: err}
	}
}

// Helper functions for styling
func titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)
}

func paginationStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
}

func helpStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
}

// applyFilter applies the current source and date filters to memories
func (m *MemoryList) applyFilter() tea.Cmd {
	var filteredMemories []store.Memory
	
	// Apply filters
	for _, memory := range m.allMemories {
		// Source filter
		if m.sourceFilter != "" {
			if memory.Source != m.sourceFilter && 
				(len(memory.Source) < len(m.sourceFilter) || 
				 memory.Source[:len(m.sourceFilter)] != m.sourceFilter) {
				continue
			}
		}
		
		// Date filter
		if !m.matchesDateFilter(memory) {
			continue
		}
		
		filteredMemories = append(filteredMemories, memory)
	}
	
	// Convert filtered memories to list items
	items := make([]list.Item, len(filteredMemories))
	for i, memory := range filteredMemories {
		items[i] = MemoryItem{
			ID:            memory.ID,
			Content:       memory.Content,
			Source:        memory.Source,
			RelevanceDate: memory.RelevanceDate,
			CreatedAt:     memory.CreatedAt,
		}
	}
	
	// Update list title to show filters
	title := "Memories"
	var filterParts []string
	
	if m.sourceFilter != "" {
		filterParts = append(filterParts, m.sourceFilter)
	}
	
	if m.dateFilter != AllDates {
		filterParts = append(filterParts, m.getDateFilterName())
	}
	
	if len(filterParts) > 0 {
		title = fmt.Sprintf("Memories [%s]", filterParts[0])
		if len(filterParts) > 1 {
			title = fmt.Sprintf("Memories [%s, %s]", filterParts[0], filterParts[1])
		}
	}
	
	m.list.Title = title
	
	return m.list.SetItems(items)
}

// cycleSourceFilter cycles through common source filters
func (m *MemoryList) cycleSourceFilter() tea.Cmd {
	filters := []string{"", "manual", "calendar", "weather"}
	
	// Find current filter index
	currentIndex := 0
	for i, filter := range filters {
		if filter == m.sourceFilter {
			currentIndex = i
			break
		}
	}
	
	// Move to next filter
	nextIndex := (currentIndex + 1) % len(filters)
	m.sourceFilter = filters[nextIndex]
	
	return m.applyFilter()
}

// cycleDateFilter cycles through date filters
func (m *MemoryList) cycleDateFilter() tea.Cmd {
	filters := []DateFilter{AllDates, Today, ThisWeek, ThisMonth, ThisYear}
	
	// Find current filter index
	currentIndex := 0
	for i, filter := range filters {
		if filter == m.dateFilter {
			currentIndex = i
			break
		}
	}
	
	// Move to next filter
	nextIndex := (currentIndex + 1) % len(filters)
	m.dateFilter = filters[nextIndex]
	
	return m.applyFilter()
}

// matchesDateFilter checks if a memory matches the current date filter
func (m *MemoryList) matchesDateFilter(memory store.Memory) bool {
	if m.dateFilter == AllDates {
		return true
	}
	
	now := time.Now()
	var checkDate time.Time
	
	// Use relevance date if available, otherwise use created date
	if memory.RelevanceDate != nil {
		checkDate = *memory.RelevanceDate
	} else {
		checkDate = memory.CreatedAt
	}
	
	switch m.dateFilter {
	case Today:
		return checkDate.Year() == now.Year() &&
			checkDate.YearDay() == now.YearDay()
	case ThisWeek:
		// Get start of this week (Monday)
		weekStart := now.AddDate(0, 0, -int(now.Weekday())+1)
		weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
		weekEnd := weekStart.AddDate(0, 0, 7)
		return checkDate.After(weekStart) && checkDate.Before(weekEnd)
	case ThisMonth:
		return checkDate.Year() == now.Year() &&
			checkDate.Month() == now.Month()
	case ThisYear:
		return checkDate.Year() == now.Year()
	default:
		return true
	}
}

// getDateFilterName returns a human-readable name for the current date filter
func (m *MemoryList) getDateFilterName() string {
	switch m.dateFilter {
	case Today:
		return "today"
	case ThisWeek:
		return "this week"
	case ThisMonth:
		return "this month" 
	case ThisYear:
		return "this year"
	default:
		return "all"
	}
}

// truncateString truncates a string to the specified length
func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}