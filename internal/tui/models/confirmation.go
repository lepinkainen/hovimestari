package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmationDialog represents a confirmation dialog
type ConfirmationDialog struct {
	title    string
	message  string
	width    int
	height   int
	focused  bool
	selected bool // true for Yes, false for No
}

// ConfirmationResult represents the result of a confirmation dialog
type ConfirmationResult struct {
	Confirmed bool
	Cancelled bool
}

// NewConfirmationDialog creates a new confirmation dialog
func NewConfirmationDialog(title, message string) *ConfirmationDialog {
	return &ConfirmationDialog{
		title:    title,
		message:  message,
		focused:  true,
		selected: false, // Default to "No"
	}
}

// Init initializes the confirmation dialog
func (c *ConfirmationDialog) Init() tea.Cmd {
	return nil
}

// Update handles messages for the confirmation dialog
func (c *ConfirmationDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			c.selected = false // Select "No"
		case "right", "l":
			c.selected = true // Select "Yes"
		case "tab":
			c.selected = !c.selected // Toggle selection
		case "enter":
			return c, func() tea.Msg {
				return ConfirmationResult{Confirmed: c.selected, Cancelled: false}
			}
		case "esc", "q":
			return c, func() tea.Msg {
				return ConfirmationResult{Confirmed: false, Cancelled: true}
			}
		}
	}
	
	return c, nil
}

// View renders the confirmation dialog
func (c *ConfirmationDialog) View() string {
	// Calculate dialog dimensions
	dialogWidth := 50
	if c.width > 0 && c.width-10 < dialogWidth {
		dialogWidth = c.width - 10
	}
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Align(lipgloss.Center).
		Width(dialogWidth-4)
	
	title := titleStyle.Render(c.title)
	
	// Message
	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Align(lipgloss.Center).
		Width(dialogWidth-4).
		MarginTop(1).
		MarginBottom(2)
	
	message := messageStyle.Render(c.message)
	
	// Buttons
	noStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("255"))
	
	yesStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("255"))
	
	if !c.selected {
		// "No" is selected
		noStyle = noStyle.
			Background(lipgloss.Color("196")).
			BorderForeground(lipgloss.Color("196")).
			Foreground(lipgloss.Color("255"))
	} else {
		// "Yes" is selected
		yesStyle = yesStyle.
			Background(lipgloss.Color("46")).
			BorderForeground(lipgloss.Color("46")).
			Foreground(lipgloss.Color("0"))
	}
	
	noButton := noStyle.Render("No")
	yesButton := yesStyle.Render("Yes")
	
	buttons := lipgloss.JoinHorizontal(lipgloss.Center, noButton, "  ", yesButton)
	buttonsContainer := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(dialogWidth-4).
		Render(buttons)
	
	// Instructions
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Align(lipgloss.Center).
		Width(dialogWidth-4).
		MarginTop(1).
		Render("← → / Tab: Select  •  Enter: Confirm  •  ESC: Cancel")
	
	// Content
	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		message,
		buttonsContainer,
		instructions,
	)
	
	// Dialog box
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(dialogWidth).
		Align(lipgloss.Center)
	
	dialog := dialogStyle.Render(content)
	
	// Center the dialog on screen
	if c.height > 0 {
		verticalPadding := (c.height - lipgloss.Height(dialog)) / 2
		if verticalPadding > 0 {
			dialog = lipgloss.NewStyle().
				Padding(verticalPadding, 0).
				Render(dialog)
		}
	}
	
	return lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(c.width).
		Render(dialog)
}