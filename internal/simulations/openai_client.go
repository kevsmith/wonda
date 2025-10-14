package simulations

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"

	"github.com/poiesic/wonda/internal/config"
)

// OpenAIClient implements the Client interface for OpenAI-compatible APIs.
type OpenAIClient struct {
	client   *openai.Client
	model    *config.Model
	parser   ResponseParser
	modelID  string
}

// newOpenAIClient creates a new OpenAI-compatible client.
func newOpenAIClient(provider *config.Provider, model *config.Model, parser ResponseParser) (*OpenAIClient, error) {
	// Get API key
	apiKey := ""
	if provider.APIKey != nil {
		apiKey = *provider.APIKey
	}

	// Create OpenAI client configuration
	clientConfig := openai.DefaultConfig(apiKey)
	clientConfig.BaseURL = provider.BaseURL

	client := openai.NewClientWithConfig(clientConfig)

	return &OpenAIClient{
		client:   client,
		model:    model,
		parser:   parser,
		modelID:  model.Name,
	}, nil
}

// Chat sends a chat completion request to an OpenAI-compatible API.
func (c *OpenAIClient) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	// Convert messages to OpenAI format
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Use model from request if specified, otherwise use client's default
	modelID := req.Model
	if modelID == "" {
		modelID = c.modelID
	}

	// Create chat completion request
	chatReq := openai.ChatCompletionRequest{
		Model:    modelID,
		Messages: messages,
	}

	// Add tools if provided
	if len(req.Tools) > 0 {
		tools := make([]openai.Tool, len(req.Tools))
		for i, toolDef := range req.Tools {
			// Extract function definition
			if fn, ok := toolDef["function"].(map[string]interface{}); ok {
				tools[i] = openai.Tool{
					Type: openai.ToolTypeFunction,
					Function: &openai.FunctionDefinition{
						Name:        fn["name"].(string),
						Description: fn["description"].(string),
						Parameters:  fn["parameters"],
					},
				}
			}
		}
		chatReq.Tools = tools
	}

	// Send request
	resp, err := c.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("openai api error: %w", err)
	}

	// Check for empty response
	if len(resp.Choices) == 0 {
		return ChatResponse{}, fmt.Errorf("no response choices returned")
	}

	message := resp.Choices[0].Message

	// Extract message content
	content := message.Content

	// Extract tool calls if present
	var toolCalls []ToolCall
	if len(message.ToolCalls) > 0 {
		toolCalls = make([]ToolCall, len(message.ToolCalls))
		for i, tc := range message.ToolCalls {
			// Parse arguments JSON
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				// If parsing fails, use empty args
				args = make(map[string]interface{})
			}

			toolCalls[i] = ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: args,
			}
		}
	}

	// Extract thinking based on parser type
	var thinking string
	if outOfBandParser, ok := c.parser.(*OutOfBandParser); ok {
		// For out-of-band parsers (like o1 models), extract thinking from response JSON
		// We need to access the raw JSON to get the reasoning field
		// For now, we'll marshal the response back to JSON and extract
		if jsonData, err := json.Marshal(resp); err == nil {
			thinking = extractJSONField(jsonData, outOfBandParser.FieldPath())
		}
	} else {
		// For in-band parsers, parse the content text
		content, thinking = c.parser.Parse(content)
	}

	return ChatResponse{
		Message:   content,
		Thinking:  thinking,
		ToolCalls: toolCalls,
	}, nil
}
