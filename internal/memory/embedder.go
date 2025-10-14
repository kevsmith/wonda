package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/poiesic/wonda/internal/config"
)

// Embedder generates vector embeddings from text.
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// OllamaEmbedder implements Embedder using Ollama's API.
type OllamaEmbedder struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaEmbedder creates a new Ollama embedder.
// Despite the name, this works with both Ollama and OpenAI-compatible endpoints.
func NewOllamaEmbedder(provider *config.Provider) *OllamaEmbedder {
	baseURL := provider.BaseURL
	if baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}

	// Detect endpoint style:
	// - If base URL ends with /v1/, use OpenAI-compatible: /v1/embeddings (LM Studio)
	// - Otherwise use Ollama-style: /api/embeddings
	var embeddingURL string
	if len(provider.BaseURL) >= 3 && provider.BaseURL[len(provider.BaseURL)-3:] == "/v1" {
		embeddingURL = baseURL + "embeddings" // OpenAI-compatible
	} else {
		embeddingURL = baseURL + "api/embeddings" // Ollama-style
	}

	return &OllamaEmbedder{
		baseURL: embeddingURL,
		model:   config.RequiredEmbeddingModel,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Embed generates an embedding vector for the given text.
func (e *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	// Detect format based on URL:
	// - /v1/embeddings = OpenAI format (use "input")
	// - /api/embeddings = Ollama format (use "prompt")
	var textField string
	if strings.Contains(e.baseURL, "/v1/") {
		textField = "input" // OpenAI/LM Studio format
	} else {
		textField = "prompt" // Ollama format
	}

	embedding, err := e.tryEmbed(ctx, text, textField)
	if err == nil {
		return embedding, nil
	}

	// If the detected format fails, try the other one as fallback
	fallbackField := "input"
	if textField == "input" {
		fallbackField = "prompt"
	}

	embedding, err = e.tryEmbed(ctx, text, fallbackField)
	if err == nil {
		return embedding, nil
	}

	return nil, fmt.Errorf("failed to generate embedding with both Ollama and OpenAI formats: %w", err)
}

// tryEmbed attempts to generate an embedding using the specified field name for text.
func (e *OllamaEmbedder) tryEmbed(ctx context.Context, text string, textField string) ([]float32, error) {
	// Create request body
	reqBody := map[string]interface{}{
		"model":   e.model,
		textField: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call embedding API: %w", err)
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response - handle both Ollama format and OpenAI format
	var ollamaResult struct {
		Embedding []float32 `json:"embedding"`
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try Ollama format first
	if err := json.Unmarshal(bodyBytes, &ollamaResult); err == nil && len(ollamaResult.Embedding) > 0 {
		if len(ollamaResult.Embedding) != config.RequiredEmbeddingDimensions {
			return nil, fmt.Errorf("unexpected embedding dimensions: got %d, expected %d",
				len(ollamaResult.Embedding), config.RequiredEmbeddingDimensions)
		}
		return ollamaResult.Embedding, nil
	}

	// Try OpenAI format
	var openaiResult struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(bodyBytes, &openaiResult); err == nil && len(openaiResult.Data) > 0 {
		embedding := openaiResult.Data[0].Embedding
		if len(embedding) != config.RequiredEmbeddingDimensions {
			return nil, fmt.Errorf("unexpected embedding dimensions: got %d, expected %d",
				len(embedding), config.RequiredEmbeddingDimensions)
		}
		return embedding, nil
	}

	return nil, fmt.Errorf("no embedding returned in response")
}
