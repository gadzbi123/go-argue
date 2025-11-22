package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// OllamaClient handles communication with the Ollama API
type OllamaClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewOllamaClient creates a new Ollama client with the specified base URL.
// If baseURL is empty, defaults to http://localhost:11434
func NewOllamaClient(baseURL string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// ListModels returns a list of available models from Ollama
func (c *OllamaClient) ListModels() ([]string, error) {
	url := fmt.Sprintf("%s/api/tags", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API returned status %d", resp.StatusCode)
	}

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse Ollama response: %w", err)
	}

	models := make([]string, len(result.Models))
	for i, model := range result.Models {
		models[i] = model.Name
	}

	return models, nil
}

// ValidateModel checks if a model is available in Ollama
func (c *OllamaClient) ValidateModel(modelName string) error {
	models, err := c.ListModels()
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	for _, model := range models {
		if model == modelName {
			return nil
		}
	}

	return fmt.Errorf("model '%s' not found in Ollama", modelName)
}

// GenerateRequest represents the request body for Ollama's generate API
type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// GenerateResponse represents a single response chunk from Ollama
type GenerateResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Context  []int  `json:"context,omitempty"`
}

// GenerateResponse generates a streaming response from a model.
// It returns two channels: one for response chunks and one for errors.
// The channels will be closed when the generation is complete or an error occurs.
func (c *OllamaClient) GenerateResponse(ctx context.Context, modelName, prompt string) (<-chan string, <-chan error) {
	responseChan := make(chan string)
	errorChan := make(chan error, 1)

	go func() {
		defer close(responseChan)
		defer close(errorChan)

		// Prepare the request
		reqBody := GenerateRequest{
			Model:  modelName,
			Prompt: prompt,
			Stream: true,
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			errorChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		url := fmt.Sprintf("%s/api/generate", c.baseURL)
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			errorChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		// Send the request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			errorChan <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errorChan <- fmt.Errorf("Ollama API returned status %d", resp.StatusCode)
			return
		}

		// Read the streaming response
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			// Check if context was cancelled
			select {
			case <-ctx.Done():
				errorChan <- ctx.Err()
				return
			default:
			}

			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var genResp GenerateResponse
			if err := json.Unmarshal(line, &genResp); err != nil {
				errorChan <- fmt.Errorf("failed to parse response: %w", err)
				return
			}

			// Send the response chunk
			if genResp.Response != "" {
				select {
				case responseChan <- genResp.Response:
				case <-ctx.Done():
					errorChan <- ctx.Err()
					return
				}
			}

			// Check if generation is complete
			if genResp.Done {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			errorChan <- fmt.Errorf("error reading response: %w", err)
			return
		}
	}()

	return responseChan, errorChan
}
