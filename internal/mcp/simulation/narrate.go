package simulation

import (
	"context"
	"fmt"

	"github.com/poiesic/wonda/internal/mcp"
	"github.com/poiesic/wonda/internal/runtime"
)

// NarrateActionResult contains confirmation of a narration action.
type NarrateActionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NewNarrateActionTool creates the narrate_action() MCP tool.
// This tool allows agents to describe physical actions they're performing.
func NewNarrateActionTool(world *WorldState) *mcp.Tool {
	return &mcp.Tool{
		Name:        "narrate_action",
		Description: "Describe a physical action you're performing. Use this when you need to express what you're doing physically (ordering drinks, looking around, moving, gesturing, etc.) without speaking out loud.",
		EndsTurn:    true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"action": map[string]interface{}{
					"type":        "string",
					"description": "Description of the action you're performing. Write in third person present tense. EXAMPLES: \"orders an Old Fashioned from the bartender\" or \"glances around the bar nervously\" or \"leans back in his chair\" or \"checks her watch\"",
				},
			},
			"required": []string{"action"},
		},
		Handler: func(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
			// Get agent name from context
			agentName, ok := ctx.Value(runtime.AgentNameKey).(string)
			if !ok || agentName == "" {
				return nil, fmt.Errorf("agent_name not found in context")
			}

			// Extract action from arguments
			action, ok := arguments["action"].(string)
			if !ok || action == "" {
				return nil, fmt.Errorf("action parameter is required and must be a string")
			}

			// Add action to pending dialogue (will be captured by simulation)
			world.AddPendingDialogue(agentName, action, MessageTypeAction)

			return &NarrateActionResult{
				Success: true,
				Message: "Action recorded",
			}, nil
		},
	}
}

// InternalMonologueResult contains confirmation of an internal monologue.
type InternalMonologueResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NewInternalMonologueTool creates the internal_monologue() MCP tool.
// This tool allows agents to express their private thoughts.
func NewInternalMonologueTool(world *WorldState) *mcp.Tool {
	return &mcp.Tool{
		Name:        "internal_monologue",
		Description: "Express your private thoughts that other agents cannot hear. Use this for internal reactions, strategic thinking, feelings, or observations you're keeping to yourself.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"thought": map[string]interface{}{
					"type":        "string",
					"description": "Your private thought. Write in first person. EXAMPLES: \"She's definitely hiding something\" or \"I can't let her see how much this bothers me\" or \"That was smooth - maybe too smooth\" or \"Time to change the subject before this gets awkward\"",
				},
			},
			"required": []string{"thought"},
		},
		Handler: func(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
			// Get agent name from context
			agentName, ok := ctx.Value(runtime.AgentNameKey).(string)
			if !ok || agentName == "" {
				return nil, fmt.Errorf("agent_name not found in context")
			}

			// Extract thought from arguments
			thought, ok := arguments["thought"].(string)
			if !ok || thought == "" {
				return nil, fmt.Errorf("thought parameter is required and must be a string")
			}

			// Add thought to pending dialogue (will be captured by simulation)
			world.AddPendingDialogue(agentName, thought, MessageTypeMonologue)

			return &InternalMonologueResult{
				Success: true,
				Message: "Thought recorded",
			}, nil
		},
	}
}
