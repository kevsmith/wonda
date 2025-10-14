package simulation

import (
	"context"
	"fmt"

	"github.com/poiesic/wonda/internal/mcp"
	"github.com/poiesic/wonda/internal/runtime"
)

// PerceptionResult contains what an agent perceives about their environment.
type PerceptionResult struct {
	Location       string   `json:"location"`
	Atmosphere     string   `json:"atmosphere"`
	Position       string   `json:"your_position"`
	NearbyAgents   []string `json:"nearby_agents"`
	RecentMessages []string `json:"recent_messages"`
}

// NewPerceiveTool creates the perceive() MCP tool.
// This tool allows agents to observe their current surroundings.
func NewPerceiveTool(world *WorldState) *mcp.Tool {
	return &mcp.Tool{
		Name:        "perceive",
		Description: "Observe your current surroundings, including your location, nearby agents, and recent conversation",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []string{},
		},
		Handler: func(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
			// Get agent name from context
			agentName, ok := ctx.Value(runtime.AgentNameKey).(string)
			if !ok || agentName == "" {
				return nil, fmt.Errorf("agent_name not found in context")
			}

			// Get agent's position
			agent, ok := world.Agents[agentName]
			if !ok {
				return nil, fmt.Errorf("agent %s not found in world", agentName)
			}

			// Find nearby agents
			nearbyAgents := world.GetNearbyAgents(agentName)

			// Get recent conversation (last 5 messages)
			recentMessages := make([]string, 0)
			messages := world.GetRecentMessages(5)
			for _, msg := range messages {
				recentMessages = append(recentMessages, fmt.Sprintf("%s: %s", msg.AgentName, msg.Content))
			}

			return &PerceptionResult{
				Location:       world.Location,
				Atmosphere:     world.Atmosphere,
				Position:       agent.Position,
				NearbyAgents:   nearbyAgents,
				RecentMessages: recentMessages,
			}, nil
		},
	}
}
