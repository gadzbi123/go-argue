# Design Document

## Overview

The AI Debate CLI is a terminal-based application built with Go and the Bubbletea TUI framework that orchestrates live debates between two AI models via the Ollama API. The application follows the Elm Architecture pattern, managing state through a single model and handling all interactions through a message-passing system. Users provide a debate topic, and two AI models engage in a continuous, alternating discussion until manually stopped.

The system is designed around concurrent operations: while one model generates a response, the UI remains responsive to user input, and the application prepares context for the next model's turn. The debate history is maintained in memory and displayed in real-time with clear visual distinction between the two participants.

## Architecture

The application follows a layered architecture with clear separation of concerns:

### Layers

1. **Presentation Layer (Bubbletea TUI)**
   - Handles all user interface rendering and input processing
   - Implements the Elm Architecture (Model-Update-View pattern)
   - Manages UI state and visual presentation
   - Processes keyboard events and terminal resizing

2. **Application Layer**
   - Orchestrates the debate flow and turn management
   - Maintains debate history and conversation context
   - Coordinates between UI and Ollama client
   - Handles concurrent model generation

3. **Integration Layer (Ollama Client)**
   - Communicates with Ollama API via HTTP
   - Manages streaming responses from AI models
   - Handles model validation and error recovery
   - Provides abstraction over Ollama API details

### Concurrency Model

The application uses Go's concurrency primitives to achieve responsive, non-blocking behavior:

- **Main goroutine**: Runs the Bubbletea event loop
- **Model generation goroutines**: One per AI model response, sends chunks back via channels
- **Message passing**: All communication between goroutines uses Bubbletea's Cmd system

## Components and Interfaces

### 1. Main Application (main.go)

**Responsibilities:**
- Parse command-line arguments for model names
- Initialize the Bubbletea program
- Handle graceful shutdown

**Interface:**
```go
func main()
func parseFlags() (model1, model2 string)
```

### 2. Bubbletea Model (model.go)

**Responsibilities:**
- Store application state (debate topic, history, current turn, UI state)
- Implement Init(), Update(), and View() methods
- Manage debate flow state machine

**State:**
```go
type debateModel struct {
    // Configuration
    model1Name    string
    model2Name    string
    ollamaClient  *OllamaClient
    
    // Debate state
    topic         string
    history       []Turn
    currentTurn   int  // 0 for model1, 1 for model2
    isGenerating  bool
    
    // UI state
    state         appState  // input, debating, stopped, error
    viewport      viewport.Model
    textInput     textinput.Model
    errorMsg      string
    
    // Dimensions
    width         int
    height        int
}

type Turn struct {
    ModelName  string
    Content    string
    Timestamp  time.Time
}

type appState int
const (
    stateInput appState = iota
    stateDebating
    stateStopped
    stateError
)
```

**Interface:**
```go
func (m debateModel) Init() tea.Cmd
func (m debateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m debateModel) View() string
```

### 3. Ollama Client (ollama.go)

**Responsibilities:**
- Validate model availability
- Generate streaming responses from models
- Handle API errors and retries
- Format prompts with debate context

**Interface:**
```go
type OllamaClient struct {
    baseURL    string
    httpClient *http.Client
}

func NewOllamaClient(baseURL string) *OllamaClient
func (c *OllamaClient) ValidateModel(modelName string) error
func (c *OllamaClient) GenerateResponse(ctx context.Context, modelName, prompt string) (<-chan string, <-chan error)
func (c *OllamaClient) ListModels() ([]string, error)
```

### 4. Prompt Builder (prompt.go)

**Responsibilities:**
- Construct debate prompts with full context
- Format conversation history for model consumption
- Assign debate positions to models

**Interface:**
```go
func BuildDebatePrompt(topic string, history []Turn, currentModel string, isFirstTurn bool) string
func FormatHistory(history []Turn) string
```

### 5. UI Renderer (view.go)

**Responsibilities:**
- Render different application states (input, debating, stopped, error)
- Format debate history with visual distinction
- Display generation indicators and status

**Interface:**
```go
func (m debateModel) renderInputView() string
func (m debateModel) renderDebateView() string
func (m debateModel) renderStoppedView() string
func (m debateModel) renderErrorView() string
func formatTurn(turn Turn, isModel1 bool) string
```

### 6. Message Types (messages.go)

**Responsibilities:**
- Define all message types for the Bubbletea Update function
- Encapsulate events and data for state transitions

**Types:**
```go
type topicSubmittedMsg struct{ topic string }
type responseChunkMsg struct{ chunk string }
type responseCompleteMsg struct{ fullResponse string }
type responseErrorMsg struct{ err error }
type nextTurnMsg struct{}
```

## Data Models

### Turn
Represents a single contribution to the debate from one model.

```go
type Turn struct {
    ModelName  string    // Name of the model that generated this turn
    Content    string    // The full response text
    Timestamp  time.Time // When the turn was completed
}
```

### Debate Context
The complete conversation context passed to models.

```go
type DebateContext struct {
    Topic   string
    History []Turn
}
```

### Ollama API Request/Response

```go
type GenerateRequest struct {
    Model  string `json:"model"`
    Prompt string `json:"prompt"`
    Stream bool   `json:"stream"`
}

type GenerateResponse struct {
    Model     string `json:"model"`
    Response  string `json:"response"`
    Done      bool   `json:"done"`
    Context   []int  `json:"context,omitempty"`
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property 1: Topic validation rejects empty input
*For any* input string, if it consists only of whitespace characters or is empty, the validation function should reject it and return an error.
**Validates: Requirements 1.2**

### Property 2: Valid topics initialize both models
*For any* non-empty topic string, initializing the debate should successfully prepare context for both AI models.
**Validates: Requirements 1.3**

### Property 3: Topic appears in rendered output
*For any* topic string, after initialization, the rendered UI view should contain that exact topic text.
**Validates: Requirements 1.4**

### Property 4: Model validation identifies availability
*For any* model name, the validation function should correctly determine whether that model exists in the Ollama instance.
**Validates: Requirements 2.2**

### Property 5: Valid models proceed to topic selection
*For any* pair of valid model names, after validation, the application state should transition to topic selection.
**Validates: Requirements 2.4**

### Property 6: Responses display with model labels
*For any* model response and model name, the rendered turn should include the model name as a visible label.
**Validates: Requirements 3.1**

### Property 7: Models have distinct visual styling
*For any* two different models, their rendered turns should have distinguishable visual attributes (colors, prefixes, or formatting).
**Validates: Requirements 3.2**

### Property 8: New responses append to history
*For any* debate state with N turns, adding a new response should result in N+1 turns in the history.
**Validates: Requirements 3.3**

### Property 9: Generation indicator shows active model
*For any* debate state where a model is generating, the rendered view should include an indicator showing which model is active.
**Validates: Requirements 3.4**

### Property 10: Turns include timestamps
*For any* turn in the debate history, the rendered output should include a timestamp.
**Validates: Requirements 3.5**

### Property 11: Turns alternate between models
*For any* sequence of turns in a debate, consecutive turns should alternate between model1 and model2 (no model should have two consecutive turns).
**Validates: Requirements 4.1**

### Property 12: Context includes full history
*For any* turn N in the debate, the prompt sent to the model should include the topic and all previous turns (0 through N-1) with correct model attribution.
**Validates: Requirements 4.2, 4.3, 9.1**

### Property 13: Stop command halts generation
*For any* debate state, issuing a stop command should prevent any new turns from being added to the history.
**Validates: Requirements 5.1**

### Property 14: Stop closes Ollama connections
*For any* active debate, stopping should result in all HTTP connections to Ollama being closed.
**Validates: Requirements 5.3**

### Property 15: Stop preserves history
*For any* debate with N turns, stopping the debate should leave all N turns intact and accessible.
**Validates: Requirements 5.4**

### Property 16: Terminal resize updates layout
*For any* terminal dimensions (width, height), resizing should update the viewport dimensions to match.
**Validates: Requirements 6.5**

### Property 17: Generation errors allow continuation
*For any* generation error on turn N, the system should be able to proceed to turn N+1 with the other model.
**Validates: Requirements 7.2**

### Property 18: Errors preserve history integrity
*For any* error condition, the existing debate history should remain unchanged (no turns added, removed, or modified).
**Validates: Requirements 7.4**

### Property 19: Context formatting shows model attribution
*For any* formatted context string, each turn should be clearly marked with which model generated it.
**Validates: Requirements 9.2**

### Property 20: Prompts include debate instructions
*For any* generated prompt, it should contain instructions for the model to engage in debate (taking positions, responding to arguments).
**Validates: Requirements 9.4**

## Error Handling

### Error Categories

1. **Initialization Errors**
   - Invalid or unavailable model names
   - Ollama service not running or unreachable
   - Network connectivity issues

2. **Runtime Errors**
   - Model generation failures (timeout, API errors)
   - Streaming interruptions
   - Context size exceeded

3. **User Input Errors**
   - Empty or invalid topic
   - Invalid commands

### Error Handling Strategies

**Initialization Errors:**
- Display clear error messages with actionable guidance
- Validate models before starting debate
- Provide fallback to default models if available
- Exit gracefully if Ollama is unreachable

**Runtime Errors:**
- Log errors with context (model name, turn number)
- Attempt to continue with next turn
- Display error in UI without disrupting history
- After 3 consecutive failures, offer to stop debate

**User Input Errors:**
- Validate input before processing
- Display inline error messages
- Re-prompt for valid input
- Preserve application state

### Error Recovery

```go
type ErrorRecovery struct {
    MaxRetries        int
    RetryDelay        time.Duration
    ConsecutiveErrors int
}

func (e *ErrorRecovery) ShouldRetry() bool {
    return e.ConsecutiveErrors < e.MaxRetries
}

func (e *ErrorRecovery) RecordError() {
    e.ConsecutiveErrors++
}

func (e *ErrorRecovery) Reset() {
    e.ConsecutiveErrors = 0
}
```

## Testing Strategy

### Unit Testing

The application will use Go's standard `testing` package for unit tests. Unit tests will focus on:

**Core Logic:**
- Prompt building with various history lengths
- Turn alternation logic
- Input validation (empty topics, whitespace)
- Context formatting with model attribution
- State transitions (input → debating → stopped)

**Error Handling:**
- Invalid model names
- Empty responses from Ollama
- Network timeout scenarios
- Malformed API responses

**UI Components:**
- Turn formatting with timestamps
- Visual distinction between models
- Viewport scrolling behavior
- Error message display

### Property-Based Testing

The application will use **gopter** (https://github.com/leanovate/gopter) for property-based testing. Each property-based test will run a minimum of 100 iterations to ensure thorough coverage.

**Configuration:**
```go
parameters := gopter.DefaultTestParameters()
parameters.MinSuccessfulTests = 100
```

**Tagging Convention:**
Each property-based test must include a comment tag in this exact format:
```go
// Feature: ai-debate-cli, Property N: <property description>
```

**Property Test Coverage:**
- Property 1: Topic validation with random strings (empty, whitespace, valid)
- Property 2: Model initialization with random valid topics
- Property 3: UI rendering contains topic text
- Property 4: Model validation with random model names
- Property 5: State transitions with valid model pairs
- Property 6: Turn rendering includes model labels
- Property 7: Visual distinction between model outputs
- Property 8: History growth with sequential turns
- Property 9: Generation indicator presence during active generation
- Property 10: Timestamp presence in all turns
- Property 11: Turn alternation across random debate lengths
- Property 12: Context completeness with varying history sizes
- Property 13: Stop command effectiveness at random debate states
- Property 14: Connection cleanup after stop
- Property 15: History preservation after stop
- Property 16: Layout adaptation to random terminal sizes
- Property 17: Error recovery and continuation
- Property 18: History integrity under error conditions
- Property 19: Model attribution in formatted context
- Property 20: Debate instructions in all prompts

**Test Generators:**
- Random non-empty strings for topics
- Random model names (valid and invalid)
- Random debate histories of varying lengths
- Random terminal dimensions
- Random error injection points

### Integration Testing

Integration tests will verify end-to-end flows:
- Complete debate flow from topic input to multiple turns
- Model switching and context passing
- Stop command during active generation
- Error recovery across multiple turns

### Manual Testing

Due to the interactive nature of the TUI:
- Visual verification of styling and layout
- Responsiveness during model generation
- Terminal resize behavior
- Keyboard input handling

## Implementation Notes

### Ollama API Integration

The application will use the `/api/generate` endpoint with streaming enabled:

```go
POST http://localhost:11434/api/generate
{
    "model": "mistral:7b",
    "prompt": "<debate prompt>",
    "stream": true
}
```

Responses arrive as newline-delimited JSON:
```json
{"model":"mistral:7b","response":"The","done":false}
{"model":"mistral:7b","response":" sky","done":false}
{"model":"mistral:7b","response":" is","done":false}
...
{"model":"mistral:7b","response":"","done":true}
```

### Prompt Engineering

Debate prompts will follow this structure:

```
You are participating in a debate on the topic: "<TOPIC>"

You are <MODEL_NAME>. Your role is to present arguments and respond to your opponent's points.

Previous discussion:
<HISTORY>

Provide your next argument or response. Be thoughtful, specific, and engage directly with the previous points made.
```

For the first turn, the prompt will assign initial positions (pro/con or different perspectives).

### Bubbletea Message Flow

```
User Input → topicSubmittedMsg → Start Generation
Generation → responseChunkMsg (multiple) → Update View
Generation Complete → responseCompleteMsg → nextTurnMsg → Start Next Generation
User Stop → stopDebateMsg → Cleanup
Error → responseErrorMsg → Display Error → nextTurnMsg (if recoverable)
```

### Performance Considerations

- **Streaming**: Use streaming API to display responses as they generate
- **Buffering**: Buffer response chunks to avoid excessive UI updates
- **Context Management**: Limit context size if debate becomes very long (e.g., last 20 turns)
- **Goroutine Cleanup**: Ensure all goroutines are properly cancelled on stop

### Dependencies

```go
require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/charmbracelet/bubbles v0.17.1
    github.com/leanovate/gopter v0.2.9  // for property-based testing
)
```

## Future Enhancements

- Save debate transcripts to file
- Support for more than two models
- Configurable debate formats (structured, free-form, timed)
- Syntax highlighting for code in responses
- Model performance metrics (response time, token count)
- Resume previous debates
