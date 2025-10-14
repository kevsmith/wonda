# Scene State Management & Event System

## Overview
The scene state system maintains a single source of truth for the simulation environment, tracking locations, entities, relationships, and changes over time.

## Scene State Components

### 1. Spatial Model
Defines the physical layout of the scene:

**Locations**
- Named areas within the scene (e.g., "alley", "rooftop", "warehouse")
- Properties: description, connections, capacity, lighting
- Spatial relationships: adjacency, line-of-sight, distance
- Zones: Logical groupings (e.g., "public area", "secured zone")

**Spatial Queries**
- `get_visible_agents(location)` - Who can I see from here?
- `get_audible_range(location, volume)` - Who can hear this?
- `check_line_of_sight(location_a, location_b)` - Can they see each other?
- `get_path(from, to)` - How do I get there?

### 2. Entity Registry
Tracks all entities within the scene:

**Agents**
- Current location and facing direction
- Physical state (condition, buffs/debuffs)
- Inventory contents

**Objects**
- **Portable**: Can be picked up/moved (weapons, keys, documents)
- **Fixed**: Stationary features (doors, furniture, walls)
- **Interactive**: Actionable elements (buttons, levers, containers)
- Properties: state, location, owner, visibility

**Environmental Features**
- Non-interactive scene elements (weather, ambient sounds, lighting)

### 3. Environmental State
Overall scene conditions:

- **Lighting**: Per-location visibility levels
- **Atmosphere**: Mood, tension level, noise level
- **Weather/Conditions**: If relevant to scene (rain, fire, darkness)
- **Time State**: Current simulation time, phase of day

### 4. Change Log
Sequential record of all state changes:
- Used for goal evaluation and narrative generation
- Enables replay and debugging
- Pruned after simulation completion

## Event System

### Event Types

- **Action events**: Agent performs action
- **State change events**: Property values change
- **Perception events**: Something becomes observable
- **Goal events**: Goal progress/completion changes
- **Mode events**: Narrative ↔ Action mode transitions

### Event Structure

```json
{
  "id": "evt_1234",
  "type": "action",
  "timestamp": 1234,
  "turn_number": 5,
  "actor": "hero_agent",
  "action": "speak",
  "target": null,
  "data": {"message": "Drop your weapon!", "volume": "shout"},
  "observable_by": ["villain_agent", "civilian_agent"],
  "state_changes": [
    {
      "entity": "tension",
      "property": "level",
      "old_value": 0.6,
      "new_value": 0.9
    }
  ]
}
```

### Event Propagation Flow

1. Agent submits action via MCP tool
2. Environment validates action against current state
3. Event created with effects
4. State changes applied atomically
5. Observable events sent to relevant agents
6. Goal monitors evaluate new state
7. Event logged for narrative generation

## Transaction Model

### Turn-Based Sequential Processing

The simulation uses strict turn order with sequential action processing:

1. **Agent turn begins** with current authoritative state
2. **Perception phase**: Agent queries state via MCP servers
3. **Decision phase**: Agent LLM processes inputs and selects action
4. **Submission**: Action sent to environment
5. **Validation**: Environment checks action validity:
   - Required resources available?
   - Target still in valid state?
   - Physical constraints satisfied?
6. **Commit or Reject**:
   - **Valid**: Apply state changes atomically, generate event
   - **Invalid**: Return error with reason, agent may retry (within turn limits)
7. **State updated**: Next agent receives updated state
8. **Next turn**: Process repeats for next agent in initiative order

### No Simultaneous Conflicts

Since only one agent acts at a time:
- State is consistent at the start of each turn
- No race conditions or simultaneous modifications
- No complex conflict resolution needed
- Simple validation suffices

### Validation Edge Cases

Validation failures occur when state changes between perception and action:

**Example**:
```
1. Agent B perceives: "Door is closed"
2. Agent A's turn: Opens the door
3. Agent B's turn: Attempts to open door
4. Validation fails: "Door already open"
5. Agent B informed, chooses different action
```

This is normal validation, not conflict resolution.

## Future Optimization: Concurrent Preparation

### Parallel Pre-computation

To reduce wall-clock simulation time, agents can prepare for upcoming turns while others act:

**Optimized Flow**:
1. **Agent A's turn executes** (state = S1)
2. **Agent B starts preparing** (next in initiative):
   - Takes state snapshot (S1)
   - Queries state via MCP from snapshot
   - Begins LLM inference asynchronously
3. **Agent A completes**, state updates (S1 → S2)
4. **Agent B's turn begins**:
   - Validate prepared action against current state (S2)
   - If valid → Execute immediately (fast!)
   - If invalid → Quick retry with updated state

### Expected Success Rate

Most prepared actions will validate successfully because:
- Agents often act on different parts of the scene
- Spatial separation reduces dependencies
- Many actions don't interfere with each other

When validation fails:
- Agent receives updated state delta
- Fast re-decision with new information
- Or fallback to synchronous processing

### Benefits
- Reduced wall-clock time for simulations
- Better LLM provider utilization (parallel calls)
- Initiative order still strictly preserved
- State consistency maintained
- No changes to core model required

### Implementation Considerations
- State snapshots must be immutable
- Validation logic must be fast
- LLM calls must be cancellable (if agent removed/incapacitated)
- May prepare N agents ahead based on system capacity

## State Consistency Guarantees

### Single Source of Truth
- One authoritative state representation
- All queries read from this state
- All modifications go through event system
- No distributed or cached state

### Atomic Updates
- State changes are all-or-nothing
- Either entire action succeeds or none of it applies
- No partial state corruption possible

### Deterministic Replay
- Given same initial state and actions, produces same results
- Change log enables full replay
- Useful for debugging and narrative review

## Integration with MCP Servers

MCP servers act as the interface between agents and state:

**Perception Server**
- Queries entity registry for visible/audible entities
- Uses spatial model for range and line-of-sight
- Returns filtered state based on agent capabilities

**Action Server**
- Receives action requests from agents
- Validates against current state
- Generates events and applies state changes
- Returns results and observable effects

**Cognition Server**
- Queries relationship data from state
- Accesses knowledge graph
- Evaluates situation based on state

**Communication Server**
- Uses spatial model for message propagation
- Determines who can hear/receive messages
- Creates communication events in change log

## State Initialization

Scenarios define initial state:
```toml
[[initial_state.locations]]
name = "bank_lobby"
description = "Marble floors, high ceiling, three teller windows"
connections = ["vault_door", "front_entrance", "manager_office"]
lighting = "bright"

[[initial_state.objects]]
id = "vault_door"
type = "fixed"
state = "locked"
location = "bank_lobby"

[initial_state.environment]
time = "2:30 PM"
tension = 0.2
public_awareness = "none"
```

Agents are placed into this initial state when simulation begins.