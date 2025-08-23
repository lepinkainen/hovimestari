package models

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lepinkainen/hovimestari/internal/config"
	"github.com/lepinkainen/hovimestari/internal/store"
)

// Stats represents database statistics
type Stats struct {
	TotalMemories    int
	CalendarMemories int
	WeatherMemories  int
	ManualMemories   int
}

// Dashboard represents the dashboard model
type Dashboard struct {
	store      *store.Store
	config     *config.Config
	width      int
	height     int
	stats      *Stats
	loading    bool
}

// NewDashboard creates a new dashboard model
func NewDashboard(store *store.Store, config *config.Config) *Dashboard {
	return &Dashboard{
		store:   store,
		config:  config,
		loading: true,
	}
}

// StatsMsg represents a message containing database stats
type StatsMsg struct {
	Stats *Stats
	Err   error
}

// Init initializes the dashboard
func (m *Dashboard) Init() tea.Cmd {
	return m.fetchStats()
}

// Update handles messages for the dashboard
func (m *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case StatsMsg:
		m.loading = false
		if msg.Err != nil {
			// Handle error
			return m, nil
		}
		m.stats = msg.Stats
		
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			m.loading = true
			return m, m.fetchStats()
		}
	}
	
	return m, nil
}

// View renders the dashboard
func (m *Dashboard) View() string {
	if m.loading {
		return "Loading dashboard..."
	}
	
	if m.stats == nil {
		return "Failed to load dashboard data"
	}
	
	// Create dashboard sections
	sections := []string{
		m.renderStatsSection(),
		m.renderRecentMemoriesSection(),
		m.renderSystemInfoSection(),
	}
	
	// Join sections vertically
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderStatsSection renders the statistics section
func (m *Dashboard) renderStatsSection() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("📊 Database Statistics")
	
	stats := fmt.Sprintf(
		"Total Memories: %d\n"+
		"Calendar Events: %d\n"+
		"Weather Records: %d\n"+
		"Manual Memories: %d",
		m.stats.TotalMemories,
		m.stats.CalendarMemories,
		m.stats.WeatherMemories,
		m.stats.ManualMemories,
	)
	
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1).
		MarginBottom(1).
		Render(stats)
	
	return title + "\n" + box
}

// renderRecentMemoriesSection renders the recent memories section
func (m *Dashboard) renderRecentMemoriesSection() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("📝 Recent Memories")
	
	// For now, just show a placeholder
	// TODO: Implement recent memories fetching
	content := "Recent memories will be shown here..."
	
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1).
		MarginBottom(1).
		Render(content)
	
	return title + "\n" + box
}

// renderSystemInfoSection renders the system information section
func (m *Dashboard) renderSystemInfoSection() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("⚙️  System Information")
	
	info := fmt.Sprintf(
		"Database: %s\n"+
		"Config: %s\n"+
		"Last Updated: %s",
		m.config.DBPath,
		"Config loaded",
		time.Now().Format("2006-01-02 15:04:05"),
	)
	
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1).
		Render(info)
	
	return title + "\n" + box
}

// fetchStats fetches database statistics
func (m *Dashboard) fetchStats() tea.Cmd {
	return func() tea.Msg {
		// Calculate stats using existing store methods
		stats := &Stats{}
		
		// Get all memories to count them by source
		// Use a large date range to get all memories
		endDate := time.Now()
		startDate := time.Now().AddDate(-10, 0, 0) // Last 10 years to get all
		
		memories, err := m.store.GetRelevantMemories(startDate, endDate)
		if err != nil {
			return StatsMsg{Stats: nil, Err: err}
		}
		
		// Count memories by source
		stats.TotalMemories = len(memories)
		for _, memory := range memories {
			switch {
			case memory.Source == "manual":
				stats.ManualMemories++
			case len(memory.Source) > 8 && memory.Source[:8] == "calendar":
				stats.CalendarMemories++
			case len(memory.Source) > 7 && memory.Source[:7] == "weather":
				stats.WeatherMemories++
			}
		}
		
		return StatsMsg{Stats: stats, Err: nil}
	}
}