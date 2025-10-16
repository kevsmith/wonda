package simulation

import "github.com/poiesic/wonda/internal/mcp"

// NewSimulationServer creates an MCP server for simulation tools.
// This server provides tools for agents to perceive and interact with the simulation world.
func NewSimulationServer(world *WorldState) *mcp.Server {
	server := mcp.NewServer("simulation", "1.0.0")

	// Register world state as a resource
	server.RegisterResource(&mcp.Resource{
		URI:         "world://state",
		Name:        "World State",
		Description: "The current state of the simulation world",
		MimeType:    "application/json",
		Read: func() (interface{}, error) {
			return world, nil
		},
	})

	// Register perception and action tools
	server.RegisterTool(NewPerceiveTool(world))
	server.RegisterTool(NewSpeakTool(world))
	server.RegisterTool(NewNarrateActionTool(world))
	server.RegisterTool(NewInternalMonologueTool(world))

	// Register goal interaction tools
	server.RegisterTool(NewListGoalsTool(world))
	server.RegisterTool(NewViewGoalTool(world))
	server.RegisterTool(NewProposeSolutionTool(world))
	server.RegisterTool(NewVoteOnProposalTool(world))
	server.RegisterTool(NewWithdrawProposalTool(world))

	return server
}
