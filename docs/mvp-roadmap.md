# MVP Roadmap

## Overview

This document outlines the phased implementation approach for the multi-agent simulation system, starting with a Minimum Viable Product (MVP) and progressing toward the full vision.

## Design Philosophy

**Start Simple, Iterate to Excellence**

The MVP focuses on core functionality that proves the concept and provides value to early users, while maintaining a clear path to the complete system described in the design documents. Each phase builds on the previous one without requiring major refactoring.

## Phase 1: MVP - Core Simulation Engine

### Goal
Run a simulation with two agents sharing a common goal, demonstrating that the system can generate meaningful character interaction chronicles for fiction writing inspiration.

### Scope

#### Included in MVP

**1. Scenario Definition (Structured Format)**
- TOML-based scenario files
- Character definitions with personality traits
- Goal definitions with structured parameters (not conversational)
- Initial scene state

**2. Agent System**
- Character-to-agent instantiation
- Single LLM provider support (Ollama for local development)
- Basic state tracking (position, emotion, condition)
- LLM integration for decision-making

**3. MCP Server Layer**
- Perception Server: `observe_scene()`, `listen()`
- Communication Server: `speak()`
- Action Server: `move()`, `manipulate()`
- Cognition Server: `assess_goal()`

**4. Goal System - Structured Approach**
- Writers define goals with explicit parameters
- Pre-defined goal types (ConsensusGoal, StateGoal, RescueGoal)
- Structured parameter editing interface
- Clear success condition display

**5. Simulation Execution**
- Turn-based execution with initiative
- Narrative mode only (no Action mode)
- Sequential turn processing
- Basic event logging

**6. Chronicle Generation**
- Observable events layer
- Internal state layer
- Basic dialogue export
- JSON chronicle output

**7. Memory System (Simplified)**
- Short-term memory only (LLM context)
- No RAG/vector storage in MVP
- Simple episodic logging

#### Explicitly Out of Scope for MVP

- ❌ Conversational goal refinement (Phase 2)
- ❌ LLM-based goal translation (Phase 2)
- ❌ RAG-based long-term memory (Phase 2)
- ❌ Action mode with 30-second turns (Phase 2)
- ❌ Semantic distortion filters (Phase 3)
- ❌ Multiple LLM providers (Phase 2)
- ❌ Concurrent agent preparation optimization (Phase 3)
- ❌ Visual UI (Phase 3)
- ❌ Multi-simulation management (Phase 3)

### MVP Goal System Design

Since conversational goal refinement is deferred to Phase 2, the MVP uses a **structured parameter approach** with friendly display:

#### Writer Experience

**Goal Input (TOML):**
```toml
[[goals]]
id = "goal_001"
description = "Save the children from the evil monster"
priority = 1
assignment = ["Hero", "Guardian"]
type = "RescueGoal"

[goals.parameters]
targets = ["children"]
threat = "evil_monster"
rescue_threshold = 0.8  # 80% of children
threat_resolution = "distance_only"
distance_threshold = 50
```

**Validation Feedback (CLI):**
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
GOAL: Save the children from the evil monster
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Type: Rescue Mission

Who's responsible:
  • Hero
  • Guardian

What needs to happen:
  • At least 80% of children must reach safety
  • Evil monster must be at least 50 meters away

How completion is tracked:
  ✓ Automatic - System checks positions after each turn
  ✓ No character reporting needed

Current parameters:
  • Rescue threshold: 80% (4 out of 5 children)
  • Distance threshold: 50 meters
  • Threat handling: Distance only (no defeat required)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Status: ✓ Valid goal definition
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

#### Goal Type Reference

MVP includes documentation for each goal type:

```markdown
# RescueGoal

Protect or save entities from threats.

## Parameters

- `targets` (array): Entity IDs or tags to rescue
- `threat` (string): Entity ID of the threat
- `rescue_threshold` (float): 0.0-1.0, percentage that must be saved
- `threat_resolution` (string): "defeat", "distance_only", "defeat_or_distance"
- `distance_threshold` (int): Meters required if using distance

## Example

```toml
type = "RescueGoal"

[parameters]
targets = ["civilian_1", "civilian_2", "civilian_3"]
threat = "villain"
rescue_threshold = 1.0
threat_resolution = "defeat_or_distance"
distance_threshold = 100
```

This means all civilians must be saved AND villain defeated or 100+ meters away.
```

Writers reference this documentation when creating scenarios.

### Success Criteria for MVP

The MVP is successful if:

1. ✅ A writer can create a scenario TOML file
2. ✅ Run a simulation with 2 agents sharing a common goal
3. ✅ Agents interact naturally via LLM decision-making
4. ✅ Goal completion is detected accurately
5. ✅ A chronicle is generated with dialogue and internal thoughts
6. ✅ Writer can export chronicle as JSON or formatted text

### Stretch Goal: Semantic Distortion Filters

If time and resources permit, implement basic semantic distortion mechanism for agent perception and memory:

**Scope:**
- Extension point architecture for distortion filters
- One reference implementation (e.g., paranoia filter)
- Distortion affects perception interpretation
- Documented API for adding custom filters

**Success Criteria:**
- ✅ Filter can modify perception data before agent processes it
- ✅ Example scenario demonstrates paranoid character misinterpreting neutral actions
- ✅ Chronicle captures both original and distorted perceptions

This stretch goal validates the extension point architecture for Phase 3 psychological effects.

### MVP Architecture

```
┌─────────────────────────────────────────────┐
│         Scenario TOML File                  │
│  (Characters, Goals, Scene State)           │
└──────────────────┬──────────────────────────┘
                   │
                   ↓
┌─────────────────────────────────────────────┐
│      Simulation Framework                   │
│  • Load scenario                            │
│  • Initialize agents                        │
│  • Execute turns (Narrative mode only)      │
│  • Monitor goals                            │
│  • Generate chronicle                       │
└──────────┬──────────────────────────────────┘
           │
           ├──→ Agent Manager
           │    • Agent state
           │    • LLM calls (Ollama)
           │    • MCP tool access
           │
           ├──→ MCP Servers
           │    • Perception
           │    • Communication
           │    • Action
           │    • Cognition
           │
           ├──→ Goal Monitor
           │    • StateGoal evaluator
           │    • ConsensusGoal evaluator
           │    • RescueGoal evaluator
           │
           └──→ Chronicle Builder
                • Event logging
                • State snapshots
                • Export formats
```

### Technology Decisions for MVP

- **Language:** Go or Python (choose based on team preference)
- **LLM Provider:** Ollama (local, free, easy setup)
- **Default Model:** llama3.1:8b or similar
- **State Storage:** In-memory (no persistence in MVP)
- **Chronicle Export:** JSON files
- **Interface:** CLI (command-line interface)

### MVP Deliverables

1. **Core System**
   - Scenario loader
   - Simulation engine
   - Agent manager
   - MCP server implementations
   - Goal monitoring system
   - Chronicle generator

2. **Documentation**
   - Goal type reference guide
   - Scenario TOML format guide
   - Character definition guide
   - Getting started tutorial
   - Example scenarios (3-5)

3. **Examples**
   - "Dinner Planning" (consensus)
   - "Monster Attack" (rescue)
   - "Negotiation" (multi-goal)

4. **CLI Tools**
   - `wonda run scenario.yaml` - Execute simulation
   - `wonda validate scenario.yaml` - Check scenario validity
   - `wonda export chronicle.json --format=dialogue` - Export chronicle

### Estimated Effort: 4-6 Weeks

**Week 1-2: Core Infrastructure**
- Scenario loading
- Agent instantiation
- Basic turn execution
- LLM integration

**Week 3-4: MCP & Goals**
- MCP server implementations
- Goal evaluators (3-4 types)
- State management
- Event system

**Week 5: Chronicle**
- Event logging
- Chronicle building
- Export formats

**Week 6: Polish**
- Documentation
- Example scenarios
- Bug fixes
- CLI refinement

---

## Phase 2: Enhanced Usability

### Goal
Make the system accessible to non-technical writers through conversational interfaces and improved automation.

### Key Features

**1. Conversational Goal Refinement**
- LLM-based goal translation
- Natural language input
- Iterative refinement loop
- Automatic goal type selection

**2. RAG Memory System**
- Vector database integration
- Long-term memory for agents
- Memory retrieval via MCP
- Importance-based retention

**3. Action Mode**
- 30-second turn timing
- Strict action economy
- Mode switching (Narrative ↔ Action)
- Initiative system

**4. Multiple LLM Providers**
- OpenAI API support
- Anthropic Claude support
- Provider configuration per agent
- Fallback handling

**5. Scenario Builder Interface**
- Interactive CLI wizard
- Guided character creation
- Conversational goal definition
- Scenario testing

### Success Criteria

1. ✅ Non-technical writer can define scenarios without editing YAML
2. ✅ Goals expressed in natural language are correctly interpreted
3. ✅ Agents maintain persistent memory across long simulations
4. ✅ Action mode provides tactical time-pressure scenarios
5. ✅ Multiple LLM providers work interchangeably

### Estimated Effort: 6-8 Weeks

---

## Phase 3: Production Polish

### Goal
Production-ready system with performance optimization, rich features, and great user experience.

### Key Features

**1. Semantic Distortion Filters**
- Paranoia, trauma, delusion effects
- Configurable per-character
- Memory and perception distortion

**2. Concurrent Agent Preparation**
- Parallel LLM calls
- Validation-based conflict resolution
- Significant speedup for multi-agent scenarios

**3. Web UI**
- Visual scenario builder
- Real-time simulation viewer
- Interactive chronicle explorer
- Goal refinement interface

**4. Chronicle Query System**
- Semantic search
- Theme detection
- Character journey analysis
- Export customization

**5. Performance Optimization**
- State caching
- Incremental chronicle updates
- Optimized goal evaluation
- Memory management

**6. Multi-Simulation Management**
- Save/load simulations
- Compare simulation variants
- Simulation library
- Version control integration

### Success Criteria

1. ✅ Simulations run 3x+ faster with concurrent preparation
2. ✅ Web UI accessible to completely non-technical users
3. ✅ Chronicle query finds relevant moments in seconds
4. ✅ Semantic distortions create compelling character effects
5. ✅ System handles 10+ agent scenarios smoothly

### Estimated Effort: 10-12 Weeks

---

## Decision Points

### After MVP
**Question:** Is the core simulation valuable? Do chronicles inspire writers?

**If YES:** Proceed to Phase 2
**If NO:** Revisit core assumptions about goal monitoring, chronicle content, or agent behavior

### After Phase 2
**Question:** Are non-technical writers adopting the system? Is conversational goal definition working? Can they avoid editing TOML files directly?

**If YES:** Proceed to Phase 3
**If NO:** Iterate on Phase 2 features, particularly goal refinement UX

### After Phase 3
**Question:** Is this production-ready? Are users creating and sharing scenarios?

**If YES:** Launch publicly, build community
**If NO:** Address blockers before launch

---

## Risk Mitigation

### Risk: LLM Unpredictability
**Mitigation:**
- Extensive prompt engineering
- Validation layers
- Fallback to structured input if LLM fails
- Testing with multiple models

### Risk: Goal Evaluation Complexity
**Mitigation:**
- Start with simple goal types in MVP
- Add complexity gradually in Phase 2
- Provide clear examples and documentation
- Allow manual goal progress overrides

### Risk: Performance with Many Agents
**Mitigation:**
- MVP limited to 2-3 agents
- Phase 2 adds optimization
- Phase 3 adds concurrent preparation
- Set clear agent count limits per phase

### Risk: Writer Learning Curve
**Mitigation:**
- Excellent example scenarios
- Step-by-step tutorials
- Clear error messages
- Gradual feature introduction across phases

---

## Success Metrics

### MVP Metrics
- Can run 3+ example scenarios successfully
- Chronicle contains meaningful character interactions
- Goal completion detected accurately
- Writer feedback: "I can see how this would inspire me"

### Phase 2 Metrics
- 80%+ of natural language goals translate correctly
- Writers create scenarios without touching YAML
- Simulation length increases 3x+ with RAG memory
- Action mode creates compelling tension

### Phase 3 Metrics
- Web UI onboarding takes <10 minutes
- Writers share scenarios with community
- Chronicle queries find relevant content in <2s
- Users run 5+ simulations per week

---

## Proposed Implementation Plan

### MVP Scope Reduction

To achieve the MVP goal ("Run a simulation with two agents sharing a common goal") with minimal implementation effort, we deliberately **exclude** several features from the full design:

**Excluded from MVP:**
- ❌ Spatial system (locations, positions, movement, line-of-sight)
- ❌ Multiple goal types (implement only ConsensusGoal)
- ❌ Action Server (no physical actions/manipulation)
- ❌ Complex state tracking (simplified agent state)
- ❌ Multiple scenarios (single reference scenario included)
- ❌ Scene environment state (time, atmosphere, etc.)
- ❌ Detailed validation and error handling
- ❌ CLI commands beyond basic run/export

**What This Means:**
- Agents exist in an abstract conversational space
- They can speak and listen but not move or manipulate objects
- One goal type handles the "agree on restaurant" use case
- Minimal state: just emotion and goal tracking
- Focus is on proving dialogue + goal completion works

### Core Components (Minimal)

#### 1. Scenario Loader
**Input:** TOML file with:
- 2 characters (name, traits, communication_style, decision_style)
- 1 goal (ConsensusGoal, assigned to both)

**Output:** Parsed scenario structure

**Effort:** 2-3 days
- TOML parsing library integration
- Basic validation (required fields present)
- Load character and goal definitions

#### 2. Agent System
**Responsibilities:**
- Instantiate agents from characters
- Maintain minimal state (emotion, goal progress opinion)
- Build LLM prompts with character context
- Parse LLM responses

**Effort:** 3-4 days
- Agent struct/class with state
- LLM integration (Ollama client)
- Prompt template system
- Response parsing

#### 3. MCP Servers (Minimal Set)

**Communication Server:**
- `speak(message)` - Agent says something
- Returns: confirmation

**Perception Server:**
- `listen()` - Get recent dialogue from other agent
- Returns: last message(s) from other agent

**Cognition Server:**
- `assess_goal(goal_id, completion, reasoning)` - Report goal progress
- Returns: confirmation

**Effort:** 2-3 days
- Three simple tool implementations
- MCP protocol basics (tool calling interface)
- Response formatting

#### 4. Simulation Engine
**Responsibilities:**
- Initialize agents
- Execute turns in sequence (initiative = random or fixed)
- Route MCP tool calls to appropriate servers
- Collect events for chronicle

**Effort:** 3-4 days
- Turn loop implementation
- Agent turn execution
- Tool dispatch
- Event collection

#### 5. Goal Monitor (ConsensusGoal Only)
**Responsibilities:**
- Track goal assessments from both agents
- Detect completion when both report 1.0
- Trigger simulation end

**Effort:** 1-2 days
- Single evaluator implementation
- Completion detection logic
- Termination trigger

#### 6. Chronicle Generator
**Layers:**
1. Observable events (dialogue)
2. Internal state (emotions)
3. Internal monologue (LLM reasoning)

**Exports:**
- JSON (full chronicle)
- Formatted text (readable narrative)

**Effort:** 2-3 days
- Event logging during simulation
- Chronicle structure
- Two export formats

### Implementation Order

**Phase 1: Foundation (Week 1)**
1. Scenario loader + TOML parsing
2. Agent structure + basic LLM integration
3. Simple turn loop (agents say hello to each other)

**Phase 2: Core Interaction (Week 2)**
1. MCP Communication Server (speak)
2. MCP Perception Server (listen)
3. Agents can have back-and-forth dialogue

**Phase 3: Goals (Week 3)**
1. MCP Cognition Server (assess_goal)
2. ConsensusGoal evaluator
3. Goal completion detection

**Phase 4: Chronicle (Week 4)**
1. Event logging
2. Chronicle structure
3. Export formats
4. End-to-end test with "Dinner Plans" scenario

**Phase 5: Polish & Documentation (Week 5)**
1. CLI refinement
2. Example scenario
3. Getting started guide
4. Bug fixes

### Minimal MVP Deliverables

**Code:**
- Scenario loader
- Agent system with LLM integration
- 3 MCP servers (minimal)
- Simulation engine
- ConsensusGoal evaluator
- Chronicle generator
- CLI runner

**Documentation:**
- TOML scenario format guide
- Character definition guide
- ConsensusGoal explanation
- Quick start tutorial

**Examples:**
- "Dinner Plans" scenario (Alex & Jordan)
- Sample chronicle output

**CLI:**
```bash
wonda run scenario.toml            # Run simulation
wonda export chronicle.json        # Export as readable text
wonda export chronicle.json --json # Export as JSON
```

### Total Effort Estimate: 4-5 Weeks

- Week 1: Foundation (scenario loading, agent basics)
- Week 2: Dialogue interaction (MCP servers)
- Week 3: Goals (assessment and completion)
- Week 4: Chronicle (logging and export)
- Week 5: Polish and documentation

### Success Criteria (Minimal)

The MVP succeeds if:

1. ✅ Load "Dinner Plans" scenario from TOML
2. ✅ Two agents (Alex & Jordan) have 3-4 turn dialogue exchange
3. ✅ Both agents assess goal completion
4. ✅ System detects completion and ends simulation
5. ✅ Chronicle exported showing dialogue + internal thoughts
6. ✅ Writer can read chronicle and understand what happened

---

## Addendum: Stretch Goal Implementation

If the core MVP is completed ahead of schedule, add semantic distortion filters.

### Additional Components

#### 1. Distortion Filter Architecture
**Responsibilities:**
- Extension point for modifying perception data
- Apply filters before agent processes perception
- Log both original and distorted versions

**Effort:** 2-3 days
- Filter interface/base class
- Filter registration system
- Pre-processing pipeline
- Chronicle integration

#### 2. Paranoia Filter (Reference Implementation)
**Behavior:**
- Interprets neutral statements as suspicious
- Adds negative emotional valence
- Questions motives

**Example:**
```
Original: "What about that Italian place?"
Distorted: "What about that Italian place? [They're trying to manipulate me into going somewhere cheap]"
```

**Effort:** 2-3 days
- LLM-based distortion (prompt that reinterprets)
- OR pattern-based rules (simpler)
- Configuration (distortion intensity)

#### 3. Example Scenario with Distortion
**"Suspicious Friends":**
- Alex (normal perception)
- Jordan (paranoid perception)
- Goal: Agree on restaurant
- Demonstrates misunderstanding due to distorted perception

**Effort:** 1 day
- Scenario creation
- Testing
- Documentation

### Stretch Goal Timeline

**Week 5 (if ahead) or Week 6:**
- Days 1-2: Filter architecture
- Days 3-4: Paranoia filter implementation
- Day 5: Example scenario + documentation

### Stretch Goal Success Criteria

1. ✅ Filter can intercept perception before agent processes it
2. ✅ Paranoia filter distorts at least one perception in example
3. ✅ Chronicle shows both original and distorted versions
4. ✅ Documentation explains how to add custom filters

---

## Technology Stack (Recommendations)

### Language: Python
**Rationale:**
- Faster prototyping
- Excellent LLM library ecosystem (openai, anthropic, ollama clients)
- TOML support (tomli/tomllib)
- Rich CLI libraries (click, rich)
- Easy for writers to understand if they peek at code

### Key Libraries:
- `tomllib` (Python 3.11+) or `tomli` - TOML parsing
- `ollama` - LLM client
- `pydantic` - Data validation
- `click` - CLI framework
- `rich` - Pretty terminal output
- `pytest` - Testing

### Alternative: Go
**If team prefers:**
- Better performance
- Static typing
- Single binary distribution
- Libraries: `pelletier/go-toml`, `tmc/langchaingo`

### Project Structure (Python)
```
wonda/
├── src/
│   ├── scenario/      # TOML loading
│   ├── agent/         # Agent system
│   ├── mcp/           # MCP servers
│   ├── simulation/    # Simulation engine
│   ├── goals/         # Goal evaluators
│   ├── chronicle/     # Chronicle generation
│   └── cli/           # CLI commands
├── examples/
│   └── dinner-plans.toml
├── tests/
├── docs/
└── pyproject.toml
```

## Open Questions

These will be decided during implementation:

1. **LLM response format:** Structured JSON or parse from text?
2. **Turn limit:** Max turns before timeout (default 20)?
3. **Chronicle verbosity:** How much LLM reasoning to capture?
4. **Error recovery:** What if LLM returns garbage?
5. **Testing strategy:** Focus on integration tests with example scenarios?