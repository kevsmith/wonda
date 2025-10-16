package simulations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"

	openai "github.com/sashabaranov/go-openai"

	"github.com/poiesic/wonda/internal/config"
)

// OpenAIClient implements the Client interface for OpenAI-compatible APIs.
type OpenAIClient struct {
	client  *openai.Client
	model   *config.Model
	parser  ResponseParser
	modelID string
	baseURL string
	apiKey  string
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
		client:  client,
		model:   model,
		parser:  parser,
		modelID: model.Name,
		baseURL: provider.BaseURL,
		apiKey:  apiKey,
	}, nil
}

// Chat sends a chat completion request to an OpenAI-compatible API.
func (c *OpenAIClient) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	// If we have an out-of-band parser (need to extract custom fields like reasoning),
	// use raw HTTP request to get full JSON response
	if _, needsRawJSON := c.parser.(*OutOfBandParser); needsRawJSON {
		return c.chatRaw(ctx, req)
	}

	// Otherwise use the go-openai library (faster, more reliable for standard fields)
	return c.chatWithLibrary(ctx, req)
}

// chatWithLibrary uses the go-openai library for standard requests.
func (c *OpenAIClient) chatWithLibrary(ctx context.Context, req ChatRequest) (ChatResponse, error) {
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

	// Extract message content and clean up model artifacts
	content := cleanModelArtifacts(message.Content)

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
			fieldPath := outOfBandParser.FieldPath()
			thinking = extractJSONField(jsonData, fieldPath)

			// Log thinking extraction results
			if thinking == "" {
				slog.Info("out-of-band thinking parser found no content", "field_path", fieldPath, "hint", "check if model supports this field or if parser is misconfigured")
				// Write full response to file for inspection
				if err := os.WriteFile("/tmp/wonda-llm-response.json", jsonData, 0644); err == nil {
					slog.Debug("full response written to file for inspection", "path", "/tmp/wonda-llm-response.json")
				}
				// Log first 1000 chars of response for quick debugging
				preview := string(jsonData)
				if len(jsonData) > 1000 {
					preview = string(jsonData[:1000]) + "..."
				}
				slog.Debug("response preview", "data", preview)
			} else {
				slog.Debug("successfully extracted thinking", "length", len(thinking), "content", thinking)
			}
		}
	} else {
		// For in-band parsers, parse the content text
		content, thinking = c.parser.Parse(content)
		if thinking == "" && c.parser != nil {
			if _, isNoOp := c.parser.(*NoOpParser); !isNoOp {
				slog.Info("in-band thinking parser found no content", "hint", "check if model uses correct delimiters or if parser is misconfigured")
			}
		} else if thinking != "" {
			slog.Debug("successfully extracted thinking from in-band parser", "length", len(thinking), "content", thinking)
		}
	}

	return ChatResponse{
		Message:   content,
		Thinking:  thinking,
		ToolCalls: toolCalls,
	}, nil
}

// chatRaw makes a raw HTTP request to preserve all custom fields in the response.
func (c *OpenAIClient) chatRaw(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	// Build request body
	modelID := req.Model
	if modelID == "" {
		modelID = c.modelID
	}

	// Convert messages to proper format
	messages := make([]map[string]interface{}, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	reqBody := map[string]interface{}{
		"model":    modelID,
		"messages": messages,
	}

	// Add tools if provided
	if len(req.Tools) > 0 {
		reqBody["tools"] = req.Tools
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	// baseURL already includes /v1, just append the endpoint
	url := strings.TrimRight(c.baseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Send request
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("http request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return ChatResponse{}, fmt.Errorf("api error (status %d): %s", httpResp.StatusCode, string(respBody))
	}

	// Parse response to extract standard fields
	var rawResp map[string]interface{}
	if err := json.Unmarshal(respBody, &rawResp); err != nil {
		return ChatResponse{}, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract message content
	choices, ok := rawResp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return ChatResponse{}, fmt.Errorf("no choices in response")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return ChatResponse{}, fmt.Errorf("invalid choice format")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return ChatResponse{}, fmt.Errorf("no message in choice")
	}

	content, _ := message["content"].(string)
	content = cleanModelArtifacts(content)

	// Extract tool calls
	var toolCalls []ToolCall
	if toolCallsRaw, ok := message["tool_calls"].([]interface{}); ok {
		for _, tcRaw := range toolCallsRaw {
			tc, ok := tcRaw.(map[string]interface{})
			if !ok {
				continue
			}

			function, ok := tc["function"].(map[string]interface{})
			if !ok {
				continue
			}

			// Parse arguments
			var args map[string]interface{}
			if argsStr, ok := function["arguments"].(string); ok {
				json.Unmarshal([]byte(argsStr), &args)
			}

			toolCalls = append(toolCalls, ToolCall{
				ID:        tc["id"].(string),
				Name:      function["name"].(string),
				Arguments: args,
			})
		}
	}

	// Extract thinking using JSONPath on the raw JSON
	var thinking string
	if outOfBandParser, ok := c.parser.(*OutOfBandParser); ok {
		fieldPath := outOfBandParser.FieldPath()
		thinking = extractJSONField(respBody, fieldPath)

		// Show thinking activity
		if thinking != "" {
			slog.Debug("thinking extracted", "length", len(thinking))
		}
	}

	return ChatResponse{
		Message:   content,
		Thinking:  thinking,
		ToolCalls: toolCalls,
	}, nil
}

// cleanModelArtifacts removes internal model tokens and artifacts from output text.
// This cleans up responses from models that leak their function-calling format.
func cleanModelArtifacts(text string) string {
	// Common patterns to remove:
	// - <|start|>...to=assistant
	// - <|call|>, <|message|>, <|channel|>, <|constrain|>
	// - Tool execution traces like "Tool 'name' returned:"

	// Remove special tokens
	patterns := []string{
		`<\|start\|>[^<]*to=[^\s<]+`,         // <|start|>...to=assistant
		`<\|call\|>`,                         // <|call|>
		`<\|message\|>`,                      // <|message|>
		`<\|channel\|>[^<]*`,                 // <|channel|>...
		`<\|constrain\|>[^\s<]+`,             // <|constrain|>json
		`<\|end\|>`,                          // <|end|>
		`Tool '[^']+' returned:\s*\{[^}]*\}`, // Tool execution traces
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		text = re.ReplaceAllString(text, "")
	}

	return strings.TrimSpace(text)
}
