package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// appState represents the current state of the application
type appState int

const (
	stateInput appState = iota
	stateDebating
	stateStopped
	stateError
)

// Turn represents a single contribution to the debate from one model
type Turn struct {
	ModelName string
	Content   string
	Timestamp time.Time
}

// DebateContext represents the complete conversation context passed to models
type DebateContext struct {
	Topic   string
	History []Turn
}

// debateModel holds the application state
type debateModel struct {
	// Configuration
	model1Name   string
	model2Name   string
	ollamaClient *OllamaClient

	// Debate state
	topic        string
	history      []Turn
	currentTurn  int // 0 for model1, 1 for model2
	isGenerating bool

	// UI state
	state     appState
	viewport  viewport.Model
	textInput textinput.Model
	errorMsg  string

	// Dimensions
	width  int
	height int
}

// Init initializes the model
func (m debateModel) Init() tea.Cmd {
	// Initialize text input for topic entry
	ti := textinput.New()
	ti.Placeholder = "Enter a debate topic..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 50

	m.textInput = ti
	m.state = stateInput

	// Return command to focus the text input
	return textinput.Blink
}

// Update handles messages and updates the model
func (m debateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// Handle keyboard input
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Handle stop command
			if m.state == stateDebating {
				m.state = stateStopped
				return m, tea.Quit
			}
			return m, tea.Quit

		case "enter":
			// Handle topic submission
			if m.state == stateInput {
				topic := m.textInput.Value()
				// Validate topic is non-empty
				if len(strings.TrimSpace(topic)) == 0 {
					m.errorMsg = "Topic cannot be empty"
					return m, nil
				}

				// Transition to debating state
				m.topic = topic
				m.state = stateDebating
				m.errorMsg = ""
				m.isGenerating = true
				m.currentTurn = 0 // Start with model1

				// Start first model generation
				return m, m.generateResponse()
			}
		}

		// Pass other keys to text input component when in input state
		if m.state == stateInput {
			m.textInput, cmd = m.textInput.Update(msg)
			cmds = append(cmds, cmd)
		}

	// Handle terminal resize
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Resize viewport component
		if m.state == stateDebating || m.state == stateStopped {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 5 // Leave room for header and footer
		}

	// Handle response chunks
	case responseChunkMsg:
		if m.isGenerating {
			// Append chunk to current turn content
			if len(m.history) > 0 && m.history[len(m.history)-1].ModelName == m.getNextModel() {
				// Update the last turn if it's from the current model
				m.history[len(m.history)-1].Content += msg.chunk
			} else {
				// Create a new turn for this model
				m.history = append(m.history, Turn{
					ModelName: m.getNextModel(),
					Content:   msg.chunk,
					Timestamp: time.Now(),
				})
			}

			// Update viewport with new content
			m.viewport.SetContent(m.formatDebateHistory())
			m.viewport.GotoBottom()
		}

	// Handle response completion
	case responseCompleteMsg:
		m.isGenerating = false

		// Ensure the turn is properly recorded in history
		if len(m.history) == 0 || m.history[len(m.history)-1].Content != msg.fullResponse {
			// If the last turn doesn't match, update or create it
			if len(m.history) > 0 && m.history[len(m.history)-1].ModelName == m.getNextModel() {
				m.history[len(m.history)-1].Content = msg.fullResponse
				m.history[len(m.history)-1].Timestamp = time.Now()
			} else {
				m.history = append(m.history, Turn{
					ModelName: m.getNextModel(),
					Content:   msg.fullResponse,
					Timestamp: time.Now(),
				})
			}
		}

		// Switch to the opposite model
		m = m.switchTurn()

		// Trigger next turn
		m.isGenerating = true
		return m, m.generateResponse()

	// Handle errors
	case responseErrorMsg:
		m.isGenerating = false

		// Display error message in UI
		m.errorMsg = fmt.Sprintf("Error: %v", msg.err)

		// Preserve existing history (already done by not modifying it)

		// Attempt to continue with next turn if recoverable
		m = m.switchTurn()
		m.isGenerating = true
		return m, m.generateResponse()

	// Handle stop command
	case stopDebateMsg:
		m.isGenerating = false
		m.state = stateStopped
		return m, tea.Quit
	}

	// Update viewport if in debating state
	if m.state == stateDebating || m.state == stateStopped {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m debateModel) View() string {
	switch m.state {
	case stateInput:
		return m.renderInputView()
	case stateDebating:
		return m.renderDebateView()
	case stateStopped:
		return m.renderStoppedView()
	case stateError:
		return m.renderErrorView()
	default:
		return "Unknown state"
	}
}

// getNextModel returns the name of the model that should speak next.
// It alternates between model1 and model2 based on the current turn counter.
// currentTurn 0 means model1, currentTurn 1 means model2.
func (m debateModel) getNextModel() string {
	if m.currentTurn == 0 {
		return m.model1Name
	}
	return m.model2Name
}

// switchTurn toggles the current turn between model1 (0) and model2 (1).
// It returns the updated model.
func (m debateModel) switchTurn() debateModel {
	if m.currentTurn == 0 {
		m.currentTurn = 1
	} else {
		m.currentTurn = 0
	}
	return m
}

// generateResponse starts generating a response from the current model.
// It returns a Cmd that will send responseChunkMsg and responseCompleteMsg.
func (m debateModel) generateResponse() tea.Cmd {
	return func() tea.Msg {
		modelName := m.getNextModel()
		isFirstTurn := len(m.history) == 0

		// Build the prompt with full context
		prompt := BuildDebatePrompt(m.topic, m.history, modelName, isFirstTurn)

		// Generate response using Ollama client
		ctx := context.Background()
		responseChan, errorChan := m.ollamaClient.GenerateResponse(ctx, modelName, prompt)

		var fullResponse strings.Builder

		// Read from channels
		for {
			select {
			case chunk, ok := <-responseChan:
				if !ok {
					// Channel closed, response complete
					return responseCompleteMsg{fullResponse: fullResponse.String()}
				}
				fullResponse.WriteString(chunk)
				// Send chunk to UI
				// Note: We're only sending the complete message for simplicity
				// In a real implementation, we'd send chunks via a different mechanism

			case err, ok := <-errorChan:
				if ok && err != nil {
					return responseErrorMsg{err: err}
				}
			}
		}
	}
}

// formatDebateHistory formats the debate history for display in the viewport
func (m debateModel) formatDebateHistory() string {
	var output strings.Builder

	for _, turn := range m.history {
		output.WriteString(fmt.Sprintf("[%s]: %s\n\n", turn.ModelName, turn.Content))
	}

	if m.isGenerating {
		output.WriteString(fmt.Sprintf("[%s is thinking...]\n", m.getNextModel()))
	}

	return output.String()
}
