# Implementation Plan

- [x] 1. Initialize Go project and dependencies
  - Create Go module with `go mod init`
  - Add Bubbletea, Lipgloss, and Bubbles dependencies
  - Add gopter for property-based testing
  - Create basic project structure (main.go, model.go, ollama.go, etc.)
  - _Requirements: 8.1_

- [x] 2. Implement Ollama client
  - [x] 2.1 Create OllamaClient struct with HTTP client
    - Implement NewOllamaClient constructor
    - Configure base URL (default: http://localhost:11434)
    - _Requirements: 2.2, 7.1_
  
  - [x] 2.2 Implement model validation
    - Create ValidateModel function to check model availability
    - Implement ListModels to fetch available models from Ollama
    - Handle connection errors gracefully
    - _Requirements: 2.2, 2.3_
  
  - [ ]* 2.3 Write property test for model validation
    - **Property 4: Model validation identifies availability**
    - **Validates: Requirements 2.2**
  
  - [x] 2.4 Implement streaming response generation
    - Create GenerateResponse function with context support
    - Parse newline-delimited JSON responses
    - Return channels for response chunks and errors
    - Handle streaming interruptions
    - _Requirements: 4.2, 7.2_
  
  - [x] 2.5 Write unit tests for Ollama client
    - Test HTTP request formatting
    - Test response parsing
    - Test error handling for network failures
    - _Requirements: 7.1, 7.2_

- [-] 3. Implement prompt builder
  - [x] 3.1 Create prompt building functions
    - Implement BuildDebatePrompt with topic and history
    - Implement FormatHistory to structure conversation context
    - Add debate instructions and position assignment
    - _Requirements: 9.1, 9.4_
  
  - [x] 3.2 Write property test for context completeness
    - **Property 12: Context includes full history**
    - **Validates: Requirements 4.2, 4.3, 9.1**
  
  - [ ]* 3.3 Write property test for model attribution
    - **Property 19: Context formatting shows model attribution**
    - **Validates: Requirements 9.2**
  
  - [ ]* 3.4 Write property test for debate instructions
    - **Property 20: Prompts include debate instructions**
    - **Validates: Requirements 9.4**
  
  - [ ]* 3.5 Write unit tests for prompt builder
    - Test prompt structure with empty history
    - Test prompt structure with multiple turns
    - Test first turn position assignment
    - _Requirements: 9.1, 9.2, 9.4_

- [x] 4. Implement core data models
  - [x] 4.1 Define Turn and DebateContext structs
    - Create Turn with ModelName, Content, Timestamp
    - Create DebateContext with Topic and History
    - _Requirements: 3.5, 4.3_
  
  - [x] 4.2 Define Bubbletea model structure
    - Create debateModel with all state fields
    - Define appState enum (input, debating, stopped, error)
    - Initialize viewport and textInput components
    - _Requirements: 8.5_
  
  - [x] 4.3 Implement turn alternation logic
    - Create function to determine next model
    - Ensure alternation between model1 and model2
    - _Requirements: 4.1_
  
  - [ ]* 4.4 Write property test for turn alternation
    - **Property 11: Turns alternate between models**
    - **Validates: Requirements 4.1**

- [x] 5. Implement message types
  - [x] 5.1 Define all Bubbletea message types
    - Create topicSubmittedMsg
    - Create responseChunkMsg and responseCompleteMsg
    - Create responseErrorMsg and nextTurnMsg
    - Create stopDebateMsg
    - _Requirements: 8.3_

- [x] 6. Implement Bubbletea Init function
  - [x] 6.1 Create Init method for debateModel
    - Initialize text input for topic entry
    - Set initial state to stateInput
    - Return initial command (focus text input)
    - _Requirements: 1.1, 8.2_

- [x] 7. Implement Bubbletea Update function
  - [x] 7.1 Handle topic submission
    - Validate topic is non-empty
    - Transition to debating state
    - Start first model generation
    - _Requirements: 1.2, 1.3, 2.4_
  
  - [ ]* 7.2 Write property test for topic validation
    - **Property 1: Topic validation rejects empty input**
    - **Validates: Requirements 1.2**
  
  - [ ]* 7.3 Write property test for model initialization
    - **Property 2: Valid topics initialize both models**
    - **Validates: Requirements 1.3**
  
  - [ ]* 7.4 Write property test for state transition
    - **Property 5: Valid models proceed to topic selection**
    - **Validates: Requirements 2.4**
  
  - [x] 7.5 Handle response chunks
    - Append chunks to current turn content
    - Update viewport with new content
    - Maintain UI responsiveness
    - _Requirements: 3.3, 6.1_
  
  - [ ]* 7.6 Write property test for history growth
    - **Property 8: New responses append to history**
    - **Validates: Requirements 3.3**
  
  - [x] 7.7 Handle response completion
    - Add completed turn to history
    - Trigger next turn with opposite model
    - Reset generation state
    - _Requirements: 4.1, 4.2_
  
  - [x] 7.8 Handle errors
    - Display error message in UI
    - Preserve existing history
    - Attempt to continue with next turn if recoverable
    - _Requirements: 7.2, 7.4_
  
  - [ ]* 7.9 Write property test for error recovery
    - **Property 17: Generation errors allow continuation**
    - **Validates: Requirements 7.2**
  
  - [ ]* 7.10 Write property test for history integrity
    - **Property 18: Errors preserve history integrity**
    - **Validates: Requirements 7.4**
  
  - [x] 7.11 Handle stop command
    - Cancel ongoing generation
    - Transition to stopped state
    - Close Ollama connections
    - _Requirements: 5.1, 5.3, 5.4_
  
  - [ ]* 7.12 Write property test for stop command
    - **Property 13: Stop command halts generation**
    - **Validates: Requirements 5.1**
  
  - [ ]* 7.13 Write property test for connection cleanup
    - **Property 14: Stop closes Ollama connections**
    - **Validates: Requirements 5.3**
  
  - [ ]* 7.14 Write property test for history preservation
    - **Property 15: Stop preserves history**
    - **Validates: Requirements 5.4**
  
  - [x]* 7.15 Handle terminal resize
    - Update model width and height
    - Resize viewport component
    - _Requirements: 6.5_
  
  - [ ]* 7.16 Write property test for terminal resize
    - **Property 16: Terminal resize updates layout**
    - **Validates: Requirements 6.5**
  
  - [x] 7.17 Handle keyboard input
    - Process 'q' or Ctrl+C for stop
    - Handle Enter for topic submission
    - Pass other keys to text input component
    - _Requirements: 5.1, 1.2_

- [ ] 8. Implement Bubbletea View function
  - [ ] 8.1 Create view rendering for input state
    - Display welcome message
    - Render text input for topic
    - Show model names
    - _Requirements: 1.1, 2.1_
  
  - [ ]* 8.2 Write property test for topic display
    - **Property 3: Topic appears in rendered output**
    - **Validates: Requirements 1.4**
  
  - [ ] 8.3 Create view rendering for debating state
    - Render debate topic header
    - Display all turns with formatting
    - Show generation indicator for active model
    - Render viewport with scroll
    - _Requirements: 1.4, 3.1, 3.4_
  
  - [ ]* 8.4 Write property test for model labels
    - **Property 6: Responses display with model labels**
    - **Validates: Requirements 3.1**
  
  - [ ]* 8.5 Write property test for generation indicator
    - **Property 9: Generation indicator shows active model**
    - **Validates: Requirements 3.4**
  
  - [ ] 8.6 Create turn formatting function
    - Add model name label
    - Apply distinct styling per model (colors, borders)
    - Include timestamp
    - Format content with proper wrapping
    - _Requirements: 3.1, 3.2, 3.5_
  
  - [ ]* 8.7 Write property test for visual distinction
    - **Property 7: Models have distinct visual styling**
    - **Validates: Requirements 3.2**
  
  - [ ]* 8.8 Write property test for timestamps
    - **Property 10: Turns include timestamps**
    - **Validates: Requirements 3.5**
  
  - [ ] 8.9 Create view rendering for stopped state
    - Display final debate history
    - Show stop confirmation message
    - Provide exit instructions
    - _Requirements: 5.2, 5.4_
  
  - [ ] 8.10 Create view rendering for error state
    - Display error message prominently
    - Show existing debate history
    - Provide recovery or exit options
    - _Requirements: 7.1, 7.4_
  
  - [ ]* 8.11 Write unit tests for view rendering
    - Test each state renders correctly
    - Test turn formatting with various content
    - Test error message display
    - _Requirements: 1.1, 3.1, 7.1_

- [ ] 9. Implement main application entry point
  - [ ] 9.1 Create command-line flag parsing
    - Add flags for model1 and model2 names
    - Set defaults to mistral:7b and gemma3:4b
    - Add help text
    - _Requirements: 2.1, 2.5_
  
  - [ ] 9.2 Implement model validation at startup
    - Validate both models are available
    - Display error and exit if models not found
    - _Requirements: 2.2, 2.3_
  
  - [ ] 9.3 Initialize and run Bubbletea program
    - Create initial model with validated models
    - Configure Bubbletea program options
    - Run program and handle exit
    - _Requirements: 8.1, 8.2_
  
  - [ ]* 9.4 Write integration tests
    - Test complete flow from topic to multiple turns
    - Test stop command during generation
    - Test error recovery across turns
    - _Requirements: 1.1, 4.1, 5.1, 7.2_

- [ ] 10. Add styling with Lipgloss
  - [ ] 10.1 Define color scheme and styles
    - Create styles for model1 (e.g., blue theme)
    - Create styles for model2 (e.g., green theme)
    - Define styles for headers, errors, timestamps
    - _Requirements: 3.2_
  
  - [ ] 10.2 Apply styles to all UI components
    - Style turn rendering with model-specific colors
    - Style generation indicator
    - Style error messages
    - _Requirements: 3.1, 3.2, 7.1_

- [ ] 11. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 12. Add error recovery and resilience
  - [ ] 12.1 Implement retry logic for transient failures
    - Add exponential backoff for network errors
    - Track consecutive error count
    - Offer to stop after multiple failures
    - _Requirements: 7.2, 7.3_
  
  - [ ] 12.2 Add context size management
    - Limit history to last N turns if context grows too large
    - Provide warning when truncating context
    - _Requirements: 7.2_
  
  - [ ]* 12.3 Write unit tests for error recovery
    - Test retry logic with mock failures
    - Test consecutive error tracking
    - Test context truncation
    - _Requirements: 7.2_

- [ ] 13. Final polish and documentation
  - [ ] 13.1 Add README with usage instructions
    - Document installation steps
    - Provide usage examples
    - List requirements (Ollama, Go version)
    - _Requirements: 1.1, 2.1_
  
  - [ ] 13.2 Add code comments and documentation
    - Document all exported functions
    - Add package-level documentation
    - Include examples in comments
    - _Requirements: 8.1_
  
  - [ ] 13.3 Create example debate topics file
    - Provide sample topics for testing
    - Include various debate formats
    - _Requirements: 1.1_

- [ ] 14. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.
