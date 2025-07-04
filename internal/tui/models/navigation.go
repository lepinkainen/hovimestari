package models

import (
	"fmt"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lepinkainen/hovimestari/internal/config"
	"github.com/lepinkainen/hovimestari/internal/store"
)

// View represents different views in the TUI
type View int

const (
	DashboardView View = iota
	MemoryListView
	MemoryDetailView
	MemoryFormView
	ConfirmationView
	HelpView
)

// Navigation represents the main navigation model
type Navigation struct {
	store       *store.Store
	config      *config.Config
	currentView View
	width       int
	height      int
	
	// Sub-models for different views
	memoryList      *MemoryList
	memoryDetail    *MemoryDetail
	memoryForm      *MemoryForm
	confirmation    *ConfirmationDialog
	dashboard       *Dashboard
	
	// State for pending operations
	pendingDeleteID int64
}

// NewNavigation creates a new navigation model
func NewNavigation(store *store.Store, config *config.Config) *Navigation {
	return &Navigation{
		store:       store,
		config:      config,
		currentView: DashboardView,
		memoryList:  NewMemoryList(store),
		memoryForm:  NewMemoryForm(store),
		dashboard:   NewDashboard(store, config),
	}
}

// Init initializes the navigation model
func (m *Navigation) Init() tea.Cmd {
	return tea.Batch(
		m.dashboard.Init(),
		m.memoryList.Init(),
		m.memoryForm.Init(),
	)
}

// Update handles messages for the navigation model
func (m *Navigation) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Update sub-models with new size
		_, cmd = m.dashboard.Update(msg)
		cmds = append(cmds, cmd)
		_, cmd = m.memoryList.Update(msg)
		cmds = append(cmds, cmd)
		_, cmd = m.memoryForm.Update(msg)
		cmds = append(cmds, cmd)
		if m.memoryDetail != nil {
			_, cmd = m.memoryDetail.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.confirmation != nil {
			_, cmd = m.confirmation.Update(msg)
			cmds = append(cmds, cmd)
		}
		
	case ShowMemoryDetailCmd:
		// Find the memory and create detail view
		for _, memory := range m.memoryList.allMemories {
			if memory.ID == msg.MemoryID {
				m.memoryDetail = NewMemoryDetail(&memory)
				m.currentView = MemoryDetailView
				break
			}
		}
		
	case EditMemoryCmd:
		// Find the memory and create edit form
		for _, memory := range m.memoryList.allMemories {
			if memory.ID == msg.MemoryID {
				m.memoryForm = NewMemoryFormForEdit(m.store, &memory)
				m.currentView = MemoryFormView
				break
			}
		}
		
	case DeleteMemoryCmd:
		// Show confirmation dialog for memory deletion
		var memoryContent string
		for _, memory := range m.memoryList.allMemories {
			if memory.ID == msg.MemoryID {
				memoryContent = memory.Content
				if len(memoryContent) > 50 {
					memoryContent = memoryContent[:50] + "..."
				}
				break
			}
		}
		
		m.pendingDeleteID = msg.MemoryID
		m.confirmation = NewConfirmationDialog(
			"Delete Memory",
			fmt.Sprintf("Are you sure you want to delete this memory?\n\n\"%s\"", memoryContent),
		)
		m.currentView = ConfirmationView
		
	case MemoryFormSavedMsg:
		// Handle memory form save completion
		if msg.Err == nil {
			// Memory was saved successfully, refresh memory list
			cmd = m.memoryList.fetchMemories()
			cmds = append(cmds, cmd)
			// Stay in form view to show success message
		}
		// Forward message to form for handling
		_, cmd = m.memoryForm.Update(msg)
		cmds = append(cmds, cmd)
		
	case ConfirmationResult:
		// Handle confirmation dialog result
		if msg.Confirmed && m.pendingDeleteID != 0 {
			// User confirmed deletion, delete the memory
			err := m.store.DeleteMemory(m.pendingDeleteID)
			if err == nil {
				// Memory deleted successfully, refresh list
				cmd = m.memoryList.fetchMemories()
				cmds = append(cmds, cmd)
			}
			// TODO: Show error message if deletion failed
		}
		
		// Reset state and go back to memory list
		m.pendingDeleteID = 0
		m.currentView = MemoryListView
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "1":
			m.currentView = DashboardView
		case "2":
			m.currentView = MemoryListView
		case "3":
			m.currentView = MemoryFormView
			m.memoryForm.Reset() // Reset form when entering
		case "h", "?":
			m.currentView = HelpView
		case "esc":
			// Go back from detail, form, or confirmation view
			switch m.currentView {
			case MemoryDetailView, MemoryFormView:
				m.currentView = MemoryListView
			case ConfirmationView:
				// Cancel confirmation and go back
				m.pendingDeleteID = 0
				m.currentView = MemoryListView
			}
		}
	}

	// Update the current view
	switch m.currentView {
	case DashboardView:
		_, cmd = m.dashboard.Update(msg)
		cmds = append(cmds, cmd)
	case MemoryListView:
		_, cmd = m.memoryList.Update(msg)
		cmds = append(cmds, cmd)
	case MemoryFormView:
		_, cmd = m.memoryForm.Update(msg)
		cmds = append(cmds, cmd)
	case ConfirmationView:
		if m.confirmation != nil {
			_, cmd = m.confirmation.Update(msg)
			cmds = append(cmds, cmd)
		}
	case MemoryDetailView:
		if m.memoryDetail != nil {
			_, cmd = m.memoryDetail.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the navigation model
func (m *Navigation) View() string {
	var content string
	
	// Navigation bar
	navBar := m.renderNavBar()
	
	// Content based on current view
	switch m.currentView {
	case DashboardView:
		content = m.dashboard.View()
	case MemoryListView:
		content = m.memoryList.View()
	case MemoryFormView:
		content = m.memoryForm.View()
	case ConfirmationView:
		if m.confirmation != nil {
			content = m.confirmation.View()
		} else {
			content = "No confirmation dialog"
		}
	case MemoryDetailView:
		if m.memoryDetail != nil {
			content = m.memoryDetail.View()
		} else {
			content = "No memory selected"
		}
	case HelpView:
		content = m.renderHelp()
	default:
		content = "Unknown view"
	}
	
	// Status bar
	statusBar := m.renderStatusBar()
	
	return navBar + "\n" + content + "\n" + statusBar
}

// renderNavBar renders the navigation bar
func (m *Navigation) renderNavBar() string {
	var tabs []string
	
	dashboardStyle := lipgloss.NewStyle().Padding(0, 1)
	memoryListStyle := lipgloss.NewStyle().Padding(0, 1)
	memoryFormStyle := lipgloss.NewStyle().Padding(0, 1)
	
	if m.currentView == DashboardView {
		dashboardStyle = dashboardStyle.Background(lipgloss.Color("205")).Foreground(lipgloss.Color("0"))
	}
	if m.currentView == MemoryListView {
		memoryListStyle = memoryListStyle.Background(lipgloss.Color("205")).Foreground(lipgloss.Color("0"))
	}
	if m.currentView == MemoryFormView {
		memoryFormStyle = memoryFormStyle.Background(lipgloss.Color("205")).Foreground(lipgloss.Color("0"))
	}
	
	tabs = append(tabs, dashboardStyle.Render("1: Dashboard"))
	tabs = append(tabs, memoryListStyle.Render("2: Memories"))
	tabs = append(tabs, memoryFormStyle.Render("3: Add Memory"))
	tabs = append(tabs, lipgloss.NewStyle().Padding(0, 1).Render("h: Help"))
	tabs = append(tabs, lipgloss.NewStyle().Padding(0, 1).Render("q: Quit"))
	
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

// renderStatusBar renders the status bar
func (m *Navigation) renderStatusBar() string {
	status := "Hovimestari TUI"
	statusStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("255")).
		Padding(0, 1)
	
	return statusStyle.Width(m.width).Render(status)
}

// renderHelp renders the help view
func (m *Navigation) renderHelp() string {
	help := `
Hovimestari Terminal UI Help

Navigation:
  1          - Dashboard view
  2          - Memory list view
  3          - Add new memory
  h, ?       - Show this help
  q, Ctrl+C  - Quit

Memory List:
  ↑/↓        - Navigate memories
  Enter      - View memory details
  e          - Edit selected memory
  x/Delete   - Delete selected memory (with confirmation)
  f          - Filter by source (cycles through: all, manual, calendar, weather)
  d          - Filter by date (cycles through: all, today, this week, this month, this year)
  c          - Clear all filters
  r          - Refresh memories
  /          - Search memories (built-in)

Add Memory Form:
  Tab        - Move to next field
  Shift+Tab  - Move to previous field
  Ctrl+S     - Save memory
  ESC        - Cancel and go back

Dashboard:
  r          - Refresh data

This is a work in progress. More features coming soon!
`
	return help
}