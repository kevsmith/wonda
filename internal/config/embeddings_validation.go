package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ValidateEmbeddingModel checks if the required embedding model is available
// from the given provider by making a test embedding request.
func ValidateEmbeddingModel(provider *Provider) error {
	if provider == nil {
		return fmt.Errorf("provider is nil")
	}

	baseURL := provider.BaseURL
	if baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}

	// Try different endpoint styles:
	// 1. OpenAI-compatible: base_url/v1/embeddings (LM Studio, OpenAI)
	// 2. Ollama-style: base_url/api/embeddings (Ollama)
	// 3. OpenAI without /v1: base_url/embeddings (some providers)

	// If base URL ends with /v1/, try OpenAI-compatible first
	if len(provider.BaseURL) >= 3 && provider.BaseURL[len(provider.BaseURL)-3:] == "/v1" {
		if err := tryEmbedding(baseURL+"embeddings", provider, RequiredEmbeddingModel); err == nil {
			return nil
		}
	}

	// Try Ollama-style (strip /v1 if present)
	ollamaURL := provider.BaseURL
	if len(ollamaURL) >= 3 && ollamaURL[len(ollamaURL)-3:] == "/v1" {
		ollamaURL = ollamaURL[:len(ollamaURL)-3]
	}
	if ollamaURL[len(ollamaURL)-1] != '/' {
		ollamaURL += "/"
	}
	if err := tryEmbedding(ollamaURL+"api/embeddings", provider, RequiredEmbeddingModel); err == nil {
		return nil
	}

	// Try plain /embeddings without /v1
	if err := tryEmbedding(ollamaURL+"embeddings", provider, RequiredEmbeddingModel); err == nil {
		return nil
	}

	return fmt.Errorf("embedding model '%s' not available from provider '%s'\n\nTo install:\n  ollama pull %s",
		RequiredEmbeddingModel, provider.Name, RequiredEmbeddingModel)
}

// tryEmbedding attempts to generate a test embedding from the given endpoint.
func tryEmbedding(url string, provider *Provider, model string) error {
	// Create request body (try both "prompt" for Ollama and "input" for OpenAI)
	reqBody := map[string]interface{}{
		"model":  model,
		"prompt": "test", // Ollama format
		"input":  "test", // OpenAI format
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	// Add API key if present
	if provider.APIKey != nil && *provider.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+*provider.APIKey)
	}

	// Make request with timeout
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to embedding endpoint: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("embedding request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response to validate structure
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode embedding response: %w", err)
	}

	// Check for embedding data (Ollama format: "embedding", OpenAI format: "data[0].embedding")
	if embedding, ok := result["embedding"].([]interface{}); ok {
		if len(embedding) != RequiredEmbeddingDimensions {
			return fmt.Errorf("unexpected embedding dimensions: got %d, expected %d", len(embedding), RequiredEmbeddingDimensions)
		}
		return nil
	}

	// Check OpenAI format
	if data, ok := result["data"].([]interface{}); ok && len(data) > 0 {
		if first, ok := data[0].(map[string]interface{}); ok {
			if embedding, ok := first["embedding"].([]interface{}); ok {
				if len(embedding) != RequiredEmbeddingDimensions {
					return fmt.Errorf("unexpected embedding dimensions: got %d, expected %d", len(embedding), RequiredEmbeddingDimensions)
				}
				return nil
			}
		}
	}

	return fmt.Errorf("response missing embedding data")
}
