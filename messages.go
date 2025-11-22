package main

// topicSubmittedMsg is sent when the user submits a topic
type topicSubmittedMsg struct {
	topic string
}

// responseChunkMsg is sent when a response chunk arrives
type responseChunkMsg struct {
	chunk        string
	responseChan <-chan string
	errorChan    <-chan error
}

// responseCompleteMsg is sent when a response is complete
type responseCompleteMsg struct {
	fullResponse string
}

// responseErrorMsg is sent when an error occurs during generation
type responseErrorMsg struct {
	err error
}

// nextTurnMsg is sent to trigger the next turn
type nextTurnMsg struct{}

// stopDebateMsg is sent when the user stops the debate
type stopDebateMsg struct{}
