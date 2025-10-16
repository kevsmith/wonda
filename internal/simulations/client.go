package simulations

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/poiesic/wonda/internal/config"
)

// Message represents a single message in a conversation.
type Message struct {
	Role    string // "user", "assistant", or "system"
	Content string
}

// ChatRequest represents a request to generate a chat completion.
type ChatRequest struct {
	Messages []Message
	Model    string
	Tools    []map[string]interface{} // Tool definitions for the LLM
}

// ChatResponse represents the response from a chat completion.
type ChatResponse struct {
	Message   string     // The active/spoken content
	Thinking  string     // Internal reasoning (may be empty if model doesn't support it)
	ToolCalls []ToolCall // Tools the LLM wants to invoke
}

// ToolCall represents a request from the LLM to invoke a tool.
type ToolCall struct {
	ID        string                 // Unique ID for this call (from LLM API)
	Name      string                 // Tool name
	Arguments map[string]interface{} // Tool arguments
}

// Client is the interface for LLM clients.
// Implementations must be stateless - the caller manages conversation history.
type Client interface {
	// Chat sends a chat completion request and returns the response.
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
}

// ResponseParser extracts thinking/reasoning from model responses.
type ResponseParser interface {
	// Parse extracts the message and thinking from a raw response.
	// For models without thinking, thinking should be empty string.
	Parse(response string) (message string, thinking string)
}

// NewClient creates a Client implementation based on the provider and model configuration.
// It auto-detects the appropriate client type based on the provider's base URL.
func NewClient(provider *config.Provider, model *config.Model) (Client, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider cannot be nil")
	}
	if model == nil {
		return nil, fmt.Errorf("model cannot be nil")
	}

	// Validate that the model references this provider
	if model.Provider != provider.Name {
		return nil, fmt.Errorf("model provider '%s' does not match provider name '%s'", model.Provider, provider.Name)
	}

	// Create response parser
	parser, err := newResponseParser(model.ThinkingParser)
	if err != nil {
		return nil, fmt.Errorf("failed to create response parser: %w", err)
	}

	// Detect client type based on provider name or URL
	// Check provider name first for explicit configuration
	if strings.ToLower(provider.Name) == "anthropic" {
		return newAnthropicClient(provider, model, parser)
	}

	// Check URL for anthropic.com
	baseURL := strings.ToLower(provider.BaseURL)
	if strings.Contains(baseURL, "anthropic.com") {
		return newAnthropicClient(provider, model, parser)
	}

	// Default to OpenAI-compatible client
	return newOpenAIClient(provider, model, parser)
}

// newResponseParser creates a ResponseParser based on the thinking parser configuration.
func newResponseParser(cfg *config.ThinkingParserConfig) (ResponseParser, error) {
	if cfg == nil {
		slog.Info("no thinking parser configured, thinking will not be extracted")
		return &NoOpParser{}, nil
	}

	switch cfg.Type {
	case config.ThinkingParserNone:
		slog.Info("thinking parser disabled")
		return &NoOpParser{}, nil
	case config.ThinkingParserInBand:
		slog.Info("configured in-band thinking parser", "start_delimiter", cfg.StartDelimiter, "end_delimiter", cfg.EndDelimiter)
		return NewInBandParser(cfg.StartDelimiter, cfg.EndDelimiter), nil
	case config.ThinkingParserOutOfBand:
		slog.Info("configured out-of-band thinking parser", "field_path", cfg.FieldPath)
		return NewOutOfBandParser(cfg.FieldPath), nil
	default:
		return nil, fmt.Errorf("unknown thinking parser type: %s", cfg.Type)
	}
}
