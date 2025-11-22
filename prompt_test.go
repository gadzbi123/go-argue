package main

import (
	"strings"
	"testing"
	"time"
)

func TestBuildDebatePrompt_FirstTurn(t *testing.T) {
	topic := "Is artificial intelligence beneficial for humanity?"
	history := []Turn{}
	currentModel := "mistral:7b"
	isFirstTurn := true

	prompt := BuildDebatePrompt(topic, history, currentModel, isFirstTurn)

	// Verify topic is included
	if !strings.Contains(prompt, topic) {
		t.Errorf("Prompt should contain the topic")
	}

	// Verify model name is included
	if !strings.Contains(prompt, currentModel) {
		t.Errorf("Prompt should contain the model name")
	}

	// Verify debate instructions are included
	if !strings.Contains(prompt, "debate") {
		t.Errorf("Prompt should contain debate instructions")
	}

	// Verify opening argument instruction for first turn
	if !strings.Contains(prompt, "opening argument") {
		t.Errorf("First turn should mention opening argument")
	}
}

func TestBuildDebatePrompt_WithHistory(t *testing.T) {
	topic := "Should we colonize Mars?"
	history := []Turn{
		{
			ModelName: "mistral:7b",
			Content:   "Mars colonization is essential for humanity's survival.",
			Timestamp: time.Now(),
		},
		{
			ModelName: "gemma3:4b",
			Content:   "The costs outweigh the benefits at this time.",
			Timestamp: time.Now(),
		},
	}
	currentModel := "mistral:7b"
	isFirstTurn := false

	prompt := BuildDebatePrompt(topic, history, currentModel, isFirstTurn)

	// Verify topic is included
	if !strings.Contains(prompt, topic) {
		t.Errorf("Prompt should contain the topic")
	}

	// Verify history is included
	if !strings.Contains(prompt, "Previous discussion:") {
		t.Errorf("Prompt should indicate previous discussion")
	}

	// Verify both previous turns are in the prompt
	if !strings.Contains(prompt, history[0].Content) {
		t.Errorf("Prompt should contain first turn content")
	}
	if !strings.Contains(prompt, history[1].Content) {
		t.Errorf("Prompt should contain second turn content")
	}

	// Verify model attribution in history
	if !strings.Contains(prompt, history[0].ModelName) {
		t.Errorf("Prompt should show model attribution for first turn")
	}
	if !strings.Contains(prompt, history[1].ModelName) {
		t.Errorf("Prompt should show model attribution for second turn")
	}
}

func TestFormatHistory_Empty(t *testing.T) {
	history := []Turn{}
	formatted := FormatHistory(history)

	if formatted != "" {
		t.Errorf("Empty history should produce empty string, got: %s", formatted)
	}
}

func TestFormatHistory_SingleTurn(t *testing.T) {
	history := []Turn{
		{
			ModelName: "mistral:7b",
			Content:   "This is my argument.",
			Timestamp: time.Now(),
		},
	}

	formatted := FormatHistory(history)

	// Verify model name is included
	if !strings.Contains(formatted, "mistral:7b") {
		t.Errorf("Formatted history should contain model name")
	}

	// Verify content is included
	if !strings.Contains(formatted, "This is my argument.") {
		t.Errorf("Formatted history should contain turn content")
	}

	// Verify format includes model attribution
	if !strings.Contains(formatted, "[mistral:7b]:") {
		t.Errorf("Formatted history should use [ModelName]: format")
	}
}

func TestFormatHistory_MultipleTurns(t *testing.T) {
	history := []Turn{
		{
			ModelName: "mistral:7b",
			Content:   "First argument.",
			Timestamp: time.Now(),
		},
		{
			ModelName: "gemma3:4b",
			Content:   "Counter argument.",
			Timestamp: time.Now(),
		},
		{
			ModelName: "mistral:7b",
			Content:   "Rebuttal.",
			Timestamp: time.Now(),
		},
	}

	formatted := FormatHistory(history)

	// Verify all turns are included
	for _, turn := range history {
		if !strings.Contains(formatted, turn.ModelName) {
			t.Errorf("Formatted history should contain model name: %s", turn.ModelName)
		}
		if !strings.Contains(formatted, turn.Content) {
			t.Errorf("Formatted history should contain content: %s", turn.Content)
		}
	}

	// Verify turns are separated
	if !strings.Contains(formatted, "\n\n") {
		t.Errorf("Multiple turns should be separated by double newlines")
	}
}
