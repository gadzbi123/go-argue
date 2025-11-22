package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Color scheme
	model1Color = lipgloss.Color("#00BFFF") // Deep Sky Blue
	model2Color = lipgloss.Color("#32CD32") // Lime Green
	headerColor = lipgloss.Color("#FFD700") // Gold
	errorColor  = lipgloss.Color("#FF6347") // Tomato Red
	subtleColor = lipgloss.Color("#808080") // Gray

	// Styles for model1
	model1Style = lipgloss.NewStyle().
			Foreground(model1Color).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(model1Color).
			Padding(0, 1).
			MarginBottom(1)

	model1LabelStyle = lipgloss.NewStyle().
				Foreground(model1Color).
				Bold(true)

	// Styles for model2
	model2Style = lipgloss.NewStyle().
			Foreground(model2Color).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(model2Color).
			Padding(0, 1).
			MarginBottom(1)

	model2LabelStyle = lipgloss.NewStyle().
				Foreground(model2Color).
				Bold(true)

	// General styles
	headerStyle = lipgloss.NewStyle().
			Foreground(headerColor).
			Bold(true).
			Padding(1, 0)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	subtleStyle = lipgloss.NewStyle().
			Foreground(subtleColor).
			Italic(true)

	timestampStyle = lipgloss.NewStyle().
			Foreground(subtleColor).
			Italic(true)
)

// renderInputView renders the topic input view
func (m *debateModel) renderInputView() string {
	var b strings.Builder

	// Welcome message
	b.WriteString(headerStyle.Render("🎭 AI Debate CLI"))
	b.WriteString("\n\n")

	// Show model names
	b.WriteString(fmt.Sprintf("Models: %s vs %s\n\n",
		model1LabelStyle.Render(m.model1Name),
		model2LabelStyle.Render(m.model2Name)))

	// Render text input for topic
	b.WriteString("Enter a debate topic:\n")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	// Show error if any
	if m.errorMsg != "" {
		b.WriteString(errorStyle.Render(m.errorMsg))
		b.WriteString("\n\n")
	}

	// Instructions
	b.WriteString(subtleStyle.Render("Press Enter to start • Ctrl+C to quit"))

	return b.String()
}

// renderDebateView renders the active debate view
func (m *debateModel) renderDebateView() string {
	var b strings.Builder

	// Render debate topic header
	b.WriteString(headerStyle.Render(fmt.Sprintf("📢 Debate Topic: %s", m.topic)))
	b.WriteString("\n\n")

	// Use viewport width for content formatting
	viewportWidth := m.viewport.Width
	if viewportWidth == 0 {
		viewportWidth = m.width
	}

	// Display all turns with formatting
	for i, turn := range m.history {
		isModel1 := turn.ModelName == m.model1Name
		b.WriteString(formatTurn(turn, isModel1, viewportWidth))
		b.WriteString("\n")

		// Add spacing between turns
		if i < len(m.history)-1 {
			b.WriteString("\n")
		}
	}

	// Show generation indicator for active model
	if m.isGenerating {
		b.WriteString("\n")
		activeModel := m.getNextModel()
		isModel1 := activeModel == m.model1Name

		var indicatorStyle lipgloss.Style
		if isModel1 {
			indicatorStyle = model1LabelStyle
		} else {
			indicatorStyle = model2LabelStyle
		}

		b.WriteString(indicatorStyle.Render(fmt.Sprintf("💭 %s is thinking...", activeModel)))
		b.WriteString("\n")
	}

	// Show error if any
	if m.errorMsg != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(fmt.Sprintf("⚠️  %s", m.errorMsg)))
		b.WriteString("\n")
	}

	// Render viewport with scroll
	m.viewport.SetContent(b.String())

	// Footer with instructions
	footer := subtleStyle.Render("Press 'q' or Ctrl+C to stop the debate")

	return fmt.Sprintf("%s\n%s", m.viewport.View(), footer)
}

// renderStoppedView renders the stopped debate view
func (m *debateModel) renderStoppedView() string {
	var b strings.Builder

	// Show stop confirmation message
	b.WriteString(headerStyle.Render("🛑 Debate Stopped"))
	b.WriteString("\n\n")

	// Display final debate history
	b.WriteString(subtleStyle.Render(fmt.Sprintf("Topic: %s", m.topic)))
	b.WriteString("\n\n")

	for i, turn := range m.history {
		isModel1 := turn.ModelName == m.model1Name
		b.WriteString(formatTurn(turn, isModel1, m.width))
		b.WriteString("\n")

		// Add spacing between turns
		if i < len(m.history)-1 {
			b.WriteString("\n")
		}
	}

	// Provide exit instructions
	b.WriteString("\n\n")
	b.WriteString(subtleStyle.Render("Press any key to exit"))

	return b.String()
}

// renderErrorView renders the error view
func (m *debateModel) renderErrorView() string {
	var b strings.Builder

	// Display error message prominently
	b.WriteString(errorStyle.Render("❌ Error Occurred"))
	b.WriteString("\n\n")
	b.WriteString(errorStyle.Render(m.errorMsg))
	b.WriteString("\n\n")

	// Show existing debate history
	if len(m.history) > 0 {
		b.WriteString(subtleStyle.Render(fmt.Sprintf("Topic: %s", m.topic)))
		b.WriteString("\n\n")

		for i, turn := range m.history {
			isModel1 := turn.ModelName == m.model1Name
			b.WriteString(formatTurn(turn, isModel1, m.width))
			b.WriteString("\n")

			// Add spacing between turns
			if i < len(m.history)-1 {
				b.WriteString("\n")
			}
		}
		b.WriteString("\n\n")
	}

	// Provide recovery or exit options
	b.WriteString(subtleStyle.Render("Press 'q' to exit"))

	return b.String()
}

// formatTurn formats a single turn for display
func formatTurn(turn Turn, isModel1 bool, width int) string {
	var b strings.Builder

	// Format timestamp
	timestamp := turn.Timestamp.Format("15:04:05")

	// Choose style based on model
	var labelStyle lipgloss.Style
	var contentStyle lipgloss.Style

	if isModel1 {
		labelStyle = model1LabelStyle
		contentStyle = model1Style
	} else {
		labelStyle = model2LabelStyle
		contentStyle = model2Style
	}

	// Add model name label with timestamp
	b.WriteString(labelStyle.Render(turn.ModelName))
	b.WriteString(" ")
	b.WriteString(timestampStyle.Render(fmt.Sprintf("[%s]", timestamp)))
	b.WriteString("\n")

	// Calculate available width for content (accounting for border and padding)
	// Border takes 2 chars (left + right), padding takes 2 chars (1 on each side)
	// Also leave some margin for the viewport scrollbar
	contentWidth := width - 6
	if contentWidth < 20 {
		contentWidth = 20 // Minimum width
	}

	// Format content with proper wrapping and width constraint
	b.WriteString(contentStyle.Width(contentWidth).Render(turn.Content))

	return b.String()
}
