package scenarios

import (
	"testing"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScenarioMarshalTOML(t *testing.T) {
	t.Run("minimal scenario", func(t *testing.T) {
		scenario := NewScenario()
		scenario.Version = "1.0.0"
		scenario.Basics.Name = "Test Scenario"
		scenario.Basics.Description = "A simple test scenario"
		scenario.Basics.Location = "Test Location"
		scenario.Basics.TOD = "12:00 PM"

		scenario.Agents["agent1"] = &Agent{
			Character: "pragmatist",
		}

		scenario.Goals["goal1"] = &Goal{
			Description: "Complete the test",
			Priority:    1,
			Assignment:  []string{"agent1"},
			Type:        "ConsensusGoal",
		}

		buf, err := toml.Marshal(scenario)
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		result := string(buf)
		assert.Contains(t, result, "version = '1.0.0'")
		assert.Contains(t, result, "name = 'Test Scenario'")
		assert.Contains(t, result, "description = 'A simple test scenario'")
		assert.Contains(t, result, "location = 'Test Location'")
		assert.Contains(t, result, "time = '12:00 PM'")
		assert.Contains(t, result, "[agents.agent1]")
		assert.Contains(t, result, "character = 'pragmatist'")
		assert.Contains(t, result, "[goals.goal1]")
		assert.Contains(t, result, "description = 'Complete the test'")
	})

	t.Run("scenario with defaults", func(t *testing.T) {
		scenario := NewScenario()
		scenario.Version = "1.0.0"
		scenario.Basics.Name = "Test Scenario"
		scenario.Basics.Description = "Test with defaults"
		scenario.Basics.Location = "Test Location"
		scenario.Basics.TOD = "12:00 PM"
		scenario.Basics.Defaults = &ScenarioDefaults{
			Provider: "anthropic",
			Model:    "claude-3-5-sonnet-20241022",
		}

		scenario.Agents["agent1"] = &Agent{
			Character: "pragmatist",
		}

		scenario.Goals["goal1"] = &Goal{
			Description: "Complete the test",
			Priority:    1,
			Assignment:  []string{"agent1"},
			Type:        "ConsensusGoal",
		}

		buf, err := toml.Marshal(scenario)
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		result := string(buf)
		assert.Contains(t, result, "provider = 'anthropic'")
		assert.Contains(t, result, "model = 'claude-3-5-sonnet-20241022'")
	})

	t.Run("scenario with agent overrides", func(t *testing.T) {
		scenario := NewScenario()
		scenario.Version = "1.0.0"
		scenario.Basics.Name = "Test Scenario"
		scenario.Basics.Description = "Test with agent overrides"
		scenario.Basics.Location = "Test Location"
		scenario.Basics.TOD = "12:00 PM"
		scenario.Basics.Defaults = &ScenarioDefaults{
			Provider: "anthropic",
			Model:    "claude-3-5-sonnet-20241022",
		}

		scenario.Agents["agent1"] = &Agent{
			Character: "pragmatist",
		}

		scenario.Agents["agent2"] = &Agent{
			Character: "enthusiast",
			Provider:  "ollama",
			Model:     "llama3.1:8b",
		}

		scenario.Goals["goal1"] = &Goal{
			Description: "Complete the test",
			Priority:    1,
			Assignment:  []string{"agent1", "agent2"},
			Type:        "ConsensusGoal",
		}

		buf, err := toml.Marshal(scenario)
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		result := string(buf)
		assert.Contains(t, result, "[agents.agent2]")
		assert.Contains(t, result, "provider = 'ollama'")
		assert.Contains(t, result, "model = 'llama3.1:8b'")
	})

	t.Run("scenario with initial state overrides", func(t *testing.T) {
		scenario := NewScenario()
		scenario.Version = "1.0.0"
		scenario.Basics.Name = "Test Scenario"
		scenario.Basics.Description = "Test with initial states"
		scenario.Basics.Location = "Test Location"
		scenario.Basics.TOD = "12:00 PM"

		scenario.Agents["agent1"] = &Agent{
			Character: "pragmatist",
		}

		scenario.InitialStates["agent1"] = &InitialState{
			Position:         "living_room",
			Condition:        100,
			Emotion:          "neutral",
			EmotionIntensity: 5,
		}

		scenario.Goals["goal1"] = &Goal{
			Description: "Complete the test",
			Priority:    1,
			Assignment:  []string{"agent1"},
			Type:        "ConsensusGoal",
		}

		buf, err := toml.Marshal(scenario)
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		result := string(buf)
		assert.Contains(t, result, "[initial_state.agent1]")
		assert.Contains(t, result, "position = 'living_room'")
		assert.Contains(t, result, "condition = 100")
		assert.Contains(t, result, "emotion = 'neutral'")
		assert.Contains(t, result, "emotion_intensity = 5")
	})

	t.Run("scenario with consensus goal", func(t *testing.T) {
		scenario := NewScenario()
		scenario.Version = "1.0.0"
		scenario.Basics.Name = "Test Scenario"
		scenario.Basics.Description = "Test with consensus goal"
		scenario.Basics.Location = "Test Location"
		scenario.Basics.TOD = "12:00 PM"

		scenario.Agents["agent1"] = &Agent{
			Character: "pragmatist",
		}
		scenario.Agents["agent2"] = &Agent{
			Character: "enthusiast",
		}

		threshold := 1.0
		scenario.Goals["restaurant_agreement"] = &Goal{
			Description:        "Agree on which restaurant to eat at tonight",
			Priority:           1,
			Assignment:         []string{"agent1", "agent2"},
			Type:               "ConsensusGoal",
			ConsensusThreshold: &threshold,
			Tags:               []string{"restaurant_choice", "food"},
		}

		buf, err := toml.Marshal(scenario)
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		result := string(buf)
		assert.Contains(t, result, "[goals.restaurant_agreement]")
		assert.Contains(t, result, "consensus_threshold = 1.0")
		assert.Contains(t, result, "tags = ['restaurant_choice', 'food']")
	})

	t.Run("scenario with all optional fields", func(t *testing.T) {
		scenario := NewScenario()
		scenario.Version = "1.0.0"
		scenario.Basics.Name = "Full Scenario"
		scenario.Basics.Description = "Scenario with all optional fields"
		scenario.Basics.Tags = []string{"test", "comprehensive"}
		scenario.Basics.Location = "Test Location"
		scenario.Basics.TOD = "12:00 PM"
		scenario.Basics.Atmosphere = "Tense and urgent"
		maxRuntime := Duration(30 * time.Minute)
		scenario.Basics.MaxRuntime = maxRuntime
		scenario.Basics.Defaults = &ScenarioDefaults{
			Provider: "anthropic",
			Model:    "claude-3-5-sonnet-20241022",
		}

		scenario.Agents["agent1"] = &Agent{
			Character: "pragmatist",
		}

		threshold := 0.8
		deadline := Duration(10 * time.Minute)
		completionThreshold := 0.9
		scenario.Goals["goal1"] = &Goal{
			Description:         "Complete with all fields",
			Priority:            1,
			Assignment:          []string{"agent1"},
			Type:                "ConsensusGoal",
			Deadline:            &deadline,
			CompletionThreshold: &completionThreshold,
			ConsensusThreshold:  &threshold,
			Tags:                []string{"test", "comprehensive"},
		}

		buf, err := toml.Marshal(scenario)
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		result := string(buf)
		assert.Contains(t, result, "tags = ['test', 'comprehensive']")
		assert.Contains(t, result, "atmosphere = 'Tense and urgent'")
		assert.Contains(t, result, "max_runtime = '30m0s'")
		assert.Contains(t, result, "deadline = '10m0s'")
		assert.Contains(t, result, "completion_threshold = 0.9")
	})
}

func TestScenarioUnmarshalTOML(t *testing.T) {
	t.Run("minimal scenario", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[scenario]
name = "Test Scenario"
description = "A simple test scenario"
location = "Test Location"
time = "12:00 PM"

[agents.agent1]
character = "pragmatist"

[goals.goal1]
description = "Complete the test"
priority = 1
assignment = ["agent1"]
type = "ConsensusGoal"
`

		var scenario Scenario
		err := toml.Unmarshal([]byte(tomlData), &scenario)
		require.NoError(t, err)

		assert.Equal(t, "1.0.0", scenario.Version)
		assert.Equal(t, "Test Scenario", scenario.Basics.Name)
		assert.Equal(t, "A simple test scenario", scenario.Basics.Description)
		assert.Equal(t, "Test Location", scenario.Basics.Location)
		assert.Equal(t, "12:00 PM", scenario.Basics.TOD)

		require.Contains(t, scenario.Agents, "agent1")
		assert.Equal(t, "pragmatist", scenario.Agents["agent1"].Character)

		require.Contains(t, scenario.Goals, "goal1")
		assert.Equal(t, "Complete the test", scenario.Goals["goal1"].Description)
		assert.Equal(t, 1, scenario.Goals["goal1"].Priority)
		assert.Equal(t, []string{"agent1"}, scenario.Goals["goal1"].Assignment)
		assert.Equal(t, "ConsensusGoal", scenario.Goals["goal1"].Type)
	})

	t.Run("scenario with defaults", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[scenario]
name = "Test Scenario"
description = "Test with defaults"
location = "Test Location"
time = "12:00 PM"

[scenario.defaults]
provider = "anthropic"
model = "claude-3-5-sonnet-20241022"

[agents.agent1]
character = "pragmatist"

[goals.goal1]
description = "Complete the test"
priority = 1
assignment = ["agent1"]
type = "ConsensusGoal"
`

		var scenario Scenario
		err := toml.Unmarshal([]byte(tomlData), &scenario)
		require.NoError(t, err)

		require.NotNil(t, scenario.Basics.Defaults)
		assert.Equal(t, "anthropic", scenario.Basics.Defaults.Provider)
		assert.Equal(t, "claude-3-5-sonnet-20241022", scenario.Basics.Defaults.Model)
	})

	t.Run("scenario with agent overrides", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[scenario]
name = "Test Scenario"
description = "Test with agent overrides"
location = "Test Location"
time = "12:00 PM"

[scenario.defaults]
provider = "anthropic"
model = "claude-3-5-sonnet-20241022"

[agents.agent1]
character = "pragmatist"

[agents.agent2]
character = "enthusiast"
provider = "ollama"
model = "llama3.1:8b"

[goals.goal1]
description = "Complete the test"
priority = 1
assignment = ["agent1", "agent2"]
type = "ConsensusGoal"
`

		var scenario Scenario
		err := toml.Unmarshal([]byte(tomlData), &scenario)
		require.NoError(t, err)

		require.Contains(t, scenario.Agents, "agent1")
		assert.Equal(t, "pragmatist", scenario.Agents["agent1"].Character)
		assert.Equal(t, "", scenario.Agents["agent1"].Provider)
		assert.Equal(t, "", scenario.Agents["agent1"].Model)

		require.Contains(t, scenario.Agents, "agent2")
		assert.Equal(t, "enthusiast", scenario.Agents["agent2"].Character)
		assert.Equal(t, "ollama", scenario.Agents["agent2"].Provider)
		assert.Equal(t, "llama3.1:8b", scenario.Agents["agent2"].Model)
	})

	t.Run("scenario with initial state overrides", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[scenario]
name = "Test Scenario"
description = "Test with initial states"
location = "Test Location"
time = "12:00 PM"

[agents.agent1]
character = "pragmatist"

[initial_state.agent1]
position = "living_room"
condition = 100
emotion = "neutral"
emotion_intensity = 5

[goals.goal1]
description = "Complete the test"
priority = 1
assignment = ["agent1"]
type = "ConsensusGoal"
`

		var scenario Scenario
		err := toml.Unmarshal([]byte(tomlData), &scenario)
		require.NoError(t, err)

		require.Contains(t, scenario.InitialStates, "agent1")
		assert.Equal(t, "living_room", scenario.InitialStates["agent1"].Position)
		assert.Equal(t, 100, scenario.InitialStates["agent1"].Condition)
		assert.Equal(t, "neutral", scenario.InitialStates["agent1"].Emotion)
		assert.Equal(t, 5, scenario.InitialStates["agent1"].EmotionIntensity)
	})

	t.Run("scenario with consensus goal", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[scenario]
name = "Test Scenario"
description = "Test with consensus goal"
location = "Test Location"
time = "12:00 PM"

[agents.agent1]
character = "pragmatist"

[agents.agent2]
character = "enthusiast"

[goals.restaurant_agreement]
description = "Agree on which restaurant to eat at tonight"
priority = 1
assignment = ["agent1", "agent2"]
type = "ConsensusGoal"
consensus_threshold = 1.0
tags = ["restaurant_choice", "food"]
`

		var scenario Scenario
		err := toml.Unmarshal([]byte(tomlData), &scenario)
		require.NoError(t, err)

		require.Contains(t, scenario.Goals, "restaurant_agreement")
		goal := scenario.Goals["restaurant_agreement"]
		assert.Equal(t, "Agree on which restaurant to eat at tonight", goal.Description)
		assert.Equal(t, 1, goal.Priority)
		assert.Equal(t, []string{"agent1", "agent2"}, goal.Assignment)
		assert.Equal(t, "ConsensusGoal", goal.Type)
		require.NotNil(t, goal.ConsensusThreshold)
		assert.Equal(t, 1.0, *goal.ConsensusThreshold)
		assert.Equal(t, []string{"restaurant_choice", "food"}, goal.Tags)
	})

	t.Run("scenario with all optional fields", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[scenario]
name = "Full Scenario"
description = "Scenario with all optional fields"
tags = ["test", "comprehensive"]
location = "Test Location"
time = "12:00 PM"
atmosphere = "Tense and urgent"
max_runtime = "30m"

[scenario.defaults]
provider = "anthropic"
model = "claude-3-5-sonnet-20241022"

[agents.agent1]
character = "pragmatist"

[goals.goal1]
description = "Complete with all fields"
priority = 1
assignment = ["agent1"]
type = "ConsensusGoal"
deadline = "10m"
completion_threshold = 0.9
consensus_threshold = 0.8
tags = ["test", "comprehensive"]
`

		var scenario Scenario
		err := toml.Unmarshal([]byte(tomlData), &scenario)
		require.NoError(t, err)

		assert.Equal(t, "1.0.0", scenario.Version)
		assert.Equal(t, []string{"test", "comprehensive"}, scenario.Basics.Tags)
		assert.Equal(t, "Tense and urgent", scenario.Basics.Atmosphere)
		assert.Equal(t, Duration(30*time.Minute), scenario.Basics.MaxRuntime)

		require.NotNil(t, scenario.Basics.Defaults)
		assert.Equal(t, "anthropic", scenario.Basics.Defaults.Provider)

		require.Contains(t, scenario.Goals, "goal1")
		goal := scenario.Goals["goal1"]
		require.NotNil(t, goal.Deadline)
		assert.Equal(t, Duration(10*time.Minute), *goal.Deadline)
		require.NotNil(t, goal.CompletionThreshold)
		assert.Equal(t, 0.9, *goal.CompletionThreshold)
		require.NotNil(t, goal.ConsensusThreshold)
		assert.Equal(t, 0.8, *goal.ConsensusThreshold)
		assert.Equal(t, []string{"test", "comprehensive"}, goal.Tags)
	})

	t.Run("scenario with string duration notation", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[scenario]
name = "Duration Test Scenario"
description = "Test string duration parsing"
location = "Test Location"
time = "12:00 PM"
max_runtime = "30m"

[agents.agent1]
character = "pragmatist"

[goals.goal1]
description = "Complete the test"
priority = 1
assignment = ["agent1"]
type = "ConsensusGoal"
deadline = "10m"
`

		var scenario Scenario
		err := toml.Unmarshal([]byte(tomlData), &scenario)
		require.NoError(t, err)

		assert.Equal(t, Duration(30*time.Minute), scenario.Basics.MaxRuntime)

		require.Contains(t, scenario.Goals, "goal1")
		goal := scenario.Goals["goal1"]
		require.NotNil(t, goal.Deadline)
		assert.Equal(t, Duration(10*time.Minute), *goal.Deadline)
	})

	t.Run("scenario with various duration formats", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[scenario]
name = "Duration Format Test"
description = "Test various duration formats"
location = "Test Location"
time = "12:00 PM"
max_runtime = "2h30m"

[agents.agent1]
character = "pragmatist"

[goals.goal1]
description = "Short goal"
priority = 1
assignment = ["agent1"]
type = "ConsensusGoal"
deadline = "5m"

[goals.goal2]
description = "Medium goal"
priority = 2
assignment = ["agent1"]
type = "ConsensusGoal"
deadline = "1h"

[goals.goal3]
description = "Long goal"
priority = 3
assignment = ["agent1"]
type = "ConsensusGoal"
deadline = "90s"
`

		var scenario Scenario
		err := toml.Unmarshal([]byte(tomlData), &scenario)
		require.NoError(t, err)

		assert.Equal(t, Duration(2*time.Hour+30*time.Minute), scenario.Basics.MaxRuntime)

		require.Contains(t, scenario.Goals, "goal1")
		assert.Equal(t, Duration(5*time.Minute), *scenario.Goals["goal1"].Deadline)

		require.Contains(t, scenario.Goals, "goal2")
		assert.Equal(t, Duration(1*time.Hour), *scenario.Goals["goal2"].Deadline)

		require.Contains(t, scenario.Goals, "goal3")
		assert.Equal(t, Duration(90*time.Second), *scenario.Goals["goal3"].Deadline)
	})

	t.Run("multiple agents and goals", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[scenario]
name = "Multi-Agent Scenario"
description = "Scenario with multiple agents and goals"
location = "Conference Room"
time = "2:00 PM"

[agents.alice]
character = "pragmatist"

[agents.bob]
character = "enthusiast"

[agents.charlie]
character = "cynic"

[goals.goal1]
description = "First goal"
priority = 1
assignment = ["alice", "bob"]
type = "ConsensusGoal"

[goals.goal2]
description = "Second goal"
priority = 2
assignment = ["bob", "charlie"]
type = "ConsensusGoal"

[goals.goal3]
description = "Third goal"
priority = 1
assignment = ["alice", "bob", "charlie"]
type = "ConsensusGoal"
`

		var scenario Scenario
		err := toml.Unmarshal([]byte(tomlData), &scenario)
		require.NoError(t, err)

		assert.Len(t, scenario.Agents, 3)
		assert.Contains(t, scenario.Agents, "alice")
		assert.Contains(t, scenario.Agents, "bob")
		assert.Contains(t, scenario.Agents, "charlie")

		assert.Len(t, scenario.Goals, 3)
		assert.Contains(t, scenario.Goals, "goal1")
		assert.Contains(t, scenario.Goals, "goal2")
		assert.Contains(t, scenario.Goals, "goal3")
	})
}

func TestScenarioRoundTrip(t *testing.T) {
	t.Run("minimal scenario round trip", func(t *testing.T) {
		original := NewScenario()
		original.Version = "1.0.0"
		original.Basics.Name = "Test Scenario"
		original.Basics.Description = "A simple test scenario"
		original.Basics.Location = "Test Location"
		original.Basics.TOD = "12:00 PM"

		original.Agents["agent1"] = &Agent{
			Character: "pragmatist",
		}

		original.Goals["goal1"] = &Goal{
			Description: "Complete the test",
			Priority:    1,
			Assignment:  []string{"agent1"},
			Type:        "ConsensusGoal",
		}

		// Marshal to TOML
		buf, err := toml.Marshal(original)
		require.NoError(t, err)

		// Unmarshal back
		var decoded Scenario
		err = toml.Unmarshal(buf, &decoded)
		require.NoError(t, err)

		// Verify equality
		assert.Equal(t, original.Version, decoded.Version)
		assert.Equal(t, original.Basics.Name, decoded.Basics.Name)
		assert.Equal(t, original.Basics.Description, decoded.Basics.Description)
		assert.Equal(t, original.Basics.Location, decoded.Basics.Location)
		assert.Equal(t, original.Basics.TOD, decoded.Basics.TOD)
		assert.Len(t, decoded.Agents, len(original.Agents))
		assert.Len(t, decoded.Goals, len(original.Goals))
	})

	t.Run("full scenario round trip", func(t *testing.T) {
		original := NewScenario()
		original.Version = "1.0.0"
		original.Basics.Name = "Full Scenario"
		original.Basics.Description = "Complete scenario with all fields"
		original.Basics.Tags = []string{"test", "comprehensive"}
		original.Basics.Location = "Test Location"
		original.Basics.TOD = "12:00 PM"
		original.Basics.Atmosphere = "Tense and urgent"
		original.Basics.MaxRuntime = Duration(30 * time.Minute)
		original.Basics.Defaults = &ScenarioDefaults{
			Provider: "anthropic",
			Model:    "claude-3-5-sonnet-20241022",
		}

		original.Agents["agent1"] = &Agent{
			Character: "pragmatist",
		}

		original.Agents["agent2"] = &Agent{
			Character: "enthusiast",
			Provider:  "ollama",
			Model:     "llama3.1:8b",
		}

		original.InitialStates["agent1"] = &InitialState{
			Position:         "living_room",
			Condition:        100,
			Emotion:          "neutral",
			EmotionIntensity: 5,
		}

		threshold := 0.8
		deadline := Duration(10 * time.Minute)
		completionThreshold := 0.9
		original.Goals["goal1"] = &Goal{
			Description:         "Complete with all fields",
			Priority:            1,
			Assignment:          []string{"agent1", "agent2"},
			Type:                "ConsensusGoal",
			Deadline:            &deadline,
			CompletionThreshold: &completionThreshold,
			ConsensusThreshold:  &threshold,
			Tags:                []string{"test", "comprehensive"},
		}

		// Marshal to TOML
		buf, err := toml.Marshal(original)
		require.NoError(t, err)

		// Unmarshal back
		var decoded Scenario
		err = toml.Unmarshal(buf, &decoded)
		require.NoError(t, err)

		// Verify core fields
		assert.Equal(t, original.Version, decoded.Version)
		assert.Equal(t, original.Basics.Name, decoded.Basics.Name)
		assert.Equal(t, original.Basics.Tags, decoded.Basics.Tags)
		assert.Equal(t, original.Basics.Atmosphere, decoded.Basics.Atmosphere)
		assert.Equal(t, original.Basics.MaxRuntime, decoded.Basics.MaxRuntime)

		// Verify defaults
		require.NotNil(t, decoded.Basics.Defaults)
		assert.Equal(t, original.Basics.Defaults.Provider, decoded.Basics.Defaults.Provider)
		assert.Equal(t, original.Basics.Defaults.Model, decoded.Basics.Defaults.Model)

		// Verify agents
		assert.Len(t, decoded.Agents, len(original.Agents))
		assert.Equal(t, original.Agents["agent2"].Provider, decoded.Agents["agent2"].Provider)
		assert.Equal(t, original.Agents["agent2"].Model, decoded.Agents["agent2"].Model)

		// Verify initial states
		assert.Len(t, decoded.InitialStates, len(original.InitialStates))
		assert.Equal(t, original.InitialStates["agent1"].Position, decoded.InitialStates["agent1"].Position)

		// Verify goals
		assert.Len(t, decoded.Goals, len(original.Goals))
		require.NotNil(t, decoded.Goals["goal1"].Deadline)
		assert.Equal(t, *original.Goals["goal1"].Deadline, *decoded.Goals["goal1"].Deadline)
		require.NotNil(t, decoded.Goals["goal1"].ConsensusThreshold)
		assert.Equal(t, *original.Goals["goal1"].ConsensusThreshold, *decoded.Goals["goal1"].ConsensusThreshold)
	})

	t.Run("multiple round trips preserve data", func(t *testing.T) {
		original := NewScenario()
		original.Version = "1.0.0"
		original.Basics.Name = "Multi-Round Scenario"
		original.Basics.Description = "Test multiple round trips"
		original.Basics.Location = "Test Location"
		original.Basics.TOD = "12:00 PM"

		original.Agents["agent1"] = &Agent{
			Character: "pragmatist",
		}

		original.Goals["goal1"] = &Goal{
			Description: "Complete the test",
			Priority:    1,
			Assignment:  []string{"agent1"},
			Type:        "ConsensusGoal",
		}

		current := original
		for i := 0; i < 3; i++ {
			buf, err := toml.Marshal(current)
			require.NoError(t, err, "marshal iteration %d failed", i)

			var decoded Scenario
			err = toml.Unmarshal(buf, &decoded)
			require.NoError(t, err, "unmarshal iteration %d failed", i)

			assert.Equal(t, original.Version, decoded.Version, "version changed after iteration %d", i)
			assert.Equal(t, original.Basics.Name, decoded.Basics.Name, "name changed after iteration %d", i)
			assert.Len(t, decoded.Agents, len(original.Agents), "agents changed after iteration %d", i)
			assert.Len(t, decoded.Goals, len(original.Goals), "goals changed after iteration %d", i)

			current = &decoded
		}
	})
}

func TestNewScenario(t *testing.T) {
	t.Run("creates valid scenario", func(t *testing.T) {
		scenario := NewScenario()
		require.NotNil(t, scenario)
		require.NotNil(t, scenario.Basics)
		require.NotNil(t, scenario.Agents)
		require.NotNil(t, scenario.InitialStates)
		require.NotNil(t, scenario.Goals)
	})

	t.Run("can marshal new scenario", func(t *testing.T) {
		scenario := NewScenario()
		scenario.Version = "1.0.0"
		scenario.Basics.Name = "New Scenario"
		scenario.Basics.Description = "Test new scenario"
		scenario.Basics.Location = "Test Location"
		scenario.Basics.TOD = "12:00 PM"

		buf, err := toml.Marshal(scenario)
		require.NoError(t, err)
		require.NotEmpty(t, buf)
	})

	t.Run("can add agents to new scenario", func(t *testing.T) {
		scenario := NewScenario()
		scenario.Agents["test"] = &Agent{
			Character: "pragmatist",
		}

		assert.Len(t, scenario.Agents, 1)
		assert.Contains(t, scenario.Agents, "test")
	})

	t.Run("can add goals to new scenario", func(t *testing.T) {
		scenario := NewScenario()
		scenario.Goals["test"] = &Goal{
			Description: "Test goal",
			Priority:    1,
			Type:        "ConsensusGoal",
		}

		assert.Len(t, scenario.Goals, 1)
		assert.Contains(t, scenario.Goals, "test")
	})
}

func TestLoadScenario(t *testing.T) {
	t.Run("loads minimal scenario", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[scenario]
name = "Test Scenario"
description = "A simple test scenario"
location = "Test Location"
time = "12:00 PM"

[agents.agent1]
character = "pragmatist"

[goals.goal1]
description = "Complete the test"
priority = 1
assignment = ["agent1"]
type = "ConsensusGoal"
`

		scenario, err := LoadScenario([]byte(tomlData))
		require.NoError(t, err)

		assert.Equal(t, "1.0.0", scenario.Version)
		assert.Equal(t, "Test Scenario", scenario.Basics.Name)

		// Check agent name is set
		require.Contains(t, scenario.Agents, "agent1")
		assert.Equal(t, "agent1", scenario.Agents["agent1"].Name)
		assert.Equal(t, "pragmatist", scenario.Agents["agent1"].Character)

		// Check goal name is set
		require.Contains(t, scenario.Goals, "goal1")
		assert.Equal(t, "goal1", scenario.Goals["goal1"].Name)
		assert.Equal(t, "Complete the test", scenario.Goals["goal1"].Description)
	})

	t.Run("links initial states to agents", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[scenario]
name = "Test Scenario"
description = "Test with initial states"
location = "Test Location"
time = "12:00 PM"

[agents.alice]
character = "pragmatist"

[agents.bob]
character = "enthusiast"

[initial_state.alice]
position = "living_room"
condition = 100
emotion = "neutral"
emotion_intensity = 5

[goals.goal1]
description = "Complete the test"
priority = 1
assignment = ["alice", "bob"]
type = "ConsensusGoal"
`

		scenario, err := LoadScenario([]byte(tomlData))
		require.NoError(t, err)

		// Check alice has initial state linked
		require.Contains(t, scenario.Agents, "alice")
		require.NotNil(t, scenario.Agents["alice"].Initial)
		assert.Equal(t, "living_room", scenario.Agents["alice"].Initial.Position)
		assert.Equal(t, 100, scenario.Agents["alice"].Initial.Condition)
		assert.Equal(t, "neutral", scenario.Agents["alice"].Initial.Emotion)
		assert.Equal(t, 5, scenario.Agents["alice"].Initial.EmotionIntensity)

		// Check bob has no initial state (nil)
		require.Contains(t, scenario.Agents, "bob")
		assert.Nil(t, scenario.Agents["bob"].Initial)
	})

	t.Run("sets names for all agents", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[scenario]
name = "Multi-Agent Scenario"
description = "Scenario with multiple agents"
location = "Conference Room"
time = "2:00 PM"

[agents.alice]
character = "pragmatist"

[agents.bob]
character = "enthusiast"

[agents.charlie]
character = "cynic"

[goals.goal1]
description = "Test goal"
priority = 1
assignment = ["alice", "bob", "charlie"]
type = "ConsensusGoal"
`

		scenario, err := LoadScenario([]byte(tomlData))
		require.NoError(t, err)

		assert.Len(t, scenario.Agents, 3)
		assert.Equal(t, "alice", scenario.Agents["alice"].Name)
		assert.Equal(t, "bob", scenario.Agents["bob"].Name)
		assert.Equal(t, "charlie", scenario.Agents["charlie"].Name)
	})

	t.Run("sets names for all goals", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[scenario]
name = "Multi-Goal Scenario"
description = "Scenario with multiple goals"
location = "Test Location"
time = "12:00 PM"

[agents.agent1]
character = "pragmatist"

[goals.goal1]
description = "First goal"
priority = 1
assignment = ["agent1"]
type = "ConsensusGoal"

[goals.goal2]
description = "Second goal"
priority = 2
assignment = ["agent1"]
type = "ConsensusGoal"

[goals.goal3]
description = "Third goal"
priority = 1
assignment = ["agent1"]
type = "ConsensusGoal"
`

		scenario, err := LoadScenario([]byte(tomlData))
		require.NoError(t, err)

		assert.Len(t, scenario.Goals, 3)
		assert.Equal(t, "goal1", scenario.Goals["goal1"].Name)
		assert.Equal(t, "goal2", scenario.Goals["goal2"].Name)
		assert.Equal(t, "goal3", scenario.Goals["goal3"].Name)
	})

	t.Run("loads scenario with all optional fields", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[scenario]
name = "Full Scenario"
description = "Scenario with all optional fields"
tags = ["test", "comprehensive"]
location = "Test Location"
time = "12:00 PM"
atmosphere = "Tense and urgent"
max_runtime = "30m"

[scenario.defaults]
provider = "anthropic"
model = "claude-3-5-sonnet-20241022"

[agents.agent1]
character = "pragmatist"

[agents.agent2]
character = "enthusiast"
provider = "ollama"
model = "llama3.1:8b"

[initial_state.agent1]
position = "living_room"
condition = 100
emotion = "neutral"
emotion_intensity = 5

[goals.goal1]
description = "Complete with all fields"
priority = 1
assignment = ["agent1", "agent2"]
type = "ConsensusGoal"
deadline = "10m"
completion_threshold = 0.9
consensus_threshold = 0.8
tags = ["test", "comprehensive"]
`

		scenario, err := LoadScenario([]byte(tomlData))
		require.NoError(t, err)

		// Verify scenario basics
		assert.Equal(t, "1.0.0", scenario.Version)
		assert.Equal(t, "Full Scenario", scenario.Basics.Name)
		assert.Equal(t, []string{"test", "comprehensive"}, scenario.Basics.Tags)
		assert.Equal(t, Duration(30*time.Minute), scenario.Basics.MaxRuntime)

		// Verify defaults
		require.NotNil(t, scenario.Basics.Defaults)
		assert.Equal(t, "anthropic", scenario.Basics.Defaults.Provider)

		// Verify agents with names set
		assert.Len(t, scenario.Agents, 2)
		assert.Equal(t, "agent1", scenario.Agents["agent1"].Name)
		assert.Equal(t, "agent2", scenario.Agents["agent2"].Name)
		assert.Equal(t, "ollama", scenario.Agents["agent2"].Provider)

		// Verify initial state linked
		require.NotNil(t, scenario.Agents["agent1"].Initial)
		assert.Equal(t, "living_room", scenario.Agents["agent1"].Initial.Position)
		assert.Nil(t, scenario.Agents["agent2"].Initial)

		// Verify goal with name set
		assert.Equal(t, "goal1", scenario.Goals["goal1"].Name)
		require.NotNil(t, scenario.Goals["goal1"].Deadline)
		assert.Equal(t, Duration(10*time.Minute), *scenario.Goals["goal1"].Deadline)
	})

	t.Run("returns error for invalid TOML", func(t *testing.T) {
		tomlData := `
version = "1.0.0"
invalid toml syntax here
`

		_, err := LoadScenario([]byte(tomlData))
		require.Error(t, err)
	})

	t.Run("handles empty agents and goals maps", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[scenario]
name = "Empty Scenario"
description = "Scenario with no agents or goals"
location = "Test Location"
time = "12:00 PM"
`

		scenario, err := LoadScenario([]byte(tomlData))
		require.NoError(t, err)

		assert.Equal(t, "1.0.0", scenario.Version)
		assert.Empty(t, scenario.Agents)
		assert.Empty(t, scenario.Goals)
	})

	t.Run("loads duration fields correctly", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[scenario]
name = "Duration Test"
description = "Test duration parsing"
location = "Test Location"
time = "12:00 PM"
max_runtime = "2h30m"

[agents.agent1]
character = "pragmatist"

[goals.goal1]
description = "Test goal"
priority = 1
assignment = ["agent1"]
type = "ConsensusGoal"
deadline = "45m"
`

		scenario, err := LoadScenario([]byte(tomlData))
		require.NoError(t, err)

		assert.Equal(t, Duration(2*time.Hour+30*time.Minute), scenario.Basics.MaxRuntime)
		require.NotNil(t, scenario.Goals["goal1"].Deadline)
		assert.Equal(t, Duration(45*time.Minute), *scenario.Goals["goal1"].Deadline)
	})

	t.Run("applies default max_runtime when not specified", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[scenario]
name = "Default Runtime Test"
description = "Test default max_runtime"
location = "Test Location"
time = "12:00 PM"

[agents.agent1]
character = "pragmatist"

[goals.goal1]
description = "Test goal"
priority = 1
assignment = ["agent1"]
type = "ConsensusGoal"
`

		scenario, err := LoadScenario([]byte(tomlData))
		require.NoError(t, err)

		// Should default to 30 minutes
		assert.Equal(t, Duration(30*time.Minute), scenario.Basics.MaxRuntime)
	})

	t.Run("does not override explicit max_runtime", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[scenario]
name = "Explicit Runtime Test"
description = "Test explicit max_runtime is not overridden"
location = "Test Location"
time = "12:00 PM"
max_runtime = "5m"

[agents.agent1]
character = "pragmatist"

[goals.goal1]
description = "Test goal"
priority = 1
assignment = ["agent1"]
type = "ConsensusGoal"
`

		scenario, err := LoadScenario([]byte(tomlData))
		require.NoError(t, err)

		// Should use the explicitly set value, not the default
		assert.Equal(t, Duration(5*time.Minute), scenario.Basics.MaxRuntime)
	})
}
