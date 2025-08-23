package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lepinkainen/hovimestari/internal/config"
	"github.com/lepinkainen/hovimestari/internal/store"
	"github.com/lepinkainen/hovimestari/internal/tui/models"
)

// App represents the main TUI application
type App struct {
	store  *store.Store
	config *config.Config
}

// NewApp creates a new TUI application
func NewApp(store *store.Store, config *config.Config) *App {
	return &App{
		store:  store,
		config: config,
	}
}

// Run starts the TUI application
func (a *App) Run() error {
	// Create the main navigation model
	model := models.NewNavigation(a.store, a.config)
	
	// Create and start the Bubbletea program
	p := tea.NewProgram(model, tea.WithAltScreen())
	
	// Run the program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}
	
	return nil
}