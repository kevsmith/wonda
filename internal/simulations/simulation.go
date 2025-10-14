package simulations

import (
	"context"
	"fmt"
	"path"

	"github.com/poiesic/wonda/internal/config"
	"github.com/poiesic/wonda/internal/mcp"
	mcpsim "github.com/poiesic/wonda/internal/mcp/simulation"
	"github.com/poiesic/wonda/internal/memory"
	"github.com/poiesic/wonda/internal/runtime"
	"github.com/poiesic/wonda/internal/scenarios"
)

// Simulation represents a running instance of a scenario.
type Simulation struct {
	Scenario  *scenarios.Scenario
	Agents    map[string]*Agent
	ConfigDir string

	// Turn management
	TurnOrder []string // Agent names in turn order

	// MCP Server and World State
	MCPServer   *mcp.Server
	World       *mcpsim.WorldState
	MemoryStore *memory.Store
}

// NewSimulation creates a new simulation from a scenario.
func NewSimulation(scenario *scenarios.Scenario, configDir string) *Simulation {
	// Create world state from scenario
	world := mcpsim.NewWorldState(
		scenario.Basics.Location,
		scenario.Basics.Atmosphere,
	)

	// Create MCP server with simulation tools
	mcpServer := mcpsim.NewSimulationServer(world)

	return &Simulation{
		Scenario:  scenario,
		Agents:    make(map[string]*Agent),
		ConfigDir: configDir,
		TurnOrder: make([]string, 0),
		MCPServer: mcpServer,
		World:     world,
	}
}

// Initialize sets up the simulation by loading characters and creating agents.
func (s *Simulation) Initialize(ctx context.Context) error {
	// Load providers configuration
	providersPath := path.Join(s.ConfigDir, "providers.toml")
	providers, err := config.LoadProvidersFromFile(providersPath)
	if err != nil {
		return fmt.Errorf("failed to load providers: %w", err)
	}

	// Load embeddings configuration
	embeddings, err := config.LoadEmbeddingsFromFile(providersPath)
	if err != nil {
		return fmt.Errorf("failed to load embeddings: %w", err)
	}

	// Determine which embedding to use (from scenario defaults)
	if s.Scenario.Basics.Defaults == nil || s.Scenario.Basics.Defaults.Embedding == "" {
		return fmt.Errorf("no embedding configured in scenario defaults")
	}

	embeddingName := s.Scenario.Basics.Defaults.Embedding
	embedding, err := embeddings.Get(embeddingName)
	if err != nil {
		return fmt.Errorf("failed to get embedding '%s': %w", embeddingName, err)
	}

	// Get the provider for this embedding
	embeddingProvider, ok := providers.Providers[embedding.Provider]
	if !ok {
		return fmt.Errorf("embedding provider '%s' not found for embedding '%s'", embedding.Provider, embeddingName)
	}

	fmt.Printf("Validating embedding model availability (%s via %s)...\n", embedding.Model, embedding.Provider)
	if err := config.ValidateEmbeddingModel(embeddingProvider); err != nil {
		return fmt.Errorf(`embedding model validation failed: %w

The simulation requires %s for memory operations.

To fix:
  1. Ensure %s is running
  2. Pull the model: %s pull %s
  3. Retry the simulation
`, err, embedding.Model, embedding.Provider, embedding.Provider, embedding.Model)
	}
	fmt.Printf("✓ Embedding model validated\n\n")

	// Initialize memory store
	fmt.Printf("Initializing memory store...\n")
	embedder := memory.NewOllamaEmbedder(embeddingProvider)
	s.MemoryStore = memory.NewStore(embedder)

	// Seed scenario context (shared across all agents)
	fmt.Printf("Seeding scenario memories...\n")
	if err := memory.SeedScenario(ctx, s.MemoryStore, s.Scenario); err != nil {
		return fmt.Errorf("failed to seed scenario: %w", err)
	}
	fmt.Printf("✓ Seeded %d scenario memories\n", s.MemoryStore.CountByFilter(memory.Filter{Type: "scene"}))

	// Load models configuration
	modelsDir := path.Join(s.ConfigDir, "models")
	models, err := config.LoadModelsFromDir(modelsDir)
	if err != nil {
		return fmt.Errorf("failed to load models: %w", err)
	}

	// Create agents from scenario
	for agentName, agentConfig := range s.Scenario.Agents {
		// Load character definition
		characterPath := path.Join(s.ConfigDir, "characters", agentConfig.Character+".toml")
		character, err := scenarios.LoadCharacterFromFile(characterPath)
		if err != nil {
			return fmt.Errorf("failed to load character %s for agent %s: %w", agentConfig.Character, agentName, err)
		}

		// Determine which model to use
		modelName := agentConfig.Model
		// Use scenario defaults if not specified at agent level
		if modelName == "" && s.Scenario.Basics.Defaults != nil {
			modelName = s.Scenario.Basics.Defaults.Model
		}
		if modelName == "" {
			return fmt.Errorf("agent %s missing model configuration", agentName)
		}

		// Get model config
		model, ok := models[modelName]
		if !ok {
			return fmt.Errorf("model %s not found for agent %s", modelName, agentName)
		}

		// Get provider from model config
		providerName := model.Provider
		if providerName == "" {
			return fmt.Errorf("model %s does not specify a provider", modelName)
		}

		provider, ok := providers.Providers[providerName]
		if !ok {
			return fmt.Errorf("provider %s (from model %s) not found for agent %s", providerName, modelName, agentName)
		}

		// Create LLM client
		client, err := NewClient(provider, model)
		if err != nil {
			return fmt.Errorf("failed to create client for agent %s: %w", agentName, err)
		}

		// Create agent
		// Use model.Name (API model ID) instead of modelName (map key)
		agent := NewAgent(agentName, character, client, providerName, model.Name)

		// Apply initial state overrides from scenario
		agent.ApplyInitialState(agentConfig.Initial)

		// Seed character memories for this agent
		fmt.Printf("  Seeding memories for %s...\n", agentName)
		if err := memory.SeedCharacter(ctx, s.MemoryStore, agentName, character); err != nil {
			return fmt.Errorf("failed to seed character memories for %s: %w", agentName, err)
		}

		// Store agent
		s.Agents[agentName] = agent

		// Add to turn order
		s.TurnOrder = append(s.TurnOrder, agentName)

		// Register agent in world state
		s.World.AddAgent(agentName, agent.State.Position)

		fmt.Printf("  ✓ Initialized agent: %s (character: %s, provider: %s, model: %s)\n",
			agentName, agentConfig.Character, providerName, modelName)
	}

	// Seed knowledge about other characters for each agent
	fmt.Printf("\nSeeding inter-character knowledge...\n")
	for agentName := range s.Scenario.Agents {
		for otherAgentName, otherAgentConfig := range s.Scenario.Agents {
			if agentName == otherAgentName {
				continue
			}

			// Load other character
			otherCharacterPath := path.Join(s.ConfigDir, "characters", otherAgentConfig.Character+".toml")
			otherCharacter, err := scenarios.LoadCharacterFromFile(otherCharacterPath)
			if err != nil {
				return fmt.Errorf("failed to load character %s: %w", otherAgentConfig.Character, err)
			}

			// Seed knowledge
			if err := memory.SeedOtherCharacter(ctx, s.MemoryStore, agentName, otherAgentName, otherCharacter); err != nil {
				return fmt.Errorf("failed to seed knowledge about %s for %s: %w", otherAgentName, agentName, err)
			}
		}
	}

	fmt.Printf("✓ Memory store initialized with %d total memories\n\n", s.MemoryStore.Count())

	// Register memory tools with MCP server
	s.MCPServer.RegisterTool(mcpsim.NewQuerySelfTool(s.MemoryStore))
	s.MCPServer.RegisterTool(mcpsim.NewQueryBackgroundTool(s.MemoryStore))
	s.MCPServer.RegisterTool(mcpsim.NewQueryCommunicationStyleTool(s.MemoryStore))
	s.MCPServer.RegisterTool(mcpsim.NewQuerySceneTool(s.MemoryStore))
	s.MCPServer.RegisterTool(mcpsim.NewQueryCharacterTool(s.MemoryStore))
	s.MCPServer.RegisterTool(mcpsim.NewQueryMemoryTool(s.MemoryStore))

	return nil
}

// Start begins the simulation execution.
// Runs multiple turns until goals are completed or max turns is reached.
func (s *Simulation) Start(ctx context.Context) error {
	if len(s.Agents) == 0 {
		return fmt.Errorf("no agents initialized")
	}

	// Display scenario information
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("     Starting Simulation: %s\n", s.Scenario.Basics.Name)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	fmt.Printf("%s\n", s.Scenario.Basics.Description)
	fmt.Printf("\nLocation: %s\n", s.Scenario.Basics.Location)
	fmt.Printf("Time: %s\n", s.Scenario.Basics.TOD)
	if s.Scenario.Basics.Atmosphere != "" {
		fmt.Printf("Atmosphere: %s\n", s.Scenario.Basics.Atmosphere)
	}

	fmt.Printf("\nAgents:\n")
	for _, agentName := range s.TurnOrder {
		agent := s.Agents[agentName]
		fmt.Printf("  • %s (%s)\n", agentName, agent.Character.Basics.Archetype)
	}

	// Initialize goals in world state
	fmt.Printf("\nGoals:\n")
	for name, goal := range s.Scenario.Goals {
		fmt.Printf("  • %s: %s\n", name, goal.Description)

		// Create interactive goal in world state
		s.World.Goals[name] = mcpsim.NewInteractiveGoal(
			name,
			goal.Description,
			"consensus", // Default to consensus for now
			goal.Priority,
		)
	}

	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Multi-turn loop with two phases: deliberation and voting
	maxTurns := 10
	for turn := 1; turn <= maxTurns; turn++ {
		s.World.CurrentTurn = turn
		fmt.Printf("\n\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("                      Turn %d\n", turn)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

		// Phase 1: Deliberation - agents perceive, discuss, and propose solutions
		fmt.Printf("\n─── Deliberation Phase ───\n")
		deliberationTools := s.getDeliberationTools()
		deliberationSituation := s.buildDeliberationPrompt(turn)

		for _, agentName := range s.TurnOrder {
			agent := s.Agents[agentName]

			fmt.Printf("\n[%s]\n", agentName)

			// Create context with agent name
			agentCtx := context.WithValue(ctx, runtime.AgentNameKey, agentName)

			// Track proposals before this agent's turn
			proposalsBefore := s.countProposals()

			// Agent deliberates: perceive, speak, propose
			response, err := agent.Think(agentCtx, deliberationSituation, deliberationTools, s.MCPServer)
			if err != nil {
				return fmt.Errorf("agent %s failed to deliberate: %w", agentName, err)
			}

			// Display response
			if response.Thinking != "" {
				fmt.Printf("  🧠 Reasoning: %s\n", response.Thinking)
			}
			if response.Message != "" {
				fmt.Printf("  💬 Says: \"%s\"\n", response.Message)
			}

			// Show any proposals made
			proposalsAfter := s.countProposals()
			if proposalsAfter > proposalsBefore {
				s.displayNewProposals(agentName)
			}

			// Add to conversation history
			if len(s.World.ConversationHistory) == 0 ||
				s.World.ConversationHistory[len(s.World.ConversationHistory)-1].AgentName != agentName {
				s.World.AddMessage(agentName, response.Message, response.Thinking)
			}

			// Capture episodic memory
			if response.Message != "" {
				s.captureEpisodicMemory(agentCtx, agentName, response.Message, turn)
			}
		}

		// Check for automatic consensus (identical proposals)
		if s.checkAutomaticConsensus(turn) {
			// Goals completed via automatic consensus, skip voting
			fmt.Printf("\n✨ Automatic consensus detected! Skipping voting phase.\n")
		} else {
			// Phase 2: Voting - agents vote on all pending proposals
			fmt.Printf("\n─── Voting Phase ───\n")
		votingTools := s.getVotingTools()
		votingSituation := s.buildVotingPrompt()

		for _, agentName := range s.TurnOrder {
			agent := s.Agents[agentName]

			fmt.Printf("\n[%s]\n", agentName)

			// Create context with agent name
			agentCtx := context.WithValue(ctx, runtime.AgentNameKey, agentName)

			// Track votes before
			votesBefore := s.collectVotes()

			// Agent votes on all pending proposals
			response, err := agent.Think(agentCtx, votingSituation, votingTools, s.MCPServer)
			if err != nil {
				return fmt.Errorf("agent %s failed to vote: %w", agentName, err)
			}

			// Display response
			if response.Thinking != "" {
				fmt.Printf("  🧠 Reasoning: %s\n", response.Thinking)
			}
			if response.Message != "" {
				fmt.Printf("  💬 Says: \"%s\"\n", response.Message)
			}

			// Show any votes cast
			votesAfter := s.collectVotes()
			s.displayNewVotes(agentName, votesBefore, votesAfter)
		}

			// Display voting results
			s.displayVotingResults()
		}

		// Check if all goals are completed
		if s.allGoalsCompleted() {
			fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf("            All Goals Completed!\n")
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			break
		}
	}

	// Final summary
	s.printGoalSummary()
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("          Simulation Complete\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("\nTotal turns: %d\n", s.World.CurrentTurn)
	return nil
}

// getDeliberationTools returns only tools available during deliberation phase.
func (s *Simulation) getDeliberationTools() []map[string]interface{} {
	allowedTools := []string{
		// Memory tools - essential for discovering identity and context
		"query_self", "query_background", "query_communication_style",
		"query_scene", "query_character", "query_memory",
		// Goal and interaction tools
		"list_goals", "view_goal", "perceive", "speak", "propose_solution",
	}
	allTools := s.MCPServer.GetToolDefinitions()

	filtered := []map[string]interface{}{}
	for _, tool := range allTools {
		if fn, ok := tool["function"].(map[string]interface{}); ok {
			if name, ok := fn["name"].(string); ok {
				for _, allowed := range allowedTools {
					if name == allowed {
						filtered = append(filtered, tool)
						break
					}
				}
			}
		}
	}
	return filtered
}

// getVotingTools returns only tools available during voting phase.
func (s *Simulation) getVotingTools() []map[string]interface{} {
	allowedTools := []string{
		// Memory tools - agents still need access to their identity and memories
		"query_self", "query_background", "query_communication_style",
		"query_scene", "query_character", "query_memory",
		// Voting tools
		"view_goal", "vote_on_proposal",
	}
	allTools := s.MCPServer.GetToolDefinitions()

	filtered := []map[string]interface{}{}
	for _, tool := range allTools {
		if fn, ok := tool["function"].(map[string]interface{}); ok {
			if name, ok := fn["name"].(string); ok {
				for _, allowed := range allowedTools {
					if name == allowed {
						filtered = append(filtered, tool)
						break
					}
				}
			}
		}
	}
	return filtered
}

// buildDeliberationPrompt creates the prompt for deliberation phase.
func (s *Simulation) buildDeliberationPrompt(turn int) string {
	if turn == 1 {
		return `DELIBERATION PHASE (Turn 1):

FIRST, discover your identity (required on first turn):
1. Use query_self() to learn who you are
2. Use query_scene() to understand where you are
3. Use query_character(name) to learn about other agents present

THEN, work on the goals:
4. Use list_goals() to see what goals exist
5. Use perceive() to see the current situation
6. Use speak() to discuss with other agents
7. Use propose_solution() to suggest ONE specific solution (not a list of options)

IMPORTANT: Each proposal must be ONE single, concrete choice. Don't propose multiple alternatives - pick ONE.

DO NOT VOTE YET - voting happens in the next phase.`
	}
	return `DELIBERATION PHASE:

1. Use query_memory() to recall what happened previously
2. Use perceive() to see what others just said
3. Use view_goal() to check existing proposals
4. Use speak() to discuss further
5. Use propose_solution() to suggest ONE specific solution (not multiple options)

IMPORTANT: Each proposal = ONE choice. If you want to suggest alternatives, make separate proposals.

DO NOT VOTE YET - voting is in the next phase.`
}

// buildVotingPrompt creates the prompt for voting phase.
func (s *Simulation) buildVotingPrompt() string {
	// Build a list of all pending proposals across all goals
	proposalList := ""
	for goalName, goal := range s.World.Goals {
		if goal.Status != mcpsim.GoalPending {
			continue
		}

		pendingCount := 0
		for _, proposal := range goal.Proposals {
			if proposal.Status == mcpsim.ProposalPending {
				pendingCount++
			}
		}

		if pendingCount > 0 {
			proposalList += fmt.Sprintf("\nGoal '%s' has %d pending proposal(s)", goalName, pendingCount)
		}
	}

	if proposalList == "" {
		return "VOTING PHASE: No pending proposals to vote on. Just acknowledge and wait for next round."
	}

	return fmt.Sprintf(`VOTING PHASE: Now you must vote on proposals.%s

INSTRUCTIONS:
1. Use view_goal("goal_name") to see all PENDING proposals with their IDs and votes
2. For EACH pending proposal that YOU DID NOT PROPOSE, call vote_on_proposal("goal_name", "proposal_id", "yes" or "no") ONCE
3. Vote based on YOUR character values and preferences
4. SKIP proposals you made - you already auto-voted yes on those when you proposed them
5. If you get an error saying a proposal is "rejected" or "accepted", STOP trying to vote on it - it's already resolved
6. Once you've voted on all relevant proposals, you're done - just say "Voting complete"

Remember: You automatically voted yes on any proposals you made. Only vote on OTHER agents' proposals.`, proposalList)
}

// allGoalsCompleted checks if all goals have been completed.
func (s *Simulation) allGoalsCompleted() bool {
	for _, goal := range s.World.Goals {
		if goal.Status != mcpsim.GoalCompleted {
			return false
		}
	}
	return len(s.World.Goals) > 0 // Only return true if there are goals and they're all complete
}

// countProposals returns the total number of proposals across all goals.
func (s *Simulation) countProposals() int {
	count := 0
	for _, goal := range s.World.Goals {
		count += len(goal.Proposals)
	}
	return count
}

// displayNewProposals shows proposals that were just made by an agent.
func (s *Simulation) displayNewProposals(agentName string) {
	for _, goal := range s.World.Goals {
		for _, proposal := range goal.Proposals {
			if proposal.ProposedBy == agentName && proposal.ProposedAt == s.World.CurrentTurn {
				fmt.Printf("  🔨 Proposes: %s\n", proposal.Description)
			}
		}
	}
}

// collectVotes returns a snapshot of all votes for comparison.
func (s *Simulation) collectVotes() map[string]map[string]map[string]string {
	votes := make(map[string]map[string]map[string]string)
	for goalName, goal := range s.World.Goals {
		votes[goalName] = make(map[string]map[string]string)
		for proposalID, proposal := range goal.Proposals {
			votes[goalName][proposalID] = make(map[string]string)
			for agentName, vote := range proposal.Votes {
				votes[goalName][proposalID][agentName] = vote.Choice
			}
		}
	}
	return votes
}

// displayNewVotes shows votes that were just cast by an agent.
func (s *Simulation) displayNewVotes(agentName string, before, after map[string]map[string]map[string]string) {
	for goalName, goalVotesAfter := range after {
		goalVotesBefore := before[goalName]
		for proposalID, proposalVotesAfter := range goalVotesAfter {
			proposalVotesBefore := goalVotesBefore[proposalID]

			// Check if this agent voted
			voteAfter, hasVoteAfter := proposalVotesAfter[agentName]
			_, hasVoteBefore := proposalVotesBefore[agentName]

			if hasVoteAfter && !hasVoteBefore {
				// Find the proposal to get its description
				goal := s.World.Goals[goalName]
				if proposal, ok := goal.Proposals[proposalID]; ok {
					voteSymbol := "✗"
					if voteAfter == "yes" {
						voteSymbol = "✓"
					}
					fmt.Printf("  🔨 Votes %s on: %s\n", voteSymbol, proposal.Description)
				}
			}
		}
	}
}

// displayVotingResults shows the outcome of the voting phase.
func (s *Simulation) displayVotingResults() {
	hasResults := false

	for _, goal := range s.World.Goals {
		for _, proposal := range goal.Proposals {
			// Only show proposals that were resolved this turn
			if proposal.ResolvedAt == s.World.CurrentTurn {
				if !hasResults {
					fmt.Printf("\nResults:\n")
					hasResults = true
				}

				yesCount := 0
				noCount := 0
				for _, vote := range proposal.Votes {
					if vote.Choice == "yes" {
						yesCount++
					} else {
						noCount++
					}
				}

				switch proposal.Status {
				case mcpsim.ProposalAccepted:
					fmt.Printf("  ✓ %s - Accepted (%d yes, %d no)\n", proposal.Description, yesCount, noCount)
				case mcpsim.ProposalRejected:
					fmt.Printf("  ✗ %s - Rejected (%d yes, %d no)\n", proposal.Description, yesCount, noCount)
				}
			}
		}
	}
}

// printGoalSummary displays a summary of goal completion.
func (s *Simulation) printGoalSummary() {
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("                 Goal Summary\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for _, goal := range s.World.Goals {
		statusSymbol := "○"
		statusText := string(goal.Status)

		switch goal.Status {
		case mcpsim.GoalCompleted:
			statusSymbol = "✓"
			statusText = "COMPLETED"
		case mcpsim.GoalFailed:
			statusSymbol = "✗"
			statusText = "FAILED"
		}

		fmt.Printf("%s %s: %s\n", statusSymbol, goal.Name, statusText)

		if goal.Status == mcpsim.GoalCompleted {
			fmt.Printf("  Completed at turn: %d\n", goal.CompletedAt)

			// Show accepted proposal
			for _, proposal := range goal.Proposals {
				if proposal.Status == mcpsim.ProposalAccepted {
					fmt.Printf("  Solution: %s\n", proposal.Description)
					fmt.Printf("  Proposed by: %s\n", proposal.ProposedBy)

					// Show who voted yes
					voters := []string{}
					for agentName, vote := range proposal.Votes {
						if vote.Choice == "yes" {
							voters = append(voters, agentName)
						}
					}
					if len(voters) > 0 {
						fmt.Printf("  Votes: ")
						for i, voter := range voters {
							if i > 0 {
								fmt.Printf(", ")
							}
							fmt.Printf("%s (yes)", voter)
						}
						fmt.Printf("\n")
					}
				}
			}
		}
		fmt.Printf("\n")
	}
}

// captureEpisodicMemory stores agent dialogue and actions as episodic memories.
func (s *Simulation) captureEpisodicMemory(ctx context.Context, agentName, content string, turn int) {
	if s.MemoryStore == nil {
		return
	}

	// Format the content with speaker
	episodicContent := fmt.Sprintf("%s said: %s", agentName, content)

	// Embed the content
	embedding, err := s.MemoryStore.Embed(ctx, episodicContent)
	if err != nil {
		// Log error but don't fail the simulation
		fmt.Printf("  Warning: failed to embed episodic memory: %v\n", err)
		return
	}

	// Store as episodic memory
	s.MemoryStore.Add(memory.Memory{
		Content:   episodicContent,
		Embedding: embedding,
		Metadata: map[string]string{
			"type":     "episodic",
			"category": "dialogue",
			"turn":     fmt.Sprintf("%d", turn),
			"speaker":  agentName,
		},
	})
}

// checkAutomaticConsensus detects when all agents have made identical proposals.
// If consensus is detected, auto-accepts the proposal and returns true.
func (s *Simulation) checkAutomaticConsensus(turn int) bool {
	foundConsensus := false

	for _, goal := range s.World.Goals {
		// Only check pending goals
		if goal.Status != mcpsim.GoalPending {
			continue
		}

		// Get all proposals made this turn
		turnProposals := make([]*mcpsim.Proposal, 0)
		for _, proposal := range goal.Proposals {
			if proposal.ProposedAt == turn && proposal.Status == mcpsim.ProposalPending {
				turnProposals = append(turnProposals, proposal)
			}
		}

		// Need exactly as many proposals as agents
		if len(turnProposals) != len(s.TurnOrder) {
			continue
		}

		// Check if all proposals have identical descriptions
		if len(turnProposals) == 0 {
			continue
		}

		firstDescription := turnProposals[0].Description
		allIdentical := true
		for _, proposal := range turnProposals[1:] {
			if proposal.Description != firstDescription {
				allIdentical = false
				break
			}
		}

		if allIdentical {
			// Auto-accept the first proposal (they're all the same)
			acceptedProposal := turnProposals[0]

			// Mark all agents as having voted yes
			for _, agentName := range s.TurnOrder {
				acceptedProposal.Votes[agentName] = &mcpsim.Vote{
					AgentName: agentName,
					Choice:    "yes",
					VotedAt:   turn,
				}
			}

			// Update proposal status
			acceptedProposal.Status = mcpsim.ProposalAccepted
			acceptedProposal.ResolvedAt = turn

			// Mark other identical proposals as withdrawn
			for _, proposal := range turnProposals[1:] {
				proposal.Status = mcpsim.ProposalWithdrawn
				proposal.ResolvedAt = turn
			}

			// Complete the goal
			goal.CheckConsensus(turn)

			fmt.Printf("\n  🎯 Goal '%s': All agents proposed \"%s\"\n", goal.Name, firstDescription)
			foundConsensus = true
		}
	}

	return foundConsensus
}
