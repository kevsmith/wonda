package simulation

// WorldState represents the shared simulation world that all agents exist in.
// This is an MCP resource that tools can read from and modify.
type WorldState struct {
	// Location is the primary scene location
	Location string

	// Atmosphere describes the environmental feel
	Atmosphere string

	// Agents tracks all agents and their positions
	Agents map[string]*AgentInWorld

	// ConversationHistory stores all messages
	ConversationHistory []ConversationMessage

	// Goals tracks interactive goals that agents can work toward
	Goals map[string]*InteractiveGoal

	// CurrentTurn tracks which turn we're on
	CurrentTurn int
}

// AgentInWorld represents an agent's presence in the world.
type AgentInWorld struct {
	Name     string
	Position string // Sublocation (e.g., "coffee_table", "doorway")
	Visible  bool   // Can this agent be perceived by others?
}

// ConversationMessage represents a message in the conversation history.
type ConversationMessage struct {
	AgentName string
	Content   string
	Thinking  string
}

// NewWorldState creates a new world state.
func NewWorldState(location, atmosphere string) *WorldState {
	return &WorldState{
		Location:            location,
		Atmosphere:          atmosphere,
		Agents:              make(map[string]*AgentInWorld),
		ConversationHistory: make([]ConversationMessage, 0),
		Goals:               make(map[string]*InteractiveGoal),
		CurrentTurn:         0,
	}
}

// AddAgent registers an agent in the world.
func (w *WorldState) AddAgent(name, position string) {
	w.Agents[name] = &AgentInWorld{
		Name:     name,
		Position: position,
		Visible:  true,
	}
}

// AddMessage records a message in the conversation history.
func (w *WorldState) AddMessage(agentName, content, thinking string) {
	w.ConversationHistory = append(w.ConversationHistory, ConversationMessage{
		AgentName: agentName,
		Content:   content,
		Thinking:  thinking,
	})
}

// GetNearbyAgents returns all agents at the same position as the querying agent.
func (w *WorldState) GetNearbyAgents(agentName string) []string {
	queryAgent, ok := w.Agents[agentName]
	if !ok {
		return []string{}
	}

	nearby := make([]string, 0)
	for name, agent := range w.Agents {
		if name == agentName {
			continue // Don't include self
		}
		if agent.Position == queryAgent.Position && agent.Visible {
			nearby = append(nearby, name)
		}
	}

	return nearby
}

// GetRecentMessages returns the last N messages from conversation history.
func (w *WorldState) GetRecentMessages(limit int) []ConversationMessage {
	if limit <= 0 || limit > len(w.ConversationHistory) {
		return w.ConversationHistory
	}
	start := len(w.ConversationHistory) - limit
	return w.ConversationHistory[start:]
}
