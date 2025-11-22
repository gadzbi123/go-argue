package main

import (
	"strings"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: ai-debate-cli, Property 12: Context includes full history
// Validates: Requirements 4.2, 4.3, 9.1
//
// For any turn N in the debate, the prompt sent to the model should include
// the topic and all previous turns (0 through N-1) with correct model attribution.
func TestProperty_ContextIncludesFullHistory(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("prompt includes topic and all previous turns with attribution", prop.ForAll(
		func(topic string, historySize int, currentModel string) bool {
			// Generate a history of the specified size
			history := make([]Turn, historySize)
			modelNames := []string{"mistral:7b", "gemma3:4b"}

			for i := 0; i < historySize; i++ {
				history[i] = Turn{
					ModelName: modelNames[i%2],
					Content:   generateContent(i),
					Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
				}
			}

			// Build the prompt for turn N (after all history)
			isFirstTurn := historySize == 0
			prompt := BuildDebatePrompt(topic, history, currentModel, isFirstTurn)

			// Property 1: Prompt must include the topic
			if !strings.Contains(prompt, topic) {
				return false
			}

			// Property 2: Prompt must include all previous turns (0 through N-1)
			for i := 0; i < historySize; i++ {
				// Each turn's content must be in the prompt
				if !strings.Contains(prompt, history[i].Content) {
					return false
				}

				// Each turn must have correct model attribution
				// The format is [ModelName]: Content
				expectedAttribution := "[" + history[i].ModelName + "]:"
				if !strings.Contains(prompt, expectedAttribution) {
					return false
				}
			}

			// Property 3: The current model name should be in the prompt
			if !strings.Contains(prompt, currentModel) {
				return false
			}

			return true
		},
		genNonEmptyTopic(),
		gen.IntRange(0, 10), // History size from 0 to 10 turns
		genModelName(),
	))

	properties.TestingRun(t)
}

// genNonEmptyTopic generates non-empty topic strings
func genNonEmptyTopic() gopter.Gen {
	return gen.AnyString().SuchThat(func(s string) bool {
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

// genModelName generates valid model names
func genModelName() gopter.Gen {
	modelNames := []string{"mistral:7b", "gemma3:4b", "llama2:13b", "phi:latest"}
	return gen.OneConstOf(
		modelNames[0],
		modelNames[1],
		modelNames[2],
		modelNames[3],
	)
}

// generateContent creates content for a turn based on its index
func generateContent(index int) string {
	contents := []string{
		"This is a strong opening argument.",
		"I disagree with that perspective.",
		"Let me provide additional evidence.",
		"That's an interesting point, but consider this.",
		"I must respectfully counter that claim.",
		"The data supports my position.",
		"Your argument has merit, however.",
		"Let me clarify my stance.",
		"This is a critical consideration.",
		"I concede that point, but.",
	}
	return contents[index%len(contents)]
}
