# MCP Server Interface Design

## Overview
The simulation environment provides multiple specialized MCP servers that enable agents to perceive and interact with the simulated scene.

## MCP Server Architecture

### 1. Perception Server
Provides sensory input for agents to understand their environment.

#### Tools
- `observe_scene(radius?, focus?)` - Returns visual description of surroundings
  - Includes: visible characters, objects, environmental details
  - Filtered by line-of-sight and distance
  - Optional focus parameter for detailed examination

- `listen(radius?)` - Returns audible information
  - Conversations, ambient sounds, distant noises
  - Volume/clarity degrades with distance

- `sense_atmosphere()` - Returns emotional/social context
  - Tension levels, crowd mood, general ambiance
  - Subtle social cues that a character might pick up

### 2. Action Server
Handles physical interactions within the scene.

#### Tools
- `move(destination, speed?)` - Movement within scene
  - Validates path availability and obstacles
  - Triggers proximity events for other agents
  - Updates agent visibility/audibility to others

- `manipulate(object, action)` - Object interaction
  - Pick up, use, throw, open, close, etc.
  - Validates physical possibility and permissions
  - Updates world state and triggers consequences

- `gesture(type, target?)` - Non-verbal communication
  - Wave, point, threaten, comfort, etc.
  - Visible to agents within line of sight

### 3. Communication Server
Manages verbal and written interactions between agents.

#### Tools
- `speak(message, volume?, target?)` - Verbal communication
  - Volume levels: whisper, normal, shout
  - Can be directed or broadcast
  - Automatically heard by agents in range

- `write(message, medium)` - Create written content
  - Notes, signs, messages
  - Persists in environment for others to read

### 4. Cognition Server
Provides knowledge and reasoning support based on character context.

#### Tools
- `recall_relationship(character_name)` - Query relationship status
  - History of interactions, trust level, emotional valence

- `assess_situation()` - Get strategic overview
  - Progress toward goals, threats, opportunities
  - Filtered through character's perception abilities

- `check_knowledge(topic)` - Verify what character knows
  - Both character background and learned information

## MCP Response Patterns

Each MCP tool returns structured responses that include:
- **Primary data**: The requested information
- **Confidence level**: How certain/complete the information is
- **Side effects**: Any state changes triggered
- **Visibility flags**: What other agents might observe

### Example Response Structure
```json
{
  "data": {
    "description": "You see a dimly lit alley with a figure lurking in shadows",
    "details": {
      "lighting": "poor",
      "exits": ["north", "south"],
      "characters": ["shadowy figure"],
      "objects": ["dumpster", "fire escape ladder"]
    }
  },
  "confidence": 0.7,
  "side_effects": ["shadowy figure notices you looking"],
  "observable_by": ["shadowy figure"]
}
```

## Action Economy

Actions have different costs within the turn structure:
- **Major actions**: Combat, complex manipulations (consume turn)
- **Minor actions**: Observation, speech (may be free or limited)
- **Extended actions**: Multi-turn tasks with progress tracking

## Event Propagation

Actions generate events that ripple through the simulation:
1. Agent performs action via MCP tool
2. Environment validates and applies action
3. State changes propagate to affected entities
4. Other agents receive relevant perception updates
5. Goal monitors check for completion conditions

## Error Handling

MCP tools include robust error handling:
- Invalid actions return clear explanations
- Partial success states for complex actions
- Graceful degradation when information is incomplete
- Character-appropriate filtering of error messages