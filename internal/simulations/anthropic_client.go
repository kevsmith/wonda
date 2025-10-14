package simulations

import (
	"context"
	"encoding/json"
	"fmt"

	anthropic "github.com/liushuangls/go-anthropic/v2"

	"github.com/poiesic/wonda/internal/config"
)

// AnthropicClient implements the Client interface for Anthropic's Claude API.
type AnthropicClient struct {
	client  *anthropic.Client
	model   *config.Model
	parser  ResponseParser
	modelID string
}

// newAnthropicClient creates a new Anthropic client.
func newAnthropicClient(provider *config.Provider, model *config.Model, parser ResponseParser) (*AnthropicClient, error) {
	// Get API key
	apiKey := ""
	if provider.APIKey != nil {
		apiKey = *provider.APIKey
	}

	// Create Anthropic client
	// Note: Only override base URL if it's different from the default
	opts := []anthropic.ClientOption{
		anthropic.WithAPIVersion(anthropic.APIVersion20230601),
	}
	if provider.BaseURL != "" && provider.BaseURL != "https://api.anthropic.com" {
		opts = append(opts, anthropic.WithBaseURL(provider.BaseURL))
	}
	client := anthropic.NewClient(apiKey, opts...)

	return &AnthropicClient{
		client:  client,
		model:   model,
		parser:  parser,
		modelID: model.Name,
	}, nil
}

// Chat sends a chat completion request to Anthropic's API.
func (c *AnthropicClient) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	// Convert messages to Anthropic format
	messages := make([]anthropic.Message, 0, len(req.Messages))
	var systemPrompt string

	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			// Anthropic handles system messages separately
			if systemPrompt != "" {
				systemPrompt += "\n\n"
			}
			systemPrompt += msg.Content
		case "user":
			if msg.Content != "" {
				messages = append(messages, anthropic.NewUserTextMessage(msg.Content))
			}
		case "assistant":
			if msg.Content != "" {
				messages = append(messages, anthropic.NewAssistantTextMessage(msg.Content))
			}
		case "tool":
			// Anthropic expects tool results as user messages
			// Skip empty tool messages
			if msg.Content != "" {
				messages = append(messages, anthropic.NewUserTextMessage(msg.Content))
			}
		default:
			return ChatResponse{}, fmt.Errorf("unsupported message role: %s", msg.Role)
		}
	}

	// Use model from request if specified, otherwise use client's default
	modelID := req.Model
	if modelID == "" {
		modelID = c.modelID
	}

	// Create message request
	msgReq := anthropic.MessagesRequest{
		Model:     anthropic.Model(modelID),
		Messages:  messages,
		MaxTokens: 4096, // Default max tokens
	}

	// Add system prompt if present
	if systemPrompt != "" {
		msgReq.System = systemPrompt
	}

	// Add tools if provided
	if len(req.Tools) > 0 {
		tools := make([]anthropic.ToolDefinition, len(req.Tools))
		for i, toolDef := range req.Tools {
			// Extract function definition
			if fn, ok := toolDef["function"].(map[string]interface{}); ok {
				tools[i] = anthropic.ToolDefinition{
					Name:        fn["name"].(string),
					Description: fn["description"].(string),
					InputSchema: fn["parameters"],
				}
			}
		}
		msgReq.Tools = tools
	}

	// Send request
	resp, err := c.client.CreateMessages(ctx, msgReq)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("anthropic api error: %w", err)
	}

	// Extract message content, thinking, and tool calls
	var content string
	var thinking string
	var toolCalls []ToolCall

	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			if block.Text != nil {
				if content != "" {
					content += "\n"
				}
				content += *block.Text
			}
		case "thinking":
			if block.MessageContentThinking != nil {
				if thinking != "" {
					thinking += "\n\n"
				}
				thinking += block.MessageContentThinking.Thinking
			}
		case "tool_use":
			if block.MessageContentToolUse != nil {
				// Parse the Input JSON into a map
				var args map[string]interface{}
				if err := json.Unmarshal(block.Input, &args); err != nil {
					// If parsing fails, use empty args
					args = make(map[string]interface{})
				}
				toolCalls = append(toolCalls, ToolCall{
					ID:        block.ID,
					Name:      block.Name,
					Arguments: args,
				})
			}
		}
	}

	// If no extended thinking found, try in-band parsing
	if thinking == "" && c.parser != nil {
		content, thinking = c.parser.Parse(content)
	}

	return ChatResponse{
		Message:   content,
		Thinking:  thinking,
		ToolCalls: toolCalls,
	}, nil
}
