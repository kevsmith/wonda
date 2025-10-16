package simulation

import (
	"context"
	"fmt"

	"github.com/poiesic/wonda/internal/mcp"
	"github.com/poiesic/wonda/internal/runtime"
)

// SpeakResult contains confirmation of a speak action.
type SpeakResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NewSpeakTool creates the speak() MCP tool.
// This tool allows agents to broadcast a message to nearby agents.
func NewSpeakTool(world *WorldState) *mcp.Tool {
	return &mcp.Tool{
		Name:        "speak",
		Description: "Say something out loud to nearby agents. Your message will be heard by all agents at your current location.",
		EndsTurn:    true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type":        "string",
					"description": "The exact words you're saying out loud. ONLY include spoken dialogue - no narration of actions, no stage directions, no meta-commentary about what you'll do or say. GOOD EXAMPLES: \"How about we grab dinner at that Italian place?\" or \"I don't know, seems pretty far from here.\" BAD EXAMPLES: \"I'll order a drink\" (narration), \"I'm going to tell him...\" (meta-narration), \"So here's my vote\" (breaking character)",
				},
			},
			"required": []string{"message"},
		},
		Handler: func(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
			// Get agent name from context
			agentName, ok := ctx.Value(runtime.AgentNameKey).(string)
			if !ok || agentName == "" {
				return nil, fmt.Errorf("agent_name not found in context")
			}

			// Extract message from arguments
			message, ok := arguments["message"].(string)
			if !ok || message == "" {
				return nil, fmt.Errorf("message parameter is required and must be a string")
			}

			// Add message to world conversation history
			world.AddMessage(agentName, message, "", MessageTypeDialogue)

			return &SpeakResult{
				Success: true,
				Message: fmt.Sprintf("You said: %s", message),
			}, nil
		},
	}
}
