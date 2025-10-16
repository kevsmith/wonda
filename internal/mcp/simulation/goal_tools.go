package simulation

import (
	"context"
	"fmt"

	"github.com/poiesic/wonda/internal/mcp"
	"github.com/poiesic/wonda/internal/runtime"
)

// NewListGoalsTool creates the list_goals MCP tool.
// Allows agents to discover what goals exist.
func NewListGoalsTool(world *WorldState) *mcp.Tool {
	return &mcp.Tool{
		Name:        "list_goals",
		Description: "List all available goals you can work toward",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []string{},
		},
		Handler: func(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
			goals := make([]map[string]interface{}, 0, len(world.Goals))
			for _, goal := range world.Goals {
				goals = append(goals, map[string]interface{}{
					"name":        goal.Name,
					"description": goal.Description,
					"status":      string(goal.Status),
					"priority":    goal.Priority,
				})
			}
			return map[string]interface{}{
				"goals":        goals,
				"current_turn": world.CurrentTurn,
			}, nil
		},
	}
}

// NewViewGoalTool creates the view_goal MCP tool.
// Allows agents to check the current status of goals, proposals, and votes.
func NewViewGoalTool(world *WorldState) *mcp.Tool {
	return &mcp.Tool{
		Name:        "view_goal",
		Description: "Check the current status of a goal, including pending proposals you can vote on and history of past proposals",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"goal_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the goal to check",
				},
			},
			"required": []string{"goal_name"},
		},
		Handler: func(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
			goalName, ok := arguments["goal_name"].(string)
			if !ok {
				return nil, fmt.Errorf("goal_name is required")
			}

			goal, ok := world.Goals[goalName]
			if !ok {
				return nil, fmt.Errorf("goal not found: %s", goalName)
			}

			// Separate proposals by status
			pending := []map[string]interface{}{}
			accepted := []map[string]interface{}{}
			rejected := []map[string]interface{}{}
			withdrawn := []map[string]interface{}{}

			for _, proposal := range goal.Proposals {
				votes := make(map[string]string)
				for agentName, vote := range proposal.Votes {
					votes[agentName] = vote.Choice
				}

				formatted := map[string]interface{}{
					"id":          proposal.ID,
					"description": proposal.Description,
					"proposed_by": proposal.ProposedBy,
					"proposed_at": proposal.ProposedAt,
					"votes":       votes,
				}

				switch proposal.Status {
				case ProposalPending:
					pending = append(pending, formatted)
				case ProposalAccepted:
					formatted["resolved_at"] = proposal.ResolvedAt
					accepted = append(accepted, formatted)
				case ProposalRejected:
					formatted["resolved_at"] = proposal.ResolvedAt
					rejected = append(rejected, formatted)
				case ProposalWithdrawn:
					formatted["resolved_at"] = proposal.ResolvedAt
					withdrawn = append(withdrawn, formatted)
				}
			}

			return map[string]interface{}{
				"name":                goal.Name,
				"description":         goal.Description,
				"status":              string(goal.Status),
				"priority":            goal.Priority,
				"current_turn":        world.CurrentTurn,
				"pending_proposals":   pending,
				"accepted_proposals":  accepted,
				"rejected_proposals":  rejected,
				"withdrawn_proposals": withdrawn,
			}, nil
		},
	}
}

// NewProposeSolutionTool creates the propose_solution MCP tool.
// Allows agents to propose solutions to goals.
func NewProposeSolutionTool(world *WorldState) *mcp.Tool {
	return &mcp.Tool{
		Name:        "propose_solution",
		Description: "Propose ONE specific solution for a goal with an in-character pitch. Each proposal must be a single, concrete choice - not a list of options.",
		EndsTurn:    true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"goal_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the goal",
				},
				"solution": map[string]interface{}{
					"type":        "string",
					"description": "Your proposed solution - must be ONE specific choice (e.g., 'Bella's Italian Restaurant'), NOT multiple options or alternatives",
				},
				"comment": map[string]interface{}{
					"type":        "string",
					"description": "What you SAY out loud as you propose this - an in-character pitch for your idea. Sell it, explain what makes it good, be persuasive and authentic. EXAMPLES: \"How about we hit up The Skyline Lounge? Best cocktails in the city and the view is killer.\" or \"I'm thinking Bella's - intimate, great food, and the owner owes me a favor.\"",
				},
			},
			"required": []string{"goal_name", "solution", "comment"},
		},
		Handler: func(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
			agentName, ok := ctx.Value(runtime.AgentNameKey).(string)
			if !ok || agentName == "" {
				return nil, fmt.Errorf("agent_name not found in context")
			}

			goalName, ok := arguments["goal_name"].(string)
			if !ok {
				return nil, fmt.Errorf("goal_name is required")
			}

			solution, ok := arguments["solution"].(string)
			if !ok || solution == "" {
				return nil, fmt.Errorf("solution is required and must be a string")
			}

			comment, ok := arguments["comment"].(string)
			if !ok || comment == "" {
				return nil, fmt.Errorf("comment is required - you must say something as you propose")
			}

			goal, ok := world.Goals[goalName]
			if !ok {
				return nil, fmt.Errorf("goal not found: %s", goalName)
			}

			if goal.Status != GoalPending {
				return nil, fmt.Errorf("cannot propose solutions to %s goals", goal.Status)
			}

			// Check if agent already has a proposal for this goal this turn
			for _, proposal := range goal.Proposals {
				if proposal.ProposedBy == agentName && proposal.ProposedAt == world.CurrentTurn {
					return nil, fmt.Errorf("you already proposed a solution for this goal this turn")
				}
			}

			// Add comment to pending dialogue (will be captured by simulation)
			world.AddPendingDialogue(agentName, comment, MessageTypeDialogue)

			proposalID := goal.AddProposal(agentName, solution, world.CurrentTurn)

			// Auto-vote yes on own proposal (agents always support their own proposals)
			if err := goal.Vote(proposalID, agentName, "yes", world.CurrentTurn); err != nil {
				return nil, fmt.Errorf("failed to auto-vote on proposal: %w", err)
			}

			return map[string]interface{}{
				"success":     true,
				"proposal_id": proposalID,
				"message":     fmt.Sprintf("Proposed: %s (auto-voted yes)", solution),
			}, nil
		},
	}
}

// NewVoteOnProposalTool creates the vote_on_proposal MCP tool.
// Allows agents to vote yes/no on proposals.
func NewVoteOnProposalTool(world *WorldState) *mcp.Tool {
	return &mcp.Tool{
		Name:        "vote_on_proposal",
		Description: "Cast your vote on a proposal with an in-character statement. When all agents vote yes, the proposal is accepted and the goal is completed.",
		EndsTurn:    true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"goal_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the goal",
				},
				"proposal_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the proposal to vote on (from view_goal)",
				},
				"vote": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"yes", "no"},
					"description": "Your vote (yes or no)",
				},
				"comment": map[string]interface{}{
					"type":        "string",
					"description": "What you SAY out loud as you cast this vote - an in-character statement announcing your decision. Reference the specific proposal and express your authentic reaction to it. EXAMPLES: \"I'm on board with the rooftop idea - that place has the best view in town.\" or \"Not feeling the Italian restaurant, it's too far from here.\"",
				},
			},
			"required": []string{"goal_name", "proposal_id", "vote", "comment"},
		},
		Handler: func(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
			agentName, ok := ctx.Value(runtime.AgentNameKey).(string)
			if !ok || agentName == "" {
				return nil, fmt.Errorf("agent_name not found in context")
			}

			goalName, ok := arguments["goal_name"].(string)
			if !ok {
				return nil, fmt.Errorf("goal_name is required")
			}

			proposalID, ok := arguments["proposal_id"].(string)
			if !ok {
				return nil, fmt.Errorf("proposal_id is required")
			}

			vote, ok := arguments["vote"].(string)
			if !ok {
				return nil, fmt.Errorf("vote is required")
			}

			if vote != "yes" && vote != "no" {
				return nil, fmt.Errorf("vote must be 'yes' or 'no'")
			}

			comment, ok := arguments["comment"].(string)
			if !ok || comment == "" {
				return nil, fmt.Errorf("comment is required - you must say something as you vote")
			}

			goal, ok := world.Goals[goalName]
			if !ok {
				return nil, fmt.Errorf("goal not found: %s", goalName)
			}

			if goal.Status != GoalPending {
				return nil, fmt.Errorf("cannot vote on %s goals", goal.Status)
			}

			proposal, ok := goal.Proposals[proposalID]
			if !ok {
				return nil, fmt.Errorf("proposal not found: %s", proposalID)
			}

			// Check if agent already voted on this proposal
			if _, hasVoted := proposal.Votes[agentName]; hasVoted {
				return nil, fmt.Errorf("you already voted on this proposal")
			}

			// Add comment to pending dialogue (will be captured by simulation)
			world.AddPendingDialogue(agentName, comment, MessageTypeDialogue)

			// Record vote
			if err := goal.Vote(proposalID, agentName, vote, world.CurrentTurn); err != nil {
				return nil, err
			}

			// Evaluate proposal status
			proposal.EvaluateStatus(len(world.Agents), world.CurrentTurn)

			result := map[string]interface{}{
				"success": true,
				"message": fmt.Sprintf("Voted %s on proposal", vote),
			}

			// Check outcome
			switch proposal.Status {
			case ProposalAccepted:
				goal.CheckConsensus(world.CurrentTurn)
				result["outcome"] = "accepted"
				result["message"] = "Proposal accepted! Goal completed."
				result["goal_completed"] = true
			case ProposalRejected:
				result["outcome"] = "rejected"
				result["message"] = "Proposal rejected. You can propose alternatives."
			}

			return result, nil
		},
	}
}

// NewWithdrawProposalTool creates the withdraw_proposal MCP tool.
// Allows agents to withdraw their own proposals.
func NewWithdrawProposalTool(world *WorldState) *mcp.Tool {
	return &mcp.Tool{
		Name:        "withdraw_proposal",
		Description: "Withdraw your own proposal if you've changed your mind or want to propose something different",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"goal_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the goal",
				},
				"proposal_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of your proposal to withdraw",
				},
			},
			"required": []string{"goal_name", "proposal_id"},
		},
		Handler: func(ctx context.Context, arguments map[string]interface{}) (interface{}, error) {
			agentName, ok := ctx.Value(runtime.AgentNameKey).(string)
			if !ok || agentName == "" {
				return nil, fmt.Errorf("agent_name not found in context")
			}

			goalName, ok := arguments["goal_name"].(string)
			if !ok {
				return nil, fmt.Errorf("goal_name is required")
			}

			proposalID, ok := arguments["proposal_id"].(string)
			if !ok {
				return nil, fmt.Errorf("proposal_id is required")
			}

			goal, ok := world.Goals[goalName]
			if !ok {
				return nil, fmt.Errorf("goal not found: %s", goalName)
			}

			if err := goal.WithdrawProposal(proposalID, agentName, world.CurrentTurn); err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"success": true,
				"message": "Proposal withdrawn",
			}, nil
		},
	}
}
