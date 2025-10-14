# Simulation Execution

## Overview
This document describes how simulations are executed, including turn management, agent decision flow, and action resolution.

## Execution Modes

Simulations operate in two distinct modes that determine pacing and action resolution:

### Narrative Mode (Default)
- **Time**: Turns represent narrative beats (flexible duration)
- **Focus**: Character interaction, exploration, planning, dialogue
- **Time Tracking**: Loose descriptions ("a few minutes later", "as evening approaches")
- **Actions**: No strict action economy - agents act naturally within reason
- **Turn Order**: Flexible based on narrative flow

### Action Mode
- **Time**: Each turn = exactly 30 seconds
- **Focus**: Combat, crisis resolution, time-critical goals
- **Time Tracking**: Precise countdown ("3:30 remaining")
- **Actions**: Strict economy - 1 major action OR 2 minor actions per turn
- **Turn Order**: Fixed initiative order

### Mode Transitions

**Triggers for Action Mode:**
- Combat initiated by any agent
- Goal with deadline activated (e.g., "defuse bomb in 5 minutes")
- Crisis event triggered (e.g., "building collapsing")
- Scenario-specific requirements

**Return to Narrative Mode:**
- All combat/crisis resolved
- Timed goals completed or failed
- All agents disengage/stand down
- Scene explicitly calls for transition

**Transition Example:**
```
Narrative Mode: "You're negotiating with the guard about passage..."
[Guard draws weapon]
⚡ ENTERING ACTION MODE - Roll Initiative ⚡
Action Mode: "30-second turns. Guard acts first..."
```

## Turn-Based Execution

### Initiative System (Action Mode Only)
- Determined once when entering Action Mode
- Based on character attributes, scenario rules, or dice rolls
- Order remains stable throughout the action sequence
- Re-rolled when returning to Action Mode after Narrative Mode

### Turn Structure
Each agent's turn follows this sequence:

1. **State Update**: Agent receives current state information
2. **Perception**: Agent uses MCP tools to gather information
3. **Memory Retrieval**: RAG system provides relevant memories
4. **Decision**: LLM processes all inputs to select action(s)
5. **Action Execution**: Selected actions are validated and applied
6. **World Update**: Environment state changes, consequences propagate
7. **Notification**: Other agents are notified of observable changes

### Action Economy (Action Mode)

**Free Actions** (unlimited within reason):
- Speaking/listening
- Observing immediate surroundings
- Simple gestures
- Internal decision-making

**Minor Actions** (max 2 per turn):
- Move short distance
- Pick up/drop item
- Open/close door
- Quick search
- Simple object manipulation

**Major Actions** (max 1 per turn):
- Attack/combat maneuver
- Sprint across area
- Complex manipulation
- Thorough search
- Athletic feat
- Extended task progress

## Agent Decision Flow

### Input Gathering Phase
1. Query environment state via MCP Perception Server
2. Retrieve relevant memories from RAG system
3. Check current goal progress
4. Assess character state (physical, emotional, social)

### Processing Phase
The LLM receives structured input:
- Character definition and personality
- Current observations and state
- Retrieved memories and knowledge
- Active goals with priorities
- Available actions and constraints

### Output Phase
Agent produces:
- Chosen action(s) with parameters
- Internal reasoning (logged but not shared)
- Memory formations (what to remember)
- Emotional state updates

## Action Resolution

### Validation
Before execution, actions are validated for:
- Physical possibility given current state
- Character capability and knowledge
- Resource availability
- Environmental constraints

### Conflict Resolution
When actions conflict:
1. Priority based on initiative order
2. Simultaneous effects when non-conflicting
3. Partial success for competing actions
4. Environmental adjudication for complex interactions

### Consequence Propagation
1. Direct effects applied immediately
2. Secondary effects cascade through environment
3. Other agents notified of observable changes
4. Goal monitors check for completion
5. State logged for narrative output

## Timing and Pacing

### Turn Timing
- Configurable time limit per turn (for LLM response)
- Timeout handling with default/fallback actions
- Async LLM calls with queue management

### Narrative Pacing
- Action sequences: Quick turns, high stakes
- Free play: Longer turns, more complex actions
- Dramatic moments: Possible time dilation for detail
- Montage mode: Compressed time for routine activities

## Termination Conditions

Simulations end when:
1. **Goal Completion**: Sufficient goals achieved (based on priority)
2. **Time Limit**: User-specified duration reached
3. **Stalemate**: No progress possible toward remaining goals
4. **Manual Termination**: User intervention
5. **Catastrophic State**: All agents incapacitated

## Logging and Output

### For Writers
- Narrative description of each significant action
- Character internal monologues (optional)
- Scene descriptions at key moments
- Goal progress indicators
- Dramatic tension metrics

### For Debugging
- Full agent decision traces
- MCP tool calls and responses
- State change log
- Memory formation/retrieval events
- Performance metrics