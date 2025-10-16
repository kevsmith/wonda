package mcp

import (
	"context"
	"fmt"
)

// Server represents an MCP server that provides tools and resources.
// In our in-process implementation, this is a Go struct rather than a remote process,
// but it follows MCP semantics and structure.
type Server struct {
	// Name identifies this MCP server
	Name string

	// Version of the MCP server implementation
	Version string

	// Tools available from this server
	Tools map[string]*Tool

	// Resources provided by this server
	Resources map[string]*Resource
}

// NewServer creates a new MCP server.
func NewServer(name, version string) *Server {
	return &Server{
		Name:      name,
		Version:   version,
		Tools:     make(map[string]*Tool),
		Resources: make(map[string]*Resource),
	}
}

// RegisterTool adds a tool to this server.
func (s *Server) RegisterTool(tool *Tool) {
	s.Tools[tool.Name] = tool
}

// RegisterResource adds a resource to this server.
func (s *Server) RegisterResource(resource *Resource) {
	s.Resources[resource.URI] = resource
}

// GetTool retrieves a tool by name.
func (s *Server) GetTool(name string) (*Tool, error) {
	tool, ok := s.Tools[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return tool, nil
}

// ExecuteTool executes a tool with the given arguments.
func (s *Server) ExecuteTool(ctx context.Context, toolCall *ToolCall) *ToolResult {
	tool, err := s.GetTool(toolCall.Name)
	if err != nil {
		return &ToolResult{
			ToolCallID: toolCall.ID,
			Content:    err.Error(),
			IsError:    true,
			EndsTurn:   false,
		}
	}

	result, err := tool.Handler(ctx, toolCall.Arguments)
	if err != nil {
		return &ToolResult{
			ToolCallID: toolCall.ID,
			Content:    err.Error(),
			IsError:    true,
			EndsTurn:   tool.EndsTurn,
		}
	}

	return &ToolResult{
		ToolCallID: toolCall.ID,
		Content:    result,
		IsError:    false,
		EndsTurn:   tool.EndsTurn,
	}
}

// GetToolDefinitions returns tool definitions in the format expected by LLM APIs.
// This converts our MCP Tool structs into the JSON format that OpenAI/Anthropic expect.
func (s *Server) GetToolDefinitions() []map[string]interface{} {
	definitions := make([]map[string]interface{}, 0, len(s.Tools))

	for _, tool := range s.Tools {
		definitions = append(definitions, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.InputSchema,
			},
		})
	}

	return definitions
}
