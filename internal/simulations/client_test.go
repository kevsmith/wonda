package simulations

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/poiesic/wonda/internal/config"
)

func TestNewClient(t *testing.T) {
	t.Run("creates OpenAI client for non-Anthropic URL", func(t *testing.T) {
		provider := &config.Provider{
			Name:    "openai",
			BaseURL: "https://api.openai.com/v1",
		}
		model := &config.Model{
			Name:     "gpt-4",
			Provider: "openai",
			ThinkingParser: &config.ThinkingParserConfig{
				Type: config.ThinkingParserNone,
			},
		}

		client, err := NewClient(provider, model)
		require.NoError(t, err)
		assert.IsType(t, &OpenAIClient{}, client)
	})

	t.Run("creates Anthropic client for Anthropic URL", func(t *testing.T) {
		provider := &config.Provider{
			Name:    "anthropic",
			BaseURL: "https://api.anthropic.com/v1",
		}
		model := &config.Model{
			Name:     "claude-3-5-sonnet-20241022",
			Provider: "anthropic",
			ThinkingParser: &config.ThinkingParserConfig{
				Type:      config.ThinkingParserOutOfBand,
				FieldPath: "thinking",
			},
		}

		client, err := NewClient(provider, model)
		require.NoError(t, err)
		assert.IsType(t, &AnthropicClient{}, client)
	})

	t.Run("returns error for nil provider", func(t *testing.T) {
		model := &config.Model{
			Name:     "test",
			Provider: "test",
		}

		_, err := NewClient(nil, model)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider cannot be nil")
	})

	t.Run("returns error for nil model", func(t *testing.T) {
		provider := &config.Provider{
			Name:    "test",
			BaseURL: "http://test",
		}

		_, err := NewClient(provider, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model cannot be nil")
	})

	t.Run("returns error when model provider doesn't match", func(t *testing.T) {
		provider := &config.Provider{
			Name:    "openai",
			BaseURL: "https://api.openai.com/v1",
		}
		model := &config.Model{
			Name:     "claude-3-5-sonnet-20241022",
			Provider: "anthropic", // Mismatch!
		}

		_, err := NewClient(provider, model)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not match")
	})
}

func TestOpenAIClient_Chat(t *testing.T) {
	t.Run("sends basic chat request", func(t *testing.T) {
		// Create mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.URL.Path, "/chat/completions")

			// Verify request body
			var reqBody map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			require.NoError(t, err)
			assert.Equal(t, "gpt-4", reqBody["model"])

			// Send response
			resp := map[string]interface{}{
				"id":      "chatcmpl-123",
				"object":  "chat.completion",
				"created": 1677652288,
				"model":   "gpt-4",
				"choices": []map[string]interface{}{
					{
						"index": 0,
						"message": map[string]interface{}{
							"role":    "assistant",
							"content": "Hello! How can I help you?",
						},
						"finish_reason": "stop",
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		// Create client
		apiKey := "test-key"
		provider := &config.Provider{
			Name:    "openai",
			BaseURL: server.URL,
			APIKey:  &apiKey,
		}
		model := &config.Model{
			Name:     "gpt-4",
			Provider: "openai",
			ThinkingParser: &config.ThinkingParserConfig{
				Type: config.ThinkingParserNone,
			},
		}

		client, err := NewClient(provider, model)
		require.NoError(t, err)

		// Send request
		resp, err := client.Chat(context.Background(), ChatRequest{
			Messages: []Message{
				{Role: "user", Content: "Hello"},
			},
		})

		require.NoError(t, err)
		assert.Equal(t, "Hello! How can I help you?", resp.Message)
		assert.Equal(t, "", resp.Thinking)
	})

	t.Run("extracts in-band thinking", func(t *testing.T) {
		// Create mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]interface{}{
				"id":      "chatcmpl-123",
				"object":  "chat.completion",
				"created": 1677652288,
				"model":   "qwq-32b",
				"choices": []map[string]interface{}{
					{
						"index": 0,
						"message": map[string]interface{}{
							"role":    "assistant",
							"content": "<think>Let me analyze this problem...</think>The answer is 42.",
						},
						"finish_reason": "stop",
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		// Create client with in-band parser
		provider := &config.Provider{
			Name:    "ollama",
			BaseURL: server.URL,
		}
		model := &config.Model{
			Name:     "qwq-32b",
			Provider: "ollama",
			ThinkingParser: &config.ThinkingParserConfig{
				Type:           config.ThinkingParserInBand,
				StartDelimiter: "<think>",
				EndDelimiter:   "</think>",
			},
		}

		client, err := NewClient(provider, model)
		require.NoError(t, err)

		// Send request
		resp, err := client.Chat(context.Background(), ChatRequest{
			Messages: []Message{
				{Role: "user", Content: "What is the meaning of life?"},
			},
		})

		require.NoError(t, err)
		assert.Equal(t, "The answer is 42.", resp.Message)
		assert.Equal(t, "Let me analyze this problem...", resp.Thinking)
	})
}

func TestAnthropicClient_Chat(t *testing.T) {
	t.Run("sends basic chat request", func(t *testing.T) {
		// Create mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.URL.Path, "/messages")

			// Verify request body
			var reqBody map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			require.NoError(t, err)
			assert.Equal(t, "claude-3-5-sonnet-20241022", reqBody["model"])

			// Send response
			resp := map[string]interface{}{
				"id":   "msg_123",
				"type": "message",
				"role": "assistant",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "Hello! I'm Claude. How can I help you?",
					},
				},
				"model":        "claude-3-5-sonnet-20241022",
				"stop_reason":  "end_turn",
				"stop_sequence": nil,
				"usage": map[string]interface{}{
					"input_tokens":  10,
					"output_tokens": 20,
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		// Create client
		apiKey := "test-key"
		provider := &config.Provider{
			Name:    "anthropic",
			BaseURL: server.URL,
			APIKey:  &apiKey,
		}
		model := &config.Model{
			Name:     "claude-3-5-sonnet-20241022",
			Provider: "anthropic",
			ThinkingParser: &config.ThinkingParserConfig{
				Type:      config.ThinkingParserOutOfBand,
				FieldPath: "thinking",
			},
		}

		client, err := NewClient(provider, model)
		require.NoError(t, err)

		// Send request
		resp, err := client.Chat(context.Background(), ChatRequest{
			Messages: []Message{
				{Role: "user", Content: "Hello"},
			},
		})

		require.NoError(t, err)
		assert.Equal(t, "Hello! I'm Claude. How can I help you?", resp.Message)
		assert.Equal(t, "", resp.Thinking) // No thinking in this response
	})

	t.Run("extracts extended thinking", func(t *testing.T) {
		// Create mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]interface{}{
				"id":   "msg_123",
				"type": "message",
				"role": "assistant",
				"content": []map[string]interface{}{
					{
						"type":      "thinking",
						"thinking":  "Let me carefully consider this question...",
						"signature": "abc123",
					},
					{
						"type": "text",
						"text": "Based on my analysis, the answer is 42.",
					},
				},
				"model":        "claude-3-5-sonnet-20241022",
				"stop_reason":  "end_turn",
				"stop_sequence": nil,
				"usage": map[string]interface{}{
					"input_tokens":  10,
					"output_tokens": 50,
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		// Create client
		apiKey := "test-key"
		provider := &config.Provider{
			Name:    "anthropic",
			BaseURL: server.URL,
			APIKey:  &apiKey,
		}
		model := &config.Model{
			Name:     "claude-3-5-sonnet-20241022",
			Provider: "anthropic",
			ThinkingParser: &config.ThinkingParserConfig{
				Type:      config.ThinkingParserOutOfBand,
				FieldPath: "thinking",
			},
		}

		client, err := NewClient(provider, model)
		require.NoError(t, err)

		// Send request
		resp, err := client.Chat(context.Background(), ChatRequest{
			Messages: []Message{
				{Role: "user", Content: "What is the meaning of life?"},
			},
		})

		require.NoError(t, err)
		assert.Equal(t, "Based on my analysis, the answer is 42.", resp.Message)
		assert.Equal(t, "Let me carefully consider this question...", resp.Thinking)
	})

	t.Run("handles system messages", func(t *testing.T) {
		var receivedSystem string

		// Create mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Capture system prompt
			var reqBody map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			require.NoError(t, err)
			if sys, ok := reqBody["system"].(string); ok {
				receivedSystem = sys
			}

			resp := map[string]interface{}{
				"id":   "msg_123",
				"type": "message",
				"role": "assistant",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "Understood.",
					},
				},
				"model":       "claude-3-5-sonnet-20241022",
				"stop_reason": "end_turn",
				"usage": map[string]interface{}{
					"input_tokens":  10,
					"output_tokens": 5,
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		// Create client
		apiKey := "test-key"
		provider := &config.Provider{
			Name:    "anthropic",
			BaseURL: server.URL,
			APIKey:  &apiKey,
		}
		model := &config.Model{
			Name:     "claude-3-5-sonnet-20241022",
			Provider: "anthropic",
		}

		client, err := NewClient(provider, model)
		require.NoError(t, err)

		// Send request with system message
		_, err = client.Chat(context.Background(), ChatRequest{
			Messages: []Message{
				{Role: "system", Content: "You are a helpful assistant."},
				{Role: "user", Content: "Hello"},
			},
		})

		require.NoError(t, err)
		assert.Equal(t, "You are a helpful assistant.", receivedSystem)
	})
}
