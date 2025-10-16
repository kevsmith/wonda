package chronicle

import (
	"encoding/json"
	"time"

	"github.com/oklog/ulid/v2"
)

// Metadata is the first line in the chronicle JSONL file.
type Metadata struct {
	Type         string    `json:"type"` // Always "metadata"
	SimulationID string    `json:"simulation_id"`
	Scenario     string    `json:"scenario"`
	Location     string    `json:"location"`
	Time         string    `json:"time"`
	Atmosphere   string    `json:"atmosphere,omitempty"`
	StartTime    time.Time `json:"start_time"`
}

// Turn represents all events that occurred in a single turn.
type Turn struct {
	Type            string           `json:"type"` // Always "turn"
	Number          int              `json:"number"`
	Events          []Event          `json:"events"`
	GoalCompletions []GoalCompletion `json:"goal_completions,omitempty"` // Goals completed this turn
}

// Event captures what one agent did during a turn.
type Event struct {
	AgentName string        `json:"agent_name"`
	Type      string        `json:"type,omitempty"`      // dialogue, action, monologue
	Dialogue  string        `json:"dialogue,omitempty"`  // What they said
	Reasoning string        `json:"reasoning,omitempty"` // LLM thinking
	Emotion   *AgentEmotion `json:"emotion,omitempty"`   // Emotional state change
	Proposals []string      `json:"proposals,omitempty"` // Proposals made
	Votes     []Vote        `json:"votes,omitempty"`     // Votes cast
}

// AgentEmotion captures emotional state before and after an action.
type AgentEmotion struct {
	Before EmotionState `json:"before"`
	After  EmotionState `json:"after"`
}

// EmotionState represents an emotional state at a point in time.
type EmotionState struct {
	Emotion   string `json:"emotion"`   // angry, afraid, happy, sad, neutral
	Intensity int    `json:"intensity"` // 0-10
}

// Vote represents a vote cast on a proposal.
type Vote struct {
	ProposalID string `json:"proposal_id"`
	Choice     string `json:"choice"` // yes, no
}

// GoalCompletion represents a goal that was completed this turn.
type GoalCompletion struct {
	GoalName    string   `json:"goal_name"`
	Status      string   `json:"status"`      // completed, failed
	Solution    string   `json:"solution"`    // The accepted proposal
	ProposedBy  string   `json:"proposed_by"` // Who proposed the solution
	VotedYes    []string `json:"voted_yes"`   // Agents who voted yes
	VotedNo     []string `json:"voted_no"`    // Agents who voted no
	CompletedAt int      `json:"completed_at"` // Turn number
}

// NewMetadata creates a metadata record for the chronicle.
func NewMetadata(id ulid.ULID, scenario, location, tod, atmosphere string) Metadata {
	return Metadata{
		Type:         "metadata",
		SimulationID: id.String(),
		Scenario:     scenario,
		Location:     location,
		Time:         tod,
		Atmosphere:   atmosphere,
		StartTime:    time.Now(),
	}
}

// ToJSON converts a record to JSON bytes for JSONL format.
func ToJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
