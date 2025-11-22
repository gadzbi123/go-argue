package main

import (
	"fmt"
	"strings"
)

// BuildDebatePrompt constructs a debate prompt with full context for a model.
// It includes the debate topic, conversation history, and instructions for the model
// to engage in debate. For the first turn, it assigns initial positions.
func BuildDebatePrompt(topic string, history []Turn, currentModel string, isFirstTurn bool) string {
	var prompt strings.Builder

	// Add debate context
	prompt.WriteString(fmt.Sprintf("You are participating in a debate on the topic: \"%s\"\n\n", topic))
	prompt.WriteString(fmt.Sprintf("You are %s. Your role is to present arguments and respond to your opponent's points.\n\n", currentModel))

	// For the first turn, assign positions
	if isFirstTurn {
		// Determine if this is model1 or model2 based on position in debate
		// Model1 (first to speak) takes the "pro" position
		// Model2 takes the "con" position
		if len(history) == 0 {
			prompt.WriteString("You will be presenting the opening argument. Take a clear position on this topic and present your initial arguments.\n\n")
		} else {
			prompt.WriteString("You will be responding to the opening argument. Take an opposing or alternative perspective and present your counterarguments.\n\n")
		}
	}

	// Add conversation history if it exists
	if len(history) > 0 {
		prompt.WriteString("Previous discussion:\n")
		prompt.WriteString(FormatHistory(history))
		prompt.WriteString("\n")
	}

	// Add instructions for the response
	if len(history) > 0 {
		prompt.WriteString("Provide your next argument or response. Be thoughtful, specific, and engage directly with the previous points made.\n")
	} else {
		prompt.WriteString("Provide your opening argument. Be thoughtful, specific, and clearly state your position.\n")
	}

	return prompt.String()
}

// FormatHistory structures the conversation history for model consumption.
// Each turn is formatted with the model name and content, making it clear
// which model made each statement.
func FormatHistory(history []Turn) string {
	var formatted strings.Builder

	for i, turn := range history {
		formatted.WriteString(fmt.Sprintf("[%s]: %s", turn.ModelName, turn.Content))

		// Add newline between turns, but not after the last one
		if i < len(history)-1 {
			formatted.WriteString("\n\n")
		}
	}

	return formatted.String()
}
