# Wonda Scenario Definition

**Version**: 1.0.0
**Status**: Draft
**Last Updated**: 2025-10-13

## Overview

This document defines the specification for Wonda scenarios - templates that define simulations for fiction writers. Scenarios specify which characters participate, what goals they pursue, and the initial scene conditions. They are to simulations what classes are to objects: scenarios define the structure, simulations are runtime instances.

## Design Principles

1. **Fiction-First**: Optimized for narrative quality and creative inspiration, not simulation accuracy
2. **TOML as Source**: Human-friendly TOML format for writer accessibility
3. **Separation of Concerns**: Scenarios define WHAT to simulate, not HOW to execute
4. **Character Autonomy**: Minimal prescription - let agents figure out approach based on character traits and goals
5. **Template Pattern**: Scenarios are reusable definitions, simulations are execution instances
6. **File-Based Storage**: Scenarios stored as TOML files in filesystem (not database)

## Core Scenario Structure

### TOML Format

```toml
version = "1.0.0"

[scenario]
# Metadata
name = "Dinner Planning"
description = "Two friends try to agree on a restaurant"
tags = ["dialogue", "consensus", "social"]

# Execution Configuration
max_runtime = "30m"           # Maximum simulation time (Go duration format)

# Scene Context
location = "Alex's apartment - Living room"
time = "6:30 PM"
atmosphere = "Casual, relaxed evening. Light rain outside."

# Default LLM configuration for agents
[scenario.defaults]
provider = "anthropic"
model = "claude-3-5-sonnet-20241022"

# Agents (minimum 1, maximum 50)
# Each agent has a name and references a character archetype
[agents.Alex]
character = "pragmatist"

[agents.Jordan]
character = "enthusiast"

# Optional: Override initial state for agents in this scenario
[initial_state.Alex]
position = "living_room"
condition = 100
emotion = "neutral"
emotion_intensity = 5

[initial_state.Jordan]
position = "living_room"
condition = 100
emotion = "happy"
emotion_intensity = 6

# Goals (minimum 1, maximum 8)
[goals.restaurant_agreement]
description = "Agree on which restaurant to eat at tonight"
priority = 1
assignment = ["Alex", "Jordan"]  # Agent names who have this goal
type = "ConsensusGoal"
consensus_threshold = 1.0        # Both must agree (100%)
tags = ["restaurant_choice", "food"]
```

## Field Descriptions

### Version

**version** (required, semantic versioning)
- Top-level version string in format "X.Y.Z"
- Specifies which version of the scenario file format this file uses
- Enables format compatibility checking and migration
- Example: "1.0.0", "2.1.3"

### Scenario Metadata

**scenario.name** (required, 1-100 characters)
- Human-readable scenario name
- Used for identification and display
- Example: "The Standoff", "Monster Attack", "Dinner Planning"

**scenario.description** (required, 10-1000 characters)
- What the scenario simulates
- Brief summary of the situation
- Example: "Two detectives interrogate a suspect who may know about a kidnapping"

**scenario.tags** (optional)
- Array of categorization labels
- Useful for searching and organizing scenarios
- Examples: ["combat", "rescue"], ["dialogue", "mystery"], ["comedy", "consensus"]

### Execution Configuration

**scenario.max_runtime** (optional, default "30m")
- Maximum real-world time the simulation can run
- Duration format: string notation supported by Go's `time.ParseDuration`
- Examples: `"30s"`, `"5m"`, `"2h"`, `"1h30m"`, `"90s"`, `"2h45m30s"`
- Prevents runaway simulations

**scenario.location** (required)
- Where the scene takes place
- Example: "Downtown alley - Night", "Mayor's office", "Abandoned warehouse"

**scenario.time** (required)
- Time of day or specific timestamp
- Example: "3:00 PM", "Dawn", "Late evening"
- Used for context and time-aware behavior

**scenario.atmosphere** (optional)
- Emotional/environmental tone and contextual conditions
- Can include sensory details, environmental factors, mood, urgency
- Example: "Tense standoff", "Casual conversation", "Chaotic emergency with smoke and sirens"
- Example: "Quiet evening. Light rain outside. Cozy interior lighting."

### Defaults (Optional)

**scenario.defaults.provider** (optional)
- Default LLM provider for all agents in this scenario
- Example: "anthropic", "openai", "google", "ollama"
- Supports self-hosted solutions like Ollama for 100% local execution
- Can be overridden per-agent

**scenario.defaults.model** (optional)
- Default model identifier for all agents in this scenario
- Example: "claude-3-5-sonnet-20241022", "gpt-4", "gemini-pro", "llama3.1:8b"
- Can be overridden per-agent

### Agents (Required, min 1, max 50)

Agents are named instances in the scenario that embody character archetypes. Each agent is defined as `[agents.agent_name]` where `agent_name` is a unique identifier for this scenario.

**Agent name** (from section header)
- Unique identifier for the agent in this scenario
- Format: `[agents.agent_name]` - can use any valid TOML key (quoted if contains spaces)
- Examples: `Alex`, `"Bob Johnson"`, `detective_chen`
- The name is implicit from the TOML structure, not an explicit field

**agent.character** (required)
- Character archetype ID that this agent embodies
- References a character file by ID (filename without `.toml` extension)
- Example: `character = "pragmatist"` loads `characters/pragmatist.toml`
- Character files define personality, traits, and default initial state

**agent.provider** (optional)
- LLM provider for this specific agent
- Overrides scenario.defaults.provider if specified
- Example: `provider = "anthropic"`, `provider = "ollama"`

**agent.model** (optional)
- Model identifier for this specific agent
- Overrides scenario.defaults.model if specified
- Example: `model = "claude-3-5-sonnet-20241022"`, `model = "llama3.1:8b"`

### Initial State Overrides (Optional)

**initial_state.{agent_name}** (optional)
- Override the agent's default initial state for this specific scenario
- Uses agent name as the key
- Available fields:
  - `position`: Starting location in scene (string)
  - `condition`: Health/energy (0-100, default from character file or 100)
  - `emotion`: Starting emotion (default from character file or "neutral")
  - `emotion_intensity`: Emotion strength (0-10, default from character file or 5)

**Example:**
```toml
[agents."Detective Chen"]
character = "negotiator"

[initial_state."Detective Chen"]
position = "outside_store"
condition = 90
emotion = "focused"
emotion_intensity = 7
```

**Note:** Character files are stored separately in `characters/` directory and define reusable archetypes (personality, traits, default state). Agents in scenarios are named instances that reference these character archetypes. See [Character Definition](./character-definition.md) for complete character file format.

### Goals (Required, min 1, max 8)

Goals define success conditions that drive agent behavior and determine simulation completion. Each goal is defined as `[goals.goal_name]` where `goal_name` is a unique identifier that serves as the goal's key in the goals map.

**Goal name** (from section header)
- Unique identifier for the goal, specified in the TOML section header
- Format: `[goals.goal_name]` where goal_name uses snake_case
- Examples: `restaurant_agreement`, `rescue_civilians`, `get_information`
- The name is implicit from the TOML structure, not an explicit field

**goal.description** (required, 10-500 characters)
- Natural language description of what must be accomplished
- Example: "Protect the civilians from the monster"
- Example: "Convince the mayor to evacuate the town"

**goal.priority** (required, 1-5)
- Importance level (1 = highest, 5 = lowest)
- Priority 1 goals are usually required for scenario success
- Lower priorities are nice-to-have or optional objectives

**goal.assignment** (required)
- Array of agent names who have this goal
- Can assign to specific agents: `["Alex", "Jordan"]` or `["Detective Chen", "Officer Kim"]`
- Can assign to all: `["all"]`
- Empty array means no agent assigned (scenario-level tracking only)

**goal.type** (required for MVP)
- Goal evaluation type
- MVP supports: "ConsensusGoal"
- Future: "StateGoal", "RescueGoal", "ProximityGoal", etc.

**Type-specific fields** (varies by goal type)
- Each goal type has additional required/optional fields
- ConsensusGoal: `consensus_threshold` (0.0-1.0), `tags` (array of strings)
- Future goal types will have their own specific fields
- All fields are placed directly in the goal section (no nested parameters table)

**goal.deadline** (optional)
- Time limit that triggers Action Mode
- Duration format: string notation supported by Go's `time.ParseDuration`
- Examples: `"5m"`, `"10m"`, `"1h"`, `"90s"`, `"1h30m"`
- Presence of deadline creates time pressure

**goal.completion_threshold** (optional, default 1.0)
- Minimum evaluation score for success (0.0-1.0)
- Allows partial completion goals
- Example: 0.8 = "80% complete is success"

See [Goal System](./goal-system.md) for evaluation details.

## Goal Types Reference

### ConsensusGoal (MVP)

Characters must agree on a decision or choice.

**Parameters:**
- `consensus_threshold` (float): 0.0-1.0, percentage who must agree (1.0 = unanimous)
- `tags` (array of strings): Tags categorizing what they're agreeing on

**Evaluation:**
Agents report their level of agreement via `assess_goal()` MCP tool. When enough agents report full agreement (based on threshold), the goal completes.

**Example:**
```toml
[goals.dinner_decision]
description = "Decide where to eat dinner tonight"
priority = 1
assignment = ["Alex", "Jordan"]
type = "ConsensusGoal"
consensus_threshold = 1.0
tags = ["restaurant_choice", "decision_making"]
```

### Future Goal Types

Phase 2+ will add:

**StateGoal**: World state reaches target condition
- Parameters: `target_state`, `evaluation_function`
- Example: "All civilians evacuated" (check civilian positions)

**RescueGoal**: Protect entities from threats
- Parameters: `targets`, `threat`, `rescue_threshold`, `distance_threshold`
- Example: "Save 80% of civilians from the monster"

**ProximityGoal**: Reach a location or distance
- Parameters: `target_location`, `characters`, `distance_threshold`
- Example: "Get to the safehouse"

**ThresholdGoal**: Maintain value above/below threshold
- Parameters: `metric`, `threshold`, `comparison`
- Example: "Keep trust level above 0.7"

## Minimal Example

**Scenario file:** `scenarios/quick-chat.toml`
```toml
version = "1.0.0"

[scenario]
name = "Quick Chat"
description = "Two colleagues discuss a project over coffee"
location = "Coffee shop corner table"
time = "10:00 AM"
atmosphere = "Casual morning meeting"

[scenario.defaults]
provider = "anthropic"
model = "claude-3-5-sonnet-20241022"

[agents.Sam]
character = "pragmatist"

[agents.Taylor]
character = "creative"

[goals.project_alignment]
description = "Align on project priorities"
priority = 1
assignment = ["Sam", "Taylor"]
type = "ConsensusGoal"
consensus_threshold = 1.0
tags = ["project_priorities", "planning"]
```

**Character file:** `characters/pragmatist.toml`
```toml
version = "1.0.0"

[basics]
archetype = "The Pragmatist"
description = "Practical project manager who wants clarity"
background = ""
communication_style = "Clear and to-the-point"
decision_style = "Fact-based and pragmatic"
traits = ["organized", "direct"]
skills = []
values = []
```

**Character file:** `characters/creative.toml`
```toml
version = "1.0.0"

[basics]
archetype = "The Creative"
description = "Enthusiastic designer with big ideas"
background = ""
communication_style = "Animated and metaphorical"
decision_style = "Intuition-driven and exploratory"
traits = ["creative", "optimistic"]
skills = []
values = []
```

## Rich Example

**Scenario file:** `scenarios/the-standoff.toml`
```toml
version = "1.0.0"

[scenario]
name = "The Standoff"
description = "A detective negotiates with an armed suspect holding a hostage in a convenience store"
tags = ["tension", "negotiation", "crisis", "dialogue"]
max_runtime = "45m"
location = "24-hour convenience store on Martin Street - Inside"
time = "2:47 AM"
atmosphere = "High tension. Silent except for occasional sirens outside. Harsh fluorescent lighting, visible from outside. Everyone visible to each other and to snipers. Cool night, store AC running. Smell of spilled coffee. 12 officers outside, SWAT on standby. News vans arriving, growing crowd. SWAT commander authorized to intervene at 30-minute mark if no progress."

[scenario.defaults]
provider = "anthropic"
model = "claude-3-5-sonnet-20241022"

[agents."Detective Chen"]
character = "negotiator"

[agents."Marcus Webb"]
character = "desperate-man"
# Override model for this agent - use a different model for distinct personality
model = "claude-3-5-haiku-20241022"

[agents."Sarah Miller"]
character = "hostage"

# Override initial states for agents
[initial_state."Detective Chen"]
position = "outside_store_front"
condition = 90
emotion = "focused"
emotion_intensity = 7

[initial_state."Marcus Webb"]
position = "behind_store_counter"
condition = 60
emotion = "afraid"
emotion_intensity = 9

[initial_state."Sarah Miller"]
position = "behind_store_counter"
condition = 70
emotion = "afraid"
emotion_intensity = 8

[goals.save_everyone]
description = "Get everyone out alive and unharmed"
priority = 1
assignment = ["Detective Chen"]
type = "StateGoal"
deadline = "30m"  # After 30 minutes, SWAT takes over
target_state = "all_alive_and_safe"
hostage_released = true
suspect_surrendered = true

[goals.escape_free]
description = "Escape without going to prison"
priority = 1
assignment = ["Marcus Webb"]
type = "StateGoal"
target_state = "escaped_or_reduced_charges"
freedom_preserved = true

[goals.survive]
description = "Survive and get out safely"
priority = 1
assignment = ["Sarah Miller"]
type = "StateGoal"
target_state = "released_unharmed"
safe = true
```

**Character file:** `characters/negotiator.toml`
```toml
version = "1.0.0"

[basics]
archetype = "The Negotiator"
description = "Veteran hostage negotiator with years of experience. Calm under pressure, highly empathetic."
background = "Former patrol officer who witnessed a botched hostage situation early in career. Trained extensively in crisis intervention and psychology. Has talked down dozens of armed suspects without violence."
communication_style = "Calm, steady voice. Uses open questions. Validates emotions. Strategic pauses. Never threatens or escalates."
decision_style = "Patient and methodical. Prioritizes preserving all lives above all else. Willing to negotiate for hours. Reads micro-expressions and tone."
traits = ["calm", "empathetic", "patient", "observant", "strategic"]
skills = ["crisis_negotiation", "psychology", "reading_people", "de-escalation"]
values = ["preserving_life", "empathy", "patience", "trust_building"]
```

*(Characters `desperate-man.toml` and `hostage.toml` would be similarly structured)*

## Validation Rules

All validation performed during scenario loading:

1. **Required fields**:
   - version (top-level)
   - scenario: name, description, location, time
   - At least 1 agent (via `[agents.agent_name]` sections)
   - Each agent must have a character field
   - At least 1 goal

   Note: scenario.defaults (provider, model) are optional. Agent-level provider/model are also optional and override scenario defaults.

2. **String lengths**:
   - scenario.name: 1-100 characters
   - scenario.description: 10-1000 characters
   - goal.description: 10-500 characters

3. **Map limits**:
   - agents: 1-50 agents (via `[agents.agent_name]` sections)
   - goals: 1-8 (design limit for performance)

4. **Agent and character validation**:
   - Each agent name must be unique within the scenario
   - All agent.character values must reference existing files in `characters/` directory
   - Character files must be valid TOML and conform to character specification
   - Agent names in goal assignments must match defined agents

5. **Enum validation**:
   - initial_state emotion: "neutral", "angry", "afraid", "happy", "sad"
   - goal.priority: 1-5

6. **Range validation**:
   - initial_state condition: 0-100
   - initial_state emotion_intensity: 0-10
   - goal.consensus_threshold: 0.0-1.0

7. **Semantic versioning**: version must match pattern `^\d+\.\d+\.\d+$`

8. **Goal validation**:
   - Each goal section `[goals.goal_name]` must have a unique name
   - Goal names (from section headers) should use snake_case (1-100 characters)
   - Agent names in assignment must match defined agents
   - At least one goal should be assigned to agents (not all unassigned)

9. **Duration format**: max_runtime must be valid Go duration

10. **Initial state overrides**:
    - Keys in initial_state must match agent names defined in `[agents.agent_name]` sections
    - Cannot specify initial state for agents not defined in the scenario

## File Organization

```
wonda/
├── characters/                      # Reusable character archetypes
│   ├── pragmatist.toml             # The pragmatist archetype
│   ├── creative.toml               # The creative archetype
│   ├── enthusiast.toml             # The enthusiast archetype
│   ├── negotiator.toml             # The negotiator archetype
│   ├── desperate-man.toml          # The desperate man archetype
│   ├── hostage.toml                # The hostage/victim archetype
│   ├── optimist.toml               # The optimist archetype
│   ├── cynic.toml                  # The cynic archetype
│   ├── leader.toml                 # The leader archetype
│   ├── specialist.toml             # The specialist archetype
│   ├── investigator.toml           # The investigator archetype
│   ├── person-of-interest.toml     # The person of interest archetype
│   └── library/                    # Additional character archetypes
│       ├── heroes/
│       │   ├── mentor.toml
│       │   └── warrior.toml
│       ├── civilians/
│       │   └── bystander.toml
│       └── antagonists/
│           └── villain.toml
│
└── scenarios/                       # Scenario definitions (reference characters)
    ├── dinner-planning.toml        # Simple consensus example
    ├── the-standoff.toml          # Multi-goal tension example
    ├── monster-attack.toml        # Action mode rescue example
    └── library/
        ├── consensus/
        │   ├── restaurant-choice.toml
        │   └── movie-selection.toml
        ├── rescue/
        │   ├── civilian-evacuation.toml
        │   └── hostage-situation.toml
        └── mystery/
            └── interrogation.toml
```

## CLI Commands

```bash
# Validate scenario
wonda validate scenarios/dinner-planning.toml

# Run simulation from scenario
wonda run scenarios/dinner-planning.toml

# List all available scenarios
wonda scenarios list

# Show scenario details
wonda scenarios show dinner-planning
```

## Loading and Execution Flow

1. **Load Scenario**: Parse scenario TOML file into scenario structure
2. **Load Characters**: For each character ID in scenario, load corresponding character file from `characters/` directory
3. **Apply Initial State Overrides**: Merge scenario-specific initial_state values with character defaults
4. **Validate**: Check all required fields, constraints, and references
5. **Initialize Simulation**: Create simulation instance from scenario template with loaded characters
6. **Instantiate Agents**: Bind characters to LLM providers
7. **Setup Scene**: Initialize environment state
8. **Execute**: Run simulation until goals complete or timeout
9. **Chronicle**: Generate narrative output from events

## Relationship to Simulation Execution

**Key Distinction**: Scenarios are templates, simulations are instances.

- **Scenario**: TOML file defining WHAT to simulate (characters, goals, scene)
- **Simulation**: Runtime instance that executes a scenario
  - Has execution state, events, chronicle
  - Agents with memory and state
  - Can have different outcomes each run
  - Multiple simulations can run from same scenario

**File Persistence**:
- Scenarios: Permanent TOML files in `scenarios/` directory
- Simulations: *(Future)* Can be saved with execution state and chronicle
- Chronicles: Exported as JSON or formatted text after completion

## Example Use Cases

### Use Case 1: Character Development

```toml
version = "1.0.0"

[scenario]
name = "Coffee Shop Encounter"
description = "Two strangers strike up a conversation, revealing contrasting worldviews"
tags = ["dialogue", "character-study", "discovery"]
location = "Busy coffee shop - Shared table"
time = "Morning rush"
atmosphere = "Noisy but cozy, forced proximity"

[agents.Emma]
character = "optimist"

[agents.Michael]
character = "cynic"

[goals.find_common_ground]
description = "Find common ground despite different worldviews"
priority = 2
assignment = ["Emma", "Michael"]
type = "ConsensusGoal"
consensus_threshold = 0.7
tags = ["mutual_understanding", "worldviews"]
```

*(Character files `optimist.toml` and `cynic.toml` would define "The Optimist" and "The Cynic" archetypes)*

### Use Case 2: Action Scene

```toml
version = "1.0.0"

[scenario]
name = "Building Collapse Rescue"
description = "Firefighters have minutes to rescue trapped civilians from collapsing structure"
tags = ["action", "rescue", "time-pressure", "cooperation"]
max_runtime = "15m"
location = "Partially collapsed apartment building - Third floor"
time = "2:30 PM"
atmosphere = "Urgent, dangerous. Thick dust and smoke. Creaking structure - collapse imminent in ~5 minutes. Low visibility. Loud alarms, creaking metal, distant sirens. Hot from fires below."

[agents."Captain Rodriguez"]
character = "leader"

[agents."Tech Specialist Kim"]
character = "specialist"

[goals.rescue_civilians]
description = "Rescue all trapped civilians before building collapses"
priority = 1
assignment = ["Captain Rodriguez", "Tech Specialist Kim"]
type = "RescueGoal"
deadline = "5m"
targets = ["civilian_1", "civilian_2", "civilian_3"]
rescue_threshold = 1.0
time_limit = "5_minutes"
```

*(Character files `leader.toml` and `specialist.toml` would define experienced rescue team archetypes)*

### Use Case 3: Mystery/Investigation

```toml
version = "1.0.0"

[scenario]
name = "The Missing Witness"
description = "Detective interrogates a person of interest who may know where a missing witness is hiding"
tags = ["mystery", "interrogation", "psychological", "tension"]
max_runtime = "30m"
location = "Police station - Interrogation room"
time = "11:00 PM"
atmosphere = "Stark, uncomfortable, tension building. Witness in danger - every hour matters. Person is not under arrest, can leave anytime. Person increasingly nervous, detective patient but time-pressured."

[agents."Detective Rivera"]
character = "investigator"

[agents."Jamie Chen"]
character = "person-of-interest"

[goals.get_information]
description = "Get the person to reveal the witness's location"
priority = 1
assignment = ["Detective Rivera"]
type = "StateGoal"
target_state = "information_revealed"
location_disclosed = true

[goals.protect_friend]
description = "Protect friend's location without going to jail"
priority = 1
assignment = ["Jamie Chen"]
type = "StateGoal"
target_state = "friend_safe_self_safe"
location_secret = true
avoid_charges = true
```

*(Character files `investigator.toml` and `person-of-interest.toml` would define "The Investigator" and "The Person of Interest" archetypes)*

### Use Case 4: Self-Hosted / Ollama

```toml
version = "1.0.0"

[scenario]
name = "Coffee Shop Encounter"
description = "Two strangers strike up a conversation, revealing contrasting worldviews"
tags = ["dialogue", "character-study", "discovery"]
location = "Busy coffee shop - Shared table"
time = "Morning rush"
atmosphere = "Noisy but cozy, forced proximity"

# Use local Ollama for 100% self-hosted execution
[scenario.defaults]
provider = "ollama"
model = "llama3.1:8b"

[agents.Emma]
character = "optimist"

[agents.Michael]
character = "cynic"
# Override to use a larger model for this agent
model = "llama3.1:70b"

[goals.find_common_ground]
description = "Find common ground despite different worldviews"
priority = 2
assignment = ["Emma", "Michael"]
type = "ConsensusGoal"
consensus_threshold = 0.7
tags = ["mutual_understanding", "worldviews"]
```

*(Character files `optimist.toml` and `cynic.toml` would define "The Optimist" and "The Cynic" archetypes)*

## Future Extensions

The scenario specification is designed for expansion:

1. **More goal types**: StateGoal, RescueGoal, ProximityGoal, ThresholdGoal (Phase 2)
2. **Conversational goal refinement**: LLM-based goal translation from natural language (Phase 2)
3. **RAG memory integration**: Long-term agent memory across simulations (Phase 2)
4. **Semantic distortion filters**: Character perception distortions like paranoia, trauma (Phase 3)
5. **Relationship templates**: Pre-existing relationships between characters
6. **Environmental events**: Triggers that occur during simulation
7. **Success criteria**: Measurable outcomes beyond goal completion
8. **Scenario inheritance**: Base scenarios that others extend
9. **Multi-scene scenarios**: Sequences of connected scenes

## Implementation Status

**Phase 1 MVP (Current)**:
- ✅ TOML-based scenario definition
- ✅ Character reference system (scenarios reference separate character files)
- ✅ Initial state overrides per scenario
- ✅ Character specifications with traits and personality
- ✅ Basic goal system (ConsensusGoal)
- ✅ Scene state definition
- ⏳ Scenario validation
- ⏳ Character loading and merging
- ⏳ CLI commands for loading and running scenarios

**Phase 2 (Future)**:
- ❌ Additional goal types (StateGoal, RescueGoal)
- ❌ Conversational goal refinement
- ❌ Action mode support
- ❌ RAG memory system
- ❌ Multiple LLM provider support

**Phase 3 (Future)**:
- ❌ Semantic distortion filters
- ❌ Advanced scene features
- ❌ Scenario library management
- ❌ Web UI for scenario editing

## References

- [Character Definition](./character-definition.md) - Complete character specification
- [Goal System](./goal-system.md) - Goal types and evaluation
- [Simulation Execution](./simulation-execution.md) - How scenarios are executed
- [MVP Roadmap](./mvp-roadmap.md) - Implementation phases
- [Overview](./overview.md) - System architecture
