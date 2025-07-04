package models

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lepinkainen/hovimestari/internal/store"
)

// MemoryDetail represents the memory detail view model
type MemoryDetail struct {
	memory  *store.Memory
	width   int
	height  int
	focused bool
}

// NewMemoryDetail creates a new memory detail model
func NewMemoryDetail(memory *store.Memory) *MemoryDetail {
	return &MemoryDetail{
		memory:  memory,
		focused: true,
	}
}

// Init initializes the memory detail view
func (m *MemoryDetail) Init() tea.Cmd {
	return nil
}

// Update handles messages for the memory detail view
func (m *MemoryDetail) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	
	return m, nil
}

// View renders the memory detail view
func (m *MemoryDetail) View() string {
	if m.memory == nil {
		return "No memory selected"
	}
	
	var sections []string
	
	// Header with memory ID and source
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render(fmt.Sprintf("Memory #%d [%s]", m.memory.ID, m.memory.Source))
	
	sections = append(sections, header)
	
	// Content section
	contentTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		MarginTop(1).
		Render("Content:")
	
	content := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(m.width - 4).
		Render(m.formatContent(m.memory.Content))
	
	sections = append(sections, contentTitle, content)
	
	// Metadata section
	metadataTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		MarginTop(1).
		Render("Metadata:")
	
	metadata := m.renderMetadata()
	metadataBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(m.width - 4).
		Render(metadata)
	
	sections = append(sections, metadataTitle, metadataBox)
	
	// Instructions
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(1).
		Render("Press ESC or q to go back")
	
	sections = append(sections, instructions)
	
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// formatContent formats the memory content for display
func (m *MemoryDetail) formatContent(content string) string {
	// Word wrap the content to fit within the available width
	maxWidth := m.width - 8 // Account for border and padding
	if maxWidth < 20 {
		maxWidth = 20
	}
	
	return wordWrap(content, maxWidth)
}

// renderMetadata renders the memory metadata
func (m *MemoryDetail) renderMetadata() string {
	var lines []string
	
	// Created date
	lines = append(lines, fmt.Sprintf("Created: %s", 
		m.memory.CreatedAt.Format("2006-01-02 15:04:05")))
	
	// Relevance date
	if m.memory.RelevanceDate != nil {
		lines = append(lines, fmt.Sprintf("Relevant: %s", 
			m.memory.RelevanceDate.Format("2006-01-02")))
	} else {
		lines = append(lines, "Relevant: Always")
	}
	
	// Source
	lines = append(lines, fmt.Sprintf("Source: %s", m.memory.Source))
	
	// UID if available
	if m.memory.UID != nil {
		lines = append(lines, fmt.Sprintf("UID: %s", *m.memory.UID))
	}
	
	// Content length
	lines = append(lines, fmt.Sprintf("Length: %d characters", len(m.memory.Content)))
	
	return strings.Join(lines, "\n")
}

// wordWrap wraps text to the specified width
func wordWrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}
	
	var lines []string
	var currentLine strings.Builder
	
	for _, word := range words {
		// If adding this word would exceed the width, start a new line
		if currentLine.Len() > 0 && currentLine.Len()+len(word)+1 > width {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}
		
		// Add word to current line
		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}
	
	// Add the last line
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}
	
	return strings.Join(lines, "\n")
}