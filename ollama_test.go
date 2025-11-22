package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestNewOllamaClient tests client initialization
func TestNewOllamaClient(t *testing.T) {
	t.Run("with custom URL", func(t *testing.T) {
		client := NewOllamaClient("http://custom:8080")
		if client.baseURL != "http://custom:8080" {
			t.Errorf("Expected baseURL to be http://custom:8080, got %s", client.baseURL)
		}
	})

	t.Run("with empty URL defaults to localhost", func(t *testing.T) {
		client := NewOllamaClient("")
		if client.baseURL != "http://localhost:11434" {
			t.Errorf("Expected default baseURL to be http://localhost:11434, got %s", client.baseURL)
		}
	})
}

// TestListModels_Success tests successful model listing
func TestListModels_Success(t *testing.T) {
	// Create a test server that returns a valid response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request path
		if r.URL.Path != "/api/tags" {
			t.Errorf("Expected path /api/tags, got %s", r.URL.Path)
		}

		// Return a valid response
		response := map[string]interface{}{
			"models": []map[string]string{
				{"name": "mistral:7b"},
				{"name": "gemma3:4b"},
				{"name": "llama2:13b"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL)
	models, err := client.ListModels()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedModels := []string{"mistral:7b", "gemma3:4b", "llama2:13b"}
	if len(models) != len(expectedModels) {
		t.Fatalf("Expected %d models, got %d", len(expectedModels), len(models))
	}

	for i, expected := range expectedModels {
		if models[i] != expected {
			t.Errorf("Expected model %s at index %d, got %s", expected, i, models[i])
		}
	}
}

// TestListModels_NetworkError tests handling of network failures
func TestListModels_NetworkError(t *testing.T) {
	// Use an invalid URL to simulate network failure
	client := NewOllamaClient("http://invalid-host-that-does-not-exist:99999")
	_, err := client.ListModels()

	if err == nil {
		t.Fatal("Expected error for network failure, got nil")
	}

	if !strings.Contains(err.Error(), "failed to connect to Ollama") {
		t.Errorf("Expected error message about connection failure, got: %v", err)
	}
}

// TestListModels_NonOKStatus tests handling of non-200 status codes
func TestListModels_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL)
	_, err := client.ListModels()

	if err == nil {
		t.Fatal("Expected error for non-OK status, got nil")
	}

	if !strings.Contains(err.Error(), "status 500") {
		t.Errorf("Expected error message about status 500, got: %v", err)
	}
}

// TestListModels_InvalidJSON tests handling of malformed JSON responses
func TestListModels_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json {{{"))
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL)
	_, err := client.ListModels()

	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}

	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("Expected error message about parsing failure, got: %v", err)
	}
}

// TestValidateModel_Success tests successful model validation
func TestValidateModel_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"models": []map[string]string{
				{"name": "mistral:7b"},
				{"name": "gemma3:4b"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL)
	err := client.ValidateModel("mistral:7b")

	if err != nil {
		t.Errorf("Expected no error for valid model, got %v", err)
	}
}

// TestValidateModel_NotFound tests validation of non-existent model
func TestValidateModel_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"models": []map[string]string{
				{"name": "mistral:7b"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL)
	err := client.ValidateModel("nonexistent:model")

	if err == nil {
		t.Fatal("Expected error for non-existent model, got nil")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected error message about model not found, got: %v", err)
	}
}

// TestGenerateResponse_RequestFormatting tests HTTP request formatting
func TestGenerateResponse_RequestFormatting(t *testing.T) {
	requestReceived := false
	var receivedRequest GenerateRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true

		// Verify HTTP method
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Verify path
		if r.URL.Path != "/api/generate" {
			t.Errorf("Expected path /api/generate, got %s", r.URL.Path)
		}

		// Verify Content-Type header
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		// Parse and verify request body
		if err := json.NewDecoder(r.Body).Decode(&receivedRequest); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Send a simple response
		response := GenerateResponse{
			Model:    receivedRequest.Model,
			Response: "test response",
			Done:     true,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL)
	ctx := context.Background()
	modelName := "mistral:7b"
	prompt := "Test prompt"

	responseChan, errorChan := client.GenerateResponse(ctx, modelName, prompt)

	// Consume channels
	for range responseChan {
	}
	if err := <-errorChan; err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !requestReceived {
		t.Fatal("Request was not received by server")
	}

	// Verify request formatting
	if receivedRequest.Model != modelName {
		t.Errorf("Expected model %s, got %s", modelName, receivedRequest.Model)
	}
	if receivedRequest.Prompt != prompt {
		t.Errorf("Expected prompt %s, got %s", prompt, receivedRequest.Prompt)
	}
	if !receivedRequest.Stream {
		t.Error("Expected Stream to be true")
	}
}

// TestGenerateResponse_StreamingParsing tests parsing of streaming responses
func TestGenerateResponse_StreamingParsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send multiple response chunks as newline-delimited JSON
		chunks := []GenerateResponse{
			{Model: "mistral:7b", Response: "Hello", Done: false},
			{Model: "mistral:7b", Response: " world", Done: false},
			{Model: "mistral:7b", Response: "!", Done: false},
			{Model: "mistral:7b", Response: "", Done: true},
		}

		for _, chunk := range chunks {
			json.NewEncoder(w).Encode(chunk)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL)
	ctx := context.Background()

	responseChan, errorChan := client.GenerateResponse(ctx, "mistral:7b", "test")

	// Collect all response chunks
	var chunks []string
	for chunk := range responseChan {
		chunks = append(chunks, chunk)
	}

	if err := <-errorChan; err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify we received the correct chunks
	expectedChunks := []string{"Hello", " world", "!"}
	if len(chunks) != len(expectedChunks) {
		t.Fatalf("Expected %d chunks, got %d", len(expectedChunks), len(chunks))
	}

	for i, expected := range expectedChunks {
		if chunks[i] != expected {
			t.Errorf("Expected chunk %d to be %s, got %s", i, expected, chunks[i])
		}
	}
}

// TestGenerateResponse_NetworkError tests handling of network errors during generation
func TestGenerateResponse_NetworkError(t *testing.T) {
	// Use an invalid URL to simulate network failure
	client := NewOllamaClient("http://invalid-host-that-does-not-exist:99999")
	ctx := context.Background()

	responseChan, errorChan := client.GenerateResponse(ctx, "mistral:7b", "test")

	// Consume response channel (should be empty)
	for range responseChan {
		t.Error("Did not expect any response chunks")
	}

	// Check for error
	err := <-errorChan
	if err == nil {
		t.Fatal("Expected error for network failure, got nil")
	}

	if !strings.Contains(err.Error(), "failed to send request") {
		t.Errorf("Expected error message about request failure, got: %v", err)
	}
}

// TestGenerateResponse_NonOKStatus tests handling of non-200 status during generation
func TestGenerateResponse_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL)
	ctx := context.Background()

	responseChan, errorChan := client.GenerateResponse(ctx, "mistral:7b", "test")

	// Consume response channel
	for range responseChan {
		t.Error("Did not expect any response chunks")
	}

	// Check for error
	err := <-errorChan
	if err == nil {
		t.Fatal("Expected error for non-OK status, got nil")
	}

	if !strings.Contains(err.Error(), "status 503") {
		t.Errorf("Expected error message about status 503, got: %v", err)
	}
}

// TestGenerateResponse_ContextCancellation tests context cancellation during generation
func TestGenerateResponse_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send a chunk, then delay
		chunk := GenerateResponse{Model: "mistral:7b", Response: "Start", Done: false}
		json.NewEncoder(w).Encode(chunk)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		// Wait a bit before sending more
		time.Sleep(100 * time.Millisecond)

		// Try to send another chunk (should fail due to context cancellation)
		chunk = GenerateResponse{Model: "mistral:7b", Response: "End", Done: true}
		json.NewEncoder(w).Encode(chunk)
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL)
	ctx, cancel := context.WithCancel(context.Background())

	responseChan, errorChan := client.GenerateResponse(ctx, "mistral:7b", "test")

	// Read first chunk
	firstChunk := <-responseChan
	if firstChunk != "Start" {
		t.Errorf("Expected first chunk to be 'Start', got %s", firstChunk)
	}

	// Cancel context
	cancel()

	// Consume remaining channels
	for range responseChan {
	}

	// Check for cancellation error
	err := <-errorChan
	if err == nil {
		t.Fatal("Expected context cancellation error, got nil")
	}

	// The error may be wrapped, so check if it contains context.Canceled
	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected context cancellation error, got: %v", err)
	}
}

// TestGenerateResponse_InvalidJSON tests handling of malformed JSON in streaming response
func TestGenerateResponse_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send valid chunk first
		chunk := GenerateResponse{Model: "mistral:7b", Response: "Valid", Done: false}
		json.NewEncoder(w).Encode(chunk)

		// Send invalid JSON
		w.Write([]byte("invalid json {{{"))
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL)
	ctx := context.Background()

	responseChan, errorChan := client.GenerateResponse(ctx, "mistral:7b", "test")

	// Read the valid chunk
	firstChunk := <-responseChan
	if firstChunk != "Valid" {
		t.Errorf("Expected first chunk to be 'Valid', got %s", firstChunk)
	}

	// Consume remaining response channel
	for range responseChan {
	}

	// Check for parsing error
	err := <-errorChan
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}

	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("Expected error message about parsing failure, got: %v", err)
	}
}
