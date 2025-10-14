package scenarios

import (
	"fmt"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// Duration wraps time.Duration to provide human-readable TOML marshaling/unmarshaling.
// Uses the string notation supported by time.ParseDuration.
// Examples: "5m", "1h", "90s", "2h30m", "1h30m", "2h45m30s"
type Duration time.Duration

// MarshalText implements encoding.TextMarshaler
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler
func (d *Duration) UnmarshalText(text []byte) error {
	dur, err := time.ParseDuration(string(text))
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}
	*d = Duration(dur)
	return nil
}

// ToDuration converts Duration to time.Duration
func (d Duration) ToDuration() time.Duration {
	return time.Duration(d)
}

type Goal struct {
	Name                string    `toml:"-"`
	Description         string    `toml:"description"`
	Priority            int       `toml:"priority"`
	Assignment          []string  `toml:"assignment"`
	Type                string    `toml:"type"`
	Deadline            *Duration `toml:"deadline"`
	CompletionThreshold *float64  `toml:"completion_threshold"`
	// ConsensusGoal specific fields
	ConsensusThreshold *float64 `toml:"consensus_threshold"`
	Tags               []string `toml:"tags"`
	// Future goal types would add their specific fields here
}

type InitialState struct {
	Position         string `toml:"position"`
	Condition        int    `toml:"condition"`
	Emotion          string `toml:"emotion"`
	EmotionIntensity int    `toml:"emotion_intensity"`
}

type ScenarioDefaults struct {
	Model     string `toml:"model"`     // References a model name from models/*.toml (which knows its provider)
	Embedding string `toml:"embedding"` // References an embedding name from [embeddings.*]
}

type Agent struct {
	Name      string        `toml:"-"`
	Character string        `toml:"character"`
	Model     string        `toml:"model"` // Optional: override default model for this agent
	Initial   *InitialState `toml:"-"`
}

type BasicScenarioInformation struct {
	Name        string            `toml:"name"`
	Description string            `toml:"description"`
	Tags        []string          `toml:"tags"`
	Location    string            `toml:"location"`
	TOD         string            `toml:"time"`
	Atmosphere  string            `toml:"atmosphere"`
	MaxRuntime  Duration          `toml:"max_runtime"`
	Defaults    *ScenarioDefaults `toml:"defaults"`
}

type Scenario struct {
	Version       string                    `toml:"version"`
	Basics        *BasicScenarioInformation `toml:"scenario"`
	Agents        map[string]*Agent         `toml:"agents"`
	InitialStates map[string]*InitialState  `toml:"initial_state"`
	Goals         map[string]*Goal          `toml:"goals"`
}

func NewScenario() *Scenario {
	return &Scenario{
		Basics:        &BasicScenarioInformation{},
		Agents:        make(map[string]*Agent),
		InitialStates: make(map[string]*InitialState),
		Goals:         make(map[string]*Goal),
	}
}

// LoadScenario creates and populates a Scenario from TOML data.
// It performs post-processing to set implicit fields and defaults:
//   - Agent.Name is set from the map key
//   - Agent.Initial is linked to the corresponding InitialState
//   - Goal.Name is set from the map key
//   - MaxRuntime defaults to "30m" if not specified
func LoadScenario(data []byte) (*Scenario, error) {
	s := NewScenario()
	if err := toml.Unmarshal(data, s); err != nil {
		return nil, err
	}

	// Apply defaults for missing fields
	if s.Basics.MaxRuntime == 0 {
		s.Basics.MaxRuntime = Duration(30 * time.Minute)
	}

	// Set agent names and link initial states
	for name, agent := range s.Agents {
		agent.Name = name
		if initialState, exists := s.InitialStates[name]; exists {
			agent.Initial = initialState
		}
	}

	// Set goal names
	for name, goal := range s.Goals {
		goal.Name = name
	}

	return s, nil
}

// LoadScenarioFromFile loads a scenario definition from a file path.
func LoadScenarioFromFile(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadScenario(data)
}
