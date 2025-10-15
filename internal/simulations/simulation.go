package simulations

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/poiesic/wonda/internal/chronicle"
	"github.com/poiesic/wonda/internal/config"
	"github.com/poiesic/wonda/internal/mcp"
	mcpsim "github.com/poiesic/wonda/internal/mcp/simulation"
	"github.com/poiesic/wonda/internal/memory"
	"github.com/poiesic/wonda/internal/prompts"
	"github.com/poiesic/wonda/internal/runtime"
	"github.com/poiesic/wonda/internal/scenarios"
)

// Simulation represents a running instance of a scenario.
type Simulation struct {
	ID        ulid.ULID // Unique identifier
	Scenario  *scenarios.Scenario
	Agents    map[string]*Agent
	ConfigDir string

	// Turn management
	TurnOrder []string // Agent names in turn order

	// MCP Server and World State
	MCPServer   *mcp.Server
	World       *mcpsim.WorldState
	MemoryStore *memory.Store

	// Chronicle
	chroniclePath   string             // Path to chronicle JSONL file
	chronicleFile   *os.File           // Open file handle for appending
	currentTurnEvents []chronicle.Event // Events being collected for current turn
}

// NewSimulation creates a new simulation from a scenario.
func NewSimulation(scenario *scenarios.Scenario, configDir string) *Simulation {
	// Generate unique ULID for this simulation
	id := ulid.Make()

	// Create world state from scenario
	world := mcpsim.NewWorldState(
		scenario.Basics.Location,
		scenario.Basics.Atmosphere,
	)

	// Create MCP server with simulation tools
	mcpServer := mcpsim.NewSimulationServer(world)

	return &Simulation{
		ID:        id,
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

// initializeChronicle creates the chronicle file and writes the metadata line.
func (s *Simulation) initializeChronicle() error {
	// Generate chronicle filename
	s.chroniclePath = s.getChronicleFilename()

	// Create/open file for writing (append mode)
	file, err := os.Create(s.chroniclePath)
	if err != nil {
		return fmt.Errorf("failed to create chronicle file: %w", err)
	}
	s.chronicleFile = file

	// Create metadata
	metadata := chronicle.NewMetadata(
		s.ID,
		s.Scenario.Basics.Name,
		s.Scenario.Basics.Location,
		s.Scenario.Basics.TOD,
		s.Scenario.Basics.Atmosphere,
	)

	// Write metadata as first JSONL line
	jsonBytes, err := chronicle.ToJSON(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if _, err := s.chronicleFile.WriteString(string(jsonBytes) + "\n"); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// captureEvent adds an event to the current turn's event list.
func (s *Simulation) captureEvent(agentName, dialogue, reasoning string) {
	// Get agent's current emotional state
	agent := s.Agents[agentName]
	event := chronicle.Event{
		AgentName: agentName,
		Dialogue:  dialogue,
		Reasoning: reasoning,
	}

	// Capture emotion if available
	if agent != nil {
		event.Emotion = &chronicle.AgentEmotion{
			Before: chronicle.EmotionState{
				Emotion:   agent.State.Emotion,
				Intensity: agent.State.EmotionIntensity,
			},
			After: chronicle.EmotionState{
				Emotion:   agent.State.Emotion,
				Intensity: agent.State.EmotionIntensity,
			},
		}
	}

	s.currentTurnEvents = append(s.currentTurnEvents, event)
}

// writeTurnToChronicle writes the current turn's events to the chronicle and clears them.
func (s *Simulation) writeTurnToChronicle(turnNumber int) error {
	if s.chronicleFile == nil {
		return nil // Chronicle not initialized
	}

	// Create turn record
	turn := chronicle.Turn{
		Type:   "turn",
		Number: turnNumber,
		Events: s.currentTurnEvents,
	}

	// Convert to JSON
	jsonBytes, err := chronicle.ToJSON(turn)
	if err != nil {
		return fmt.Errorf("failed to marshal turn: %w", err)
	}

	// Write to file
	if _, err := s.chronicleFile.WriteString(string(jsonBytes) + "\n"); err != nil {
		return fmt.Errorf("failed to write turn: %w", err)
	}

	// Clear events for next turn
	s.currentTurnEvents = nil

	return nil
}

// Start begins the simulation execution.
// Runs multiple turns until goals are completed or max turns is reached.
func (s *Simulation) Start(ctx context.Context) error {
	if len(s.Agents) == 0 {
		return fmt.Errorf("no agents initialized")
	}

	// Initialize chronicle
	if err := s.initializeChronicle(); err != nil {
		return fmt.Errorf("failed to initialize chronicle: %w", err)
	}
	defer func() {
		if s.chronicleFile != nil {
			s.chronicleFile.Close()
		}
	}()

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

			// Capture event for chronicle
			s.captureEvent(agentName, response.Message, response.Thinking)
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

			// Capture event for chronicle
			s.captureEvent(agentName, response.Message, response.Thinking)
		}

			// Display voting results
			s.displayVotingResults()
		}

		// Write turn events to chronicle
		if err := s.writeTurnToChronicle(turn); err != nil {
			fmt.Printf("Warning: failed to write turn to chronicle: %v\n", err)
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
	fmt.Printf("Chronicle saved to: %s\n", s.chroniclePath)
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
// Prompts are loaded from the prompts package.
func (s *Simulation) buildDeliberationPrompt(turn int) string {
	var promptName string
	if turn == 1 {
		promptName = "deliberation_turn1"
	} else {
		promptName = "deliberation_other"
	}

	// Get prompt template
	prompt, err := prompts.GetPrompt(promptName)
	if err != nil {
		// Fallback to a simple message if file can't be read
		return fmt.Sprintf("DELIBERATION PHASE (Turn %d): Use available tools to work on goals.", turn)
	}

	return prompt
}

// buildVotingPrompt creates the prompt for voting phase.
// The prompt template is loaded from the prompts package.
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

	// Get prompt template
	promptTemplate, err := prompts.GetPrompt("voting")
	if err != nil {
		// Fallback to simple format if template can't be read
		return fmt.Sprintf("VOTING PHASE: Now you must vote on proposals.%s", proposalList)
	}

	// Parse template and execute with proposal list
	tmpl, err := template.New("voting").Parse(promptTemplate)
	if err != nil {
		// Fallback to simple format if template parsing fails
		return fmt.Sprintf("VOTING PHASE: Now you must vote on proposals.%s", proposalList)
	}

	data := struct {
		ProposalList string
	}{
		ProposalList: proposalList,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		// Fallback to simple format if template execution fails
		return fmt.Sprintf("VOTING PHASE: Now you must vote on proposals.%s", proposalList)
	}

	return buf.String()
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

// getChronicleFilename generates the chronicle filename based on scenario and simulation ID.
// Format: chronicle-<scenario-slug>-<timestamp>-<short-id>.jsonl
func (s *Simulation) getChronicleFilename() string {
	// Generate timestamp
	timestamp := time.Now().Format("20060102-150405")

	// Slugify scenario name
	scenarioSlug := slugify(s.Scenario.Basics.Name)

	// Get first 6 characters of ULID (lowercase)
	shortID := strings.ToLower(s.ID.String()[0:6])

	return fmt.Sprintf("chronicle-%s-%s-%s.jsonl", scenarioSlug, timestamp, shortID)
}

// slugify converts a string to a URL-safe slug.
func slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")

	// Remove non-alphanumeric characters (except hyphens)
	reg := regexp.MustCompile("[^a-z0-9-]+")
	s = reg.ReplaceAllString(s, "")

	// Remove consecutive hyphens
	reg = regexp.MustCompile("-+")
	s = reg.ReplaceAllString(s, "-")

	// Trim hyphens from start and end
	s = strings.Trim(s, "-")

	return s
}
