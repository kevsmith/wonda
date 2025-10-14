# Goal Monitoring & Completion System

## Overview
Goals define the success conditions for a simulation. They are continuously monitored and drive both agent behavior and simulation termination.

## Goal Structure

Each goal contains:
- **Description**: Natural language description of what must be accomplished
- **Priority**: 1-5 (1 = highest priority)
- **Assignment**: List of characters who have this goal (e.g., ["hero"], ["hero", "sidekick"], or ["all"])
- **Evaluation Function**: Returns completion percentage (0.0 to 1.0)
- **Deadline** (optional): Time limit that triggers Action Mode when active
- **Completion Threshold** (optional): Minimum evaluation score for success (default 1.0)

## Goal Evaluation

### Unified Evaluation Model
All goals use a single evaluation interface that returns a value from 0.0 (not started) to 1.0 (fully complete).

### Examples

**Binary Goal**: "Villain defeated"
```
evaluation: () => villain.status === 'defeated' ? 1.0 : 0.0
```

**Progress Goal**: "Evacuate all civilians"
```
evaluation: () => evacuated_count / total_civilian_count
// 3 of 5 evacuated returns 0.6
```

**Proximity Goal**: "Reach the safehouse"
```
evaluation: () => 1.0 - (current_distance / starting_distance)
// Gets closer to 1.0 as agent approaches
```

**Threshold Goal**: "Maintain trust level" (threshold: 0.7)
```
evaluation: () => average_trust_level
// Complete when trust >= 0.7
```

**Multi-condition Goal**: "Secure the perimeter"
```
evaluation: () => (doors_locked + windows_closed + guards_posted) / 3
// Each condition contributes 0.33 to completion
```

## Design Limits

To prevent computational bottlenecks:
- **Maximum goals per scenario**: 8 total
- **Assignment**: Flexible (each goal can be assigned to any combination of characters)

This limit ensures goal evaluation remains performant while providing sufficient complexity for rich narratives. Eight goals can create plenty of dramatic tension without overwhelming the system.

## Goal Monitoring Architecture

### Evaluation Pipeline
1. Agent performs action
2. World state updates
3. Goal monitor identifies affected goals
4. Evaluates only relevant goals
5. Updates progress tracking
6. Checks for completion/failure
7. Triggers mode changes or termination

### Optimization Strategies

**Lazy Evaluation**
- Only evaluate goals that could be affected by the recent action
- Tag goals with relevant triggers (location, characters, objects)

**Priority Ordering**
- Evaluate high-priority goals first
- May skip low-priority goals if termination triggered

**Early Termination**
- Stop evaluating once enough goals are complete for scenario success

**Caching**
- Cache evaluation results that don't change frequently
- Invalidate cache when relevant state changes

## Mode Interaction

Goals work identically in both Narrative and Action modes. The presence of a deadline is what triggers mode transitions:

**Goal without deadline**: "Convince the mayor"
- Operates in current mode (usually Narrative)
- No time pressure

**Goal with deadline**: "Defuse bomb in 5 minutes"
- Triggers Action Mode when activated
- Precise 30-second turns
- Returns to Narrative Mode when resolved

## Completion Conditions

### Scenario Success
Scenarios complete successfully when:
- All priority 1 goals are complete
- OR sufficient total priority points achieved
- OR specific combination defined by scenario

### Scenario Failure
Scenarios fail when:
- Any priority 1 goal becomes impossible
- OR deadline expires on critical goal
- OR all agents incapacitated

### Partial Success
Scenarios can define partial success thresholds:
- Complete 60% of goals by priority weight
- Complete any 2 of 3 priority 1 goals
- Achieve specific goal combinations

## Agent Goal Awareness

When agents are instantiated from characters, they inherit that character's assigned goals. Agents receive this goal information in their context:

```json
{
  "assigned_goals": [
    {
      "description": "Protect your partner",
      "priority": 1,
      "progress": 1.0,
      "deadline": null,
      "assigned_to": ["hero"]
    },
    {
      "description": "Stop the robbery",
      "priority": 2,
      "progress": 0.3,
      "deadline": "4:30 remaining",
      "assigned_to": ["hero", "police", "vigilante"]
    }
  ]
}
```

This helps agents:
- Prioritize actions based on goal importance
- Recognize urgency from deadlines
- Identify which other characters share their goals for potential collaboration
- Make trade-offs between competing goals

## Scenario Definition Example

```toml
[scenario]
name = "Bank Heist"

[[goals]]
description = "Prevent civilian casualties"
priority = 1
assignment = ["hero", "police"]
evaluation = "all_civilians_safe"

[[goals]]
description = "Stop the robbery"
priority = 2
assignment = ["all"]
evaluation = "robbers_arrested_or_fled"
deadline = "10_minutes"

[[goals]]
description = "Protect the bank vault"
priority = 3
assignment = ["security"]
evaluation = "vault_integrity"
completion_threshold = 0.8
```