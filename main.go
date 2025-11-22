package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Parse command-line flags
	model1 := flag.String("model1", "gemma3n:e4b", "First AI model for the debate")
	model2 := flag.String("model2", "gemma3:4b", "Second AI model for the debate")
	flag.Parse()

	// Create Ollama client
	client := NewOllamaClient("")

	// Validate both models are available
	fmt.Printf("Validating models...\n")
	if err := client.ValidateModel(*model1); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Model '%s' is not available.\n", *model1)
		fmt.Fprintf(os.Stderr, "Please ensure Ollama is running and the model is installed.\n")
		fmt.Fprintf(os.Stderr, "You can install it with: ollama pull %s\n", *model1)
		os.Exit(1)
	}

	if err := client.ValidateModel(*model2); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Model '%s' is not available.\n", *model2)
		fmt.Fprintf(os.Stderr, "Please ensure Ollama is running and the model is installed.\n")
		fmt.Fprintf(os.Stderr, "You can install it with: ollama pull %s\n", *model2)
		os.Exit(1)
	}

	fmt.Printf("✓ Models validated: %s and %s\n\n", *model1, *model2)

	// Create initial model with validated models
	initialModel := debateModel{
		model1Name:   *model1,
		model2Name:   *model2,
		ollamaClient: client,
		currentTurn:  0,
		history:      []Turn{},
		state:        stateInput,
	}

	// Configure and run Bubbletea program
	p := tea.NewProgram(&initialModel, tea.WithAltScreen())

	// Run program and handle exit
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
