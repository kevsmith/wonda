# Agent Architecture

## Overview
Agents are autonomous actors that embody characters during simulation execution. Each agent is a combination of:
- **LLM Provider**: The inference service (OpenAI API, Ollama, LM Studio, etc.)
- **Model & Parameters**: Specific model and generation settings (temperature, repeat penalty, etc.)
- **Character Assignment**: The character definition this agent embodies
- **State**: Current physical, emotional, cognitive, social, and goal state

## Agent Composition

### 1. Character Binding
- **Base template**: Character definition from scenario
- **Generated details**: Unique name, appearance variations
- **Personality parameters**: Traits that influence decision-making
- **Background knowledge**: Pre-existing information and relationships
- **Hidden elements**: Secrets and traits not immediately observable

### 2. State Management

#### Physical State
- **Position**: Current location within the scene
- **Condition**: Health/energy value (0-100)
- **Buffs/Debuffs**: Temporary conditions (e.g., ["stunned", "energized", "bleeding"])

#### Emotional State
- **Current emotion**: Primary emotion (angry, afraid, happy, sad, neutral)
- **Intensity**: Strength of emotion (0-10 scale)

#### Social State
- **Relationships**: Trust scores per character (-1 to 1)
- **Last interaction**: Recent exchange with each character

#### Cognitive State
- **Focus**: Current attention target (character, object, or task)
- **Certainty**: Confidence in understanding (0-1)

#### Goal State
- **Active goals**: Prioritized list (priority 1-5, with 1 highest)
- **Progress**: Completion percentage per goal

### 3. Memory Systems

Agents utilize a dual-memory architecture:
- **Short-term memory**: LLM context window for immediate situation
- **Long-term memory**: RAG-based retrieval system for persistent memories

See [RAG Memory System](./rag-memory-system.md) for detailed implementation.

### 4. Decision Engine

The agent decision process:
1. Gather perceptions via MCP tools
2. Retrieve relevant memories from RAG system
3. Apply character personality and state filters
4. Generate action decision via LLM
5. Execute through appropriate MCP tools

## Extension Points

### Semantic Distortion Filters
The architecture provides extension points for semantic distortions that can modify how agents perceive and process information:

#### Filter Integration Points
- **Perception filters**: Modify incoming sensory data
- **Emotional filters**: Adjust emotional responses
- **Cognitive filters**: Distort interpretation of information
- **Memory filters**: Affect memory formation and retrieval

#### Filter Architecture
```
Raw Input → [Distortion Filter] → Processed State → Decision Making
                     ↑
            [Configuration Parameters]
```

Filters are:
- Optional (most agents have none)
- Stackable (multiple conditions can apply)
- Configurable (intensity, triggers, duration)
- Context-aware (access to memories, state, environment)

Examples:
- Paranoia: Interprets neutral actions as threatening
- Trauma: Triggers emotional responses from specific stimuli
- Delusions: Overrides certain observations with false beliefs
- Intoxication: Reduces cognitive certainty and coordination

## LLM Integration

### Prompt Structure
Each agent turn generates a structured prompt containing:
1. Character definition and personality
2. Current state (all five categories)
3. Recent perceptions and observations
4. Retrieved memories (via RAG system)
5. Active goals and priorities
6. Available actions (MCP tools)

### Provider Abstraction
- OpenAI-compatible API interface
- Support for local (Ollama, LM Studio) and cloud providers
- Response caching for efficiency
- Fallback handling for failures

### Decision Output
Agent LLM calls produce:
- Selected action(s) with parameters
- Internal reasoning (logged, not shared)
- Memory formation candidates
- State updates (emotional shifts, focus changes)

## Behavioral Patterns

### Personality-Driven Actions
- Trait weights influence action probabilities
- Emotional state modifies thresholds
- Stress can override rational planning
- Character growth through experience

### Goal-Oriented Planning
- Multi-step planning for complex goals
- Dynamic replanning when blocked
- Priority balancing between competing goals
- Collaboration vs competition strategies

## State Transitions

### Update Triggers
States change through:
- **Actions**: Direct consequences (damage → reduced health)
- **Observations**: Environmental changes (betrayal → reduced trust)
- **Time**: Natural progression (anger cools, fatigue increases)
- **Thresholds**: Condition triggers (low health → desperation)

### State Influence
States affect behavior by:
- Filtering available actions (injured can't run)
- Modifying action selection (angry → aggressive)
- Changing perception accuracy (fear → misinterpretation)
- Adjusting goal priorities (wounded → survival focus)

## Time Awareness

### Hybrid Time Model
Agents maintain time awareness through a combination of contextual information and event-driven notifications.

#### Contextual Time Information
Every perception response includes temporal context:
```json
{
  "observation": "You see Alice arguing with Bob",
  "context": {
    "current_time": "3:15 PM",
    "time_since_last_action": "2 minutes",
    "scene_duration": "10 minutes",
    "day_phase": "afternoon"
  }
}
```

This ensures agents always know:
- Current simulation time
- How long since they last acted
- Overall scene duration
- Relevant temporal context (day/night, etc.)

#### Time-Triggered Notifications
Agents receive explicit time alerts for significant events:
- **Duration thresholds**: "You've been waiting for 10 minutes"
- **Deadline warnings**: "5 minutes until the meeting starts"
- **Phase transitions**: "The sun is setting"
- **Goal timeouts**: "Time running out to complete objective"

#### Time-Aware State
Time implicitly affects agent state:
- **Impatience**: Builds when goals are blocked
- **Fatigue**: Accumulates with activity and time
- **Urgency**: Increases as deadlines approach
- **Boredom**: Develops during inactivity

These states influence decision-making without requiring explicit time tracking.

#### Time Queries via MCP
Optional time-related tools for agents:
- `check_time()` - Get precise current time
- `estimate_duration(action)` - How long will action take
- `time_until(event)` - Check time to known events

This hybrid approach maintains time awareness without flooding the context with unnecessary updates.

## Memory Integration

Agents actively use memory through MCP tools:
- Query past experiences before decisions
- Recall relationship history with characters
- Find similar situations for guidance
- Build on learned information

Memory retrieval is encouraged through:
- Mandatory memory consultation in prompts
- Pre-populated relevant memories each turn
- Confidence gating requiring memory checks

See [RAG Memory System](./rag-memory-system.md) for implementation details.