package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lepinkainen/hovimestari/internal/store"
)

// FormField represents different form fields
type FormField int

const (
	ContentField FormField = iota
	SourceField
	RelevanceDateField
)

// MemoryForm represents the memory form model
type MemoryForm struct {
	store              *store.Store
	width              int
	height             int
	focused            bool
	currentField       FormField
	
	// Form fields
	contentTextarea    textarea.Model
	sourceInput        textinput.Model
	relevanceDateInput textinput.Model
	
	// State
	editMode           bool
	editingMemory      *store.Memory
	saving             bool
	errorMessage       string
	successMessage     string
}

// MemoryFormSavedMsg represents a message when memory is saved
type MemoryFormSavedMsg struct {
	Memory *store.Memory
	Err    error
}

// NewMemoryForm creates a new memory form model
func NewMemoryForm(store *store.Store) *MemoryForm {
	// Create content textarea
	contentTextarea := textarea.New()
	contentTextarea.Placeholder = "Enter memory content here..."
	contentTextarea.Focus()
	contentTextarea.CharLimit = 5000
	contentTextarea.SetWidth(60)
	contentTextarea.SetHeight(8)
	
	// Create source input
	sourceInput := textinput.New()
	sourceInput.Placeholder = "manual"
	sourceInput.CharLimit = 100
	sourceInput.Width = 60
	
	// Create relevance date input
	relevanceDateInput := textinput.New()
	relevanceDateInput.Placeholder = "2025-01-01 (YYYY-MM-DD, optional)"
	relevanceDateInput.CharLimit = 10
	relevanceDateInput.Width = 60
	
	return &MemoryForm{
		store:              store,
		contentTextarea:    contentTextarea,
		sourceInput:        sourceInput,
		relevanceDateInput: relevanceDateInput,
		currentField:       ContentField,
		focused:            true,
	}
}

// NewMemoryFormForEdit creates a memory form for editing an existing memory
func NewMemoryFormForEdit(store *store.Store, memory *store.Memory) *MemoryForm {
	form := NewMemoryForm(store)
	form.editMode = true
	form.editingMemory = memory
	
	// Pre-fill form with existing data
	form.contentTextarea.SetValue(memory.Content)
	form.sourceInput.SetValue(memory.Source)
	
	if memory.RelevanceDate != nil {
		form.relevanceDateInput.SetValue(memory.RelevanceDate.Format("2006-01-02"))
	}
	
	return form
}

// Init initializes the memory form
func (m *MemoryForm) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		textarea.Blink,
	)
}

// Update handles messages for the memory form
func (m *MemoryForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Adjust form field sizes
		m.contentTextarea.SetWidth(msg.Width - 10)
		m.sourceInput.Width = msg.Width - 10
		m.relevanceDateInput.Width = msg.Width - 10
		
	case MemoryFormSavedMsg:
		m.saving = false
		if msg.Err != nil {
			m.errorMessage = fmt.Sprintf("Error saving memory: %v", msg.Err)
			m.successMessage = ""
		} else {
			m.successMessage = "Memory saved successfully!"
			m.errorMessage = ""
			// Clear form if not in edit mode
			if !m.editMode {
				m.contentTextarea.SetValue("")
				m.sourceInput.SetValue("")
				m.relevanceDateInput.SetValue("")
				m.currentField = ContentField
				m.updateFieldFocus()
			}
		}
		
	case tea.KeyMsg:
		// Clear messages on new input
		m.errorMessage = ""
		m.successMessage = ""
		
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+s":
			// Save memory
			return m, m.saveMemory()
		case "tab":
			// Move to next field
			m.nextField()
			return m, nil
		case "shift+tab":
			// Move to previous field
			m.prevField()
			return m, nil
		}
	}
	
	// Update the current focused field
	switch m.currentField {
	case ContentField:
		m.contentTextarea, cmd = m.contentTextarea.Update(msg)
		cmds = append(cmds, cmd)
	case SourceField:
		m.sourceInput, cmd = m.sourceInput.Update(msg)
		cmds = append(cmds, cmd)
	case RelevanceDateField:
		m.relevanceDateInput, cmd = m.relevanceDateInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	
	return m, tea.Batch(cmds...)
}

// View renders the memory form
func (m *MemoryForm) View() string {
	var sections []string
	
	// Title
	title := "Add New Memory"
	if m.editMode {
		title = fmt.Sprintf("Edit Memory #%d", m.editingMemory.ID)
	}
	
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render(title)
	
	sections = append(sections, header)
	
	// Instructions
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginBottom(1).
		Render("Use Tab/Shift+Tab to navigate fields, Ctrl+S to save, ESC to cancel")
	
	sections = append(sections, instructions)
	
	// Content field
	contentLabel := m.renderFieldLabel("Content:", m.currentField == ContentField)
	sections = append(sections, contentLabel)
	sections = append(sections, m.contentTextarea.View())
	
	// Source field
	sourceLabel := m.renderFieldLabel("Source:", m.currentField == SourceField)
	sections = append(sections, sourceLabel)
	sections = append(sections, m.sourceInput.View())
	
	// Relevance date field
	dateLabel := m.renderFieldLabel("Relevance Date:", m.currentField == RelevanceDateField)
	sections = append(sections, dateLabel)
	sections = append(sections, m.relevanceDateInput.View())
	
	// Status messages
	if m.saving {
		status := lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			MarginTop(1).
			Render("Saving...")
		sections = append(sections, status)
	} else if m.successMessage != "" {
		status := lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			MarginTop(1).
			Render(m.successMessage)
		sections = append(sections, status)
	} else if m.errorMessage != "" {
		status := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			MarginTop(1).
			Render(m.errorMessage)
		sections = append(sections, status)
	}
	
	// Help text
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(1).
		Render("Ctrl+S: Save  •  Tab: Next field  •  ESC: Cancel")
	
	sections = append(sections, helpText)
	
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderFieldLabel renders a field label with focus indication
func (m *MemoryForm) renderFieldLabel(label string, focused bool) string {
	style := lipgloss.NewStyle().
		Bold(true).
		MarginTop(1).
		MarginBottom(0)
	
	if focused {
		style = style.Foreground(lipgloss.Color("205"))
		label = "▶ " + label
	} else {
		style = style.Foreground(lipgloss.Color("240"))
		label = "  " + label
	}
	
	return style.Render(label)
}

// nextField moves to the next form field
func (m *MemoryForm) nextField() {
	m.currentField = (m.currentField + 1) % 3
	m.updateFieldFocus()
}

// prevField moves to the previous form field
func (m *MemoryForm) prevField() {
	m.currentField = (m.currentField + 2) % 3 // +2 is equivalent to -1 in mod 3
	m.updateFieldFocus()
}

// updateFieldFocus updates the focus state of form fields
func (m *MemoryForm) updateFieldFocus() {
	// Blur all fields first
	m.contentTextarea.Blur()
	m.sourceInput.Blur()
	m.relevanceDateInput.Blur()
	
	// Focus the current field
	switch m.currentField {
	case ContentField:
		m.contentTextarea.Focus()
	case SourceField:
		m.sourceInput.Focus()
	case RelevanceDateField:
		m.relevanceDateInput.Focus()
	}
}

// saveMemory saves the memory to the store
func (m *MemoryForm) saveMemory() tea.Cmd {
	return func() tea.Msg {
		// Validate form
		content := strings.TrimSpace(m.contentTextarea.Value())
		if content == "" {
			return MemoryFormSavedMsg{Err: fmt.Errorf("content cannot be empty")}
		}
		
		source := strings.TrimSpace(m.sourceInput.Value())
		if source == "" {
			source = "manual"
		}
		
		// Parse relevance date
		var relevanceDate *time.Time
		dateStr := strings.TrimSpace(m.relevanceDateInput.Value())
		if dateStr != "" {
			parsed, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				return MemoryFormSavedMsg{Err: fmt.Errorf("invalid date format. Use YYYY-MM-DD")}
			}
			relevanceDate = &parsed
		}
		
		var memory *store.Memory
		var err error
		
		if m.editMode && m.editingMemory != nil {
			// Update existing memory
			m.editingMemory.Content = content
			m.editingMemory.Source = source
			m.editingMemory.RelevanceDate = relevanceDate
			
			err = m.store.UpdateMemory(m.editingMemory)
			memory = m.editingMemory
		} else {
			// Create new memory
			var id int64
			id, err = m.store.AddMemory(content, relevanceDate, source, nil)
			if err == nil {
				memory = &store.Memory{
					ID:            id,
					Content:       content,
					Source:        source,
					RelevanceDate: relevanceDate,
					CreatedAt:     time.Now(),
				}
			}
		}
		
		return MemoryFormSavedMsg{Memory: memory, Err: err}
	}
}

// Reset resets the form to initial state
func (m *MemoryForm) Reset() {
	m.contentTextarea.SetValue("")
	m.sourceInput.SetValue("")
	m.relevanceDateInput.SetValue("")
	m.currentField = ContentField
	m.editMode = false
	m.editingMemory = nil
	m.errorMessage = ""
	m.successMessage = ""
	m.updateFieldFocus()
}