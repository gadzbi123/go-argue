package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: ai-debate-cli, Property 2: Valid topics initialize both models
// Validates: Requirements 1.3
//
// For any non-empty topic string, initializing the debate should successfully
// prepare context for both AI models.
func TestProperty_ValidTopicsInitializeBothModels(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("valid topics initialize debate with both models ready", prop.ForAll(
		func(topic string) bool {
			// Create a debate model with two models configured
			model := debateModel{
				model1Name:   "mistral:7b",
				model2Name:   "gemma3:4b",
				ollamaClient: NewOllamaClient("http://localhost:11434"),
				state:        stateInput,
				history:      []Turn{},
				currentTurn:  0,
				isGenerating: false,
			}

			// Initialize text input
			ti := textinput.New()
			ti.SetValue(topic)
			model.textInput = ti

			// Initialize viewport
			vp := viewport.New(80, 24)
			model.viewport = vp

			// Simulate the user pressing Enter to submit the topic
			msg := tea.KeyMsg{Type: tea.KeyEnter}
			updatedModel, _ := model.Update(msg)
			m := updatedModel.(debateModel)

			// Property 1: The topic should be set in the model
			if m.topic != topic {
				return false
			}

			// Property 2: The state should transition to debating
			if m.state != stateDebating {
				return false
			}

			// Property 3: Both models should be accessible (model names should be set)
			if m.model1Name == "" || m.model2Name == "" {
				return false
			}

			// Property 4: The debate should start with model1 (currentTurn should be 0)
			if m.currentTurn != 0 {
				return false
			}

			// Property 5: Generation should be marked as active
			if !m.isGenerating {
				return false
			}

			// Property 6: No error message should be present
			if m.errorMsg != "" {
				return false
			}

			return true
		},
		genValidTopic(),
	))

	properties.TestingRun(t)
}

// genValidTopic generates non-empty topic strings (valid topics)
func genValidTopic() gopter.Gen {
	return gen.AnyString().SuchThat(func(s string) bool {
		// Valid topics are non-empty after trimming whitespace
		return len(strings.TrimSpace(s)) > 0
	}).Map(func(s string) string {
		// Ensure we have a reasonable topic
		trimmed := strings.TrimSpace(s)
		if len(trimmed) == 0 {
			return "Default debate topic"
		}
		return trimmed
	})
}
