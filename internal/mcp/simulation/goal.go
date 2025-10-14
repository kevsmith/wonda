package simulation

import "fmt"

// GoalStatus represents the current state of a goal.
type GoalStatus string

const (
	GoalPending   GoalStatus = "pending"
	GoalCompleted GoalStatus = "completed"
	GoalFailed    GoalStatus = "failed"
)

// ProposalStatus represents the state of a proposal.
type ProposalStatus string

const (
	ProposalPending   ProposalStatus = "pending"
	ProposalAccepted  ProposalStatus = "accepted"
	ProposalRejected  ProposalStatus = "rejected"
	ProposalWithdrawn ProposalStatus = "withdrawn"
)

// InteractiveGoal represents a goal that agents can interact with through MCP tools.
type InteractiveGoal struct {
	Name        string
	Description string
	Type        string // "consensus", "individual", "majority"
	Status      GoalStatus
	Priority    int

	// For consensus goals
	Proposals   map[string]*Proposal
	CompletedAt int // Turn number when completed
}

// Proposal represents a proposed solution to a goal.
type Proposal struct {
	ID          string
	Description string
	ProposedBy  string
	ProposedAt  int
	Status      ProposalStatus
	Votes       map[string]*Vote
	ResolvedAt  int // Turn when status changed from pending
}

// Vote represents an agent's vote on a proposal.
type Vote struct {
	AgentName string
	Choice    string // "yes", "no"
	VotedAt   int
}

// NewInteractiveGoal creates a new interactive goal.
func NewInteractiveGoal(name, description, goalType string, priority int) *InteractiveGoal {
	return &InteractiveGoal{
		Name:        name,
		Description: description,
		Type:        goalType,
		Priority:    priority,
		Status:      GoalPending,
		Proposals:   make(map[string]*Proposal),
	}
}

// AddProposal adds a new proposal to this goal.
func (g *InteractiveGoal) AddProposal(agentName, description string, turn int) string {
	proposalID := fmt.Sprintf("proposal_%d", len(g.Proposals)+1)
	g.Proposals[proposalID] = &Proposal{
		ID:          proposalID,
		Description: description,
		ProposedBy:  agentName,
		ProposedAt:  turn,
		Status:      ProposalPending,
		Votes:       make(map[string]*Vote),
	}
	return proposalID
}

// Vote records a vote on a proposal.
func (g *InteractiveGoal) Vote(proposalID, agentName, choice string, turn int) error {
	proposal, ok := g.Proposals[proposalID]
	if !ok {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.Status != ProposalPending {
		return fmt.Errorf("cannot vote on %s proposal", proposal.Status)
	}

	proposal.Votes[agentName] = &Vote{
		AgentName: agentName,
		Choice:    choice,
		VotedAt:   turn,
	}

	return nil
}

// EvaluateProposal checks if a proposal should be accepted or rejected.
// For consensus goals, all agents must vote yes for acceptance.
func (p *Proposal) EvaluateStatus(totalAgents int, turn int) {
	if p.Status != ProposalPending {
		return
	}

	// Check if all agents have voted
	if len(p.Votes) < totalAgents {
		return
	}

	// Count votes
	yesVotes := 0
	noVotes := 0
	for _, vote := range p.Votes {
		switch vote.Choice {
		case "yes":
			yesVotes++
		case "no":
			noVotes++
		}
	}

	// Determine outcome (unanimous yes required)
	if yesVotes == totalAgents {
		p.Status = ProposalAccepted
		p.ResolvedAt = turn
	} else if noVotes > 0 {
		p.Status = ProposalRejected
		p.ResolvedAt = turn
	}
}

// WithdrawProposal marks a proposal as withdrawn.
func (g *InteractiveGoal) WithdrawProposal(proposalID, agentName string, turn int) error {
	proposal, ok := g.Proposals[proposalID]
	if !ok {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.ProposedBy != agentName {
		return fmt.Errorf("only the proposer can withdraw a proposal")
	}

	if proposal.Status != ProposalPending {
		return fmt.Errorf("can only withdraw pending proposals")
	}

	proposal.Status = ProposalWithdrawn
	proposal.ResolvedAt = turn
	return nil
}

// CheckConsensus checks if any proposal has been accepted.
// If so, marks the goal as completed.
func (g *InteractiveGoal) CheckConsensus(turn int) bool {
	for _, proposal := range g.Proposals {
		if proposal.Status == ProposalAccepted {
			g.Status = GoalCompleted
			g.CompletedAt = turn
			return true
		}
	}
	return false
}
