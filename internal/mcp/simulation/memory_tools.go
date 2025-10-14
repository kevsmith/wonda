package simulation

import (
	"context"
	"fmt"

	"github.com/poiesic/wonda/internal/mcp"
	"github.com/poiesic/wonda/internal/memory"
	"github.com/poiesic/wonda/internal/runtime"
)

// NewQuerySelfTool creates the query_self MCP tool.
// Returns core identity information about the agent.
func NewQuerySelfTool(store *memory.Store) *mcp.Tool {
	return &mcp.Tool{
		Name:        "query_self",
		Description: "Retrieve your core identity - who you are, your personality, background",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []string{},
		},
		Handler: func(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
			agentName, ok := ctx.Value(runtime.AgentNameKey).(string)
			if !ok || agentName == "" {
				return nil, fmt.Errorf("agent_name not found in context")
			}

			// Use canonical query
			results, err := store.SearchByCanonicalQuery(
				ctx,
				"who am I?",
				memory.Filter{
					Agent:    agentName,
					Type:     "character",
					Category: "identity",
				},
				5,
			)
			if err != nil {
				return nil, err
			}

			// Format results
			memories := make([]map[string]interface{}, len(results))
			for i, mem := range results {
				memories[i] = map[string]interface{}{
					"content":   mem.Content,
					"relevance": mem.Score,
				}
			}

			return map[string]interface{}{
				"memories": memories,
			}, nil
		},
	}
}

// NewQueryBackgroundTool creates the query_background MCP tool.
func NewQueryBackgroundTool(store *memory.Store) *mcp.Tool {
	return &mcp.Tool{
		Name:        "query_background",
		Description: "Retrieve your personal history and background",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []string{},
		},
		Handler: func(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
			agentName, ok := ctx.Value(runtime.AgentNameKey).(string)
			if !ok || agentName == "" {
				return nil, fmt.Errorf("agent_name not found in context")
			}

			results, err := store.SearchByCanonicalQuery(
				ctx,
				"what is my background?",
				memory.Filter{
					Agent:    agentName,
					Type:     "character",
					Category: "background",
				},
				5,
			)
			if err != nil {
				return nil, err
			}

			memories := make([]map[string]interface{}, len(results))
			for i, mem := range results {
				memories[i] = map[string]interface{}{
					"content":   mem.Content,
					"relevance": mem.Score,
				}
			}

			return map[string]interface{}{
				"memories": memories,
			}, nil
		},
	}
}

// NewQueryCommunicationStyleTool creates the query_communication_style MCP tool.
func NewQueryCommunicationStyleTool(store *memory.Store) *mcp.Tool {
	return &mcp.Tool{
		Name:        "query_communication_style",
		Description: "Learn how you communicate and interact with others",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []string{},
		},
		Handler: func(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
			agentName, ok := ctx.Value(runtime.AgentNameKey).(string)
			if !ok || agentName == "" {
				return nil, fmt.Errorf("agent_name not found in context")
			}

			results, err := store.SearchByCanonicalQuery(
				ctx,
				"how do I communicate?",
				memory.Filter{
					Agent:    agentName,
					Type:     "character",
					Category: "communication",
				},
				3,
			)
			if err != nil {
				return nil, err
			}

			memories := make([]map[string]interface{}, len(results))
			for i, mem := range results {
				memories[i] = map[string]interface{}{
					"content":   mem.Content,
					"relevance": mem.Score,
				}
			}

			return map[string]interface{}{
				"memories": memories,
			}, nil
		},
	}
}

// NewQuerySceneTool creates the query_scene MCP tool.
func NewQuerySceneTool(store *memory.Store) *mcp.Tool {
	return &mcp.Tool{
		Name:        "query_scene",
		Description: "Understand where you are and the current atmosphere",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []string{},
		},
		Handler: func(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
			results, err := store.SearchByCanonicalQuery(
				ctx,
				"where am I?",
				memory.Filter{
					Type: "scene",
				},
				5,
			)
			if err != nil {
				return nil, err
			}

			memories := make([]map[string]interface{}, len(results))
			for i, mem := range results {
				memories[i] = map[string]interface{}{
					"content":   mem.Content,
					"relevance": mem.Score,
				}
			}

			return map[string]interface{}{
				"memories": memories,
			}, nil
		},
	}
}

// NewQueryCharacterTool creates the query_character MCP tool.
func NewQueryCharacterTool(store *memory.Store) *mcp.Tool {
	return &mcp.Tool{
		Name:        "query_character",
		Description: "Learn about another agent in the simulation",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the character to query",
				},
			},
			"required": []string{"name"},
		},
		Handler: func(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
			agentName, ok := ctx.Value(runtime.AgentNameKey).(string)
			if !ok || agentName == "" {
				return nil, fmt.Errorf("agent_name not found in context")
			}

			targetName, ok := arguments["name"].(string)
			if !ok {
				return nil, fmt.Errorf("name parameter is required")
			}

			// Fixed query pattern, parameterized by name
			query := fmt.Sprintf("who is %s?", targetName)

			results, err := store.SearchByCanonicalQuery(
				ctx,
				query,
				memory.Filter{
					Agent: agentName,
					Type:  "character_knowledge",
					About: targetName,
				},
				3,
			)
			if err != nil {
				return nil, err
			}

			memories := make([]map[string]interface{}, len(results))
			for i, mem := range results {
				memories[i] = map[string]interface{}{
					"content":   mem.Content,
					"relevance": mem.Score,
				}
			}

			return map[string]interface{}{
				"character": targetName,
				"memories":  memories,
			}, nil
		},
	}
}

// NewQueryMemoryTool creates the query_memory MCP tool for flexible episodic search.
func NewQueryMemoryTool(store *memory.Store) *mcp.Tool {
	return &mcp.Tool{
		Name:        "query_memory",
		Description: "Search your memories of what has happened during the simulation",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "What you want to remember (e.g., 'what did [other agent] say about the goal?')",
				},
			},
			"required": []string{"query"},
		},
		Handler: func(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
			query, ok := arguments["query"].(string)
			if !ok || query == "" {
				return nil, fmt.Errorf("query parameter is required")
			}

			// User-provided query for flexible semantic search
			embedding, err := store.Embed(ctx, query)
			if err != nil {
				return nil, fmt.Errorf("failed to embed query: %w", err)
			}

			results := store.Search(
				ctx,
				embedding,
				memory.Filter{
					Type: "episodic",
				},
				5,
			)

			memories := make([]map[string]interface{}, len(results))
			for i, mem := range results {
				memories[i] = map[string]interface{}{
					"content":   mem.Content,
					"relevance": mem.Score,
					"turn":      mem.Metadata["turn"],
				}
			}

			return map[string]interface{}{
				"query":   query,
				"memories": memories,
			}, nil
		},
	}
}
