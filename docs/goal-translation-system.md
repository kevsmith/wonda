# Goal Translation and Refinement System

## Overview

The goal translation system enables fiction writers to define simulation goals in natural language, then collaboratively refines those goals through conversational iteration until the writer is satisfied. The system maintains both human-readable descriptions and machine-evaluable goal structures.

## Design Principle

Writers should never need to see TOML, understand goal types, or write code. The entire goal definition process happens through natural language conversation, with the system translating to and from structured representations transparently.

## Goal Architecture

### Dual Representation

Every goal exists in two forms:

**Natural Language (Writer-Facing)**
```
"Save the children from the evil monster"
```

**Structured Form (System-Facing)**
```json
{
  "id": "goal_001",
  "description": "Save the children from the evil monster",
  "priority": 1,
  "assignment": ["Hero", "Guardian"],
  "type": "RescueGoal",
  "parameters": {
    "targets": ["children"],
    "threat": "evil_monster",
    "threat_resolution": "defeat_or_distance",
    "distance_threshold": 50,
    "rescue_threshold": 1.0
  },
  "evaluation": {
    "strategy": "world_state_check",
    "conditions": [
      "all_entities_tagged('children').location.safety >= 0.8",
      "entity('evil_monster').energy <= 0 OR distance(children, monster) > 50"
    ]
  }
}
```

Writers interact only with the natural language form. The system handles all structured translation.

## Goal Evaluation Strategies

Goals use different evaluation strategies based on their nature:

### 1. Agent-Assessed Goals
Character judgment required to determine completion.

**Examples:**
- "Agree on a restaurant" (ConsensusGoal)
- "Feel safe" (SubjectiveGoal)
- "Trust each other" (RelationshipGoal)

**Evaluation:**
- Agents use MCP tools: `assess_goal()`, `vote()`, `assess_agreement()`
- Multiple agents may need to agree
- Agent reasoning is captured in chronicle

### 2. World State Goals
Evaluated by checking simulation state automatically.

**Examples:**
- "Defeat the villain" (CombatGoal)
- "Reach the exit" (LocationGoal)
- "Save the children" (RescueGoal)

**Evaluation:**
- System checks state predicates after each turn
- No agent input needed
- Purely objective conditions

### 3. Event-Monitored Goals
Evaluated by watching event stream for patterns.

**Examples:**
- "Avoid violence" (AvoidanceGoal)
- "Survive for 10 turns" (TimeGoal)
- "Keep the secret hidden" (SecrecyGoal)

**Evaluation:**
- Monitor events for success/failure triggers
- Often binary (achieved or failed)
- May track proximity to failure

### 4. Hybrid Goals
Combine multiple evaluation methods.

**Examples:**
- "Rescue hostages without being seen" (state + stealth monitoring)
- "Convince the king and maintain his respect" (consensus + relationship)

**Evaluation:**
- Multiple conditions must all be satisfied
- Each component uses appropriate strategy

## Conversational Refinement Workflow

### Round-Trip Translation

```
Writer Input (NL) → LLM Translation → Structured Goal
                                           ↓
Writer Feedback ← LLM Generation ← Natural Language Confirmation
        ↓
   [Refine or Confirm]
```

### Phase 1: Initial Translation

**Writer Input:**
```
Goal: "Save the children from the evil monster"
Priority: Highest
Assigned to: Hero, Guardian
```

**System Processing:**
```python
# LLM receives prompt:
"""
Translate this goal to a structured goal type.

Writer's goal: "Save the children from the evil monster"
Assigned to: Hero, Guardian
Priority: 1 (highest)

Available goal types:
- RescueGoal: Protect/save entities from threats
- ConsensusGoal: Multiple parties must agree
- StateGoal: Achieve specific world state
- CombatGoal: Defeat enemies
- AvoidanceGoal: Prevent events
[... full type catalog with schemas ...]

Return JSON with:
{
  "goal_type": "<type>",
  "parameters": {...},
  "evaluation_strategy": "<strategy>",
  "confirmation": "<natural language explanation>"
}
"""

# LLM Response:
{
  "goal_type": "RescueGoal",
  "parameters": {
    "targets": ["children"],
    "threat": "evil_monster",
    "rescue_threshold": 1.0,
    "threat_resolution": "defeat_or_distance"
  },
  "evaluation_strategy": "world_state_check",
  "confirmation": "The Hero and Guardian must get all the children to safety and either defeat the monster or get it far away from them. I'll know this is complete when all children are in a safe location and the monster is either defeated or at least 50 meters away."
}
```

**User Sees:**
```
Got it. The Hero and Guardian must get all the children to safety
and either defeat the monster or get it far away from them.

I'll know this is complete when:
- All children are in a safe location
- The monster is either defeated or at least 50 meters away

Is that what you meant?

[Yes, that's right] [No, let me clarify]
```

### Phase 2: Iterative Refinement

**Writer Response:**
```
No, let me clarify: The monster just needs to be far away, not defeated.
And only most of the children need to be saved, not all.
```

**System Processing:**
```python
# LLM receives conversation history + new input
"""
Previous understanding: RescueGoal with defeat_or_distance, 100% rescue threshold

Writer correction: "Monster just needs to be far away, not defeated.
Only most of the children need to be saved, not all."

Update the goal parameters and generate new confirmation.
"""

# LLM Response:
{
  "goal_type": "RescueGoal",
  "parameters": {
    "targets": ["children"],
    "threat": "evil_monster",
    "rescue_threshold": 0.8,
    "threat_resolution": "distance_only",
    "distance_threshold": 50
  },
  "evaluation_strategy": "world_state_check",
  "confirmation": "The Hero and Guardian must get most of the children to safety (at least 80%) and drive the monster away. I'll know this is complete when at least 4 out of 5 children are in a safe location and the monster is at least 50 meters from the children.",
  "changes": ["Removed defeat requirement", "Changed rescue threshold to 80%"]
}
```

**User Sees:**
```
Updated! The Hero and Guardian must get most of the children to
safety (at least 80%) and drive the monster away.

I'll know this is complete when:
- At least 4 out of 5 children are in a safe location
- The monster is at least 50 meters from the children

Better?

[Yes, perfect] [Keep refining]
```

### Phase 3: Confirmation

**Writer:** Clicks "Yes, perfect"

**System:**
```
✓ Goal confirmed and saved.

Summary:
  Goal: "Save the children from the evil monster"
  Type: Rescue Mission
  Priority: Highest
  Assigned to: Hero, Guardian

[Edit this goal] [Add another goal] [Continue]
```

## Handling Ambiguity

When the LLM detects ambiguity, it asks clarifying questions:

**Example 1: Subjective vs. Objective**
```
Writer: "Keep everyone calm"

System: I have a question about "Keep everyone calm":

Do you mean:
A) The characters should feel calm emotionally, even if the situation is tense
B) The situation itself should remain peaceful (no fighting or chaos)
C) Both - characters calm AND situation peaceful

Which matches your story better?
```

**Example 2: Scope Clarification**
```
Writer: "The detective figures out who the killer is"

System: Just to clarify - should the detective:

A) Believe she's found the killer (even if she's wrong)?
B) Actually identify the correct killer (I'd need to know who it really is)?

Option A creates dramatic irony if she's wrong.
Option B requires the right answer to be defined.

Which fits your story?
```

**Example 3: Multi-Part Goals**
```
Writer: "The detective needs to figure out who the killer is without
anyone knowing she's investigating"

System: This seems like it has two parts:

1) The detective must discover the killer's identity
2) She must do this secretly (others don't know she's investigating)

Should I treat this as:
- One combined goal (both must succeed together), or
- Two separate goals (they could succeed/fail independently)?

For a combined goal, failure in either part means total failure.
```

## Complex Goal Examples

### Stealth Discovery Goal
```
Writer: "The spy must steal the plans without being detected"

System: The Spy must obtain the secret plans while remaining undetected
by any guards or security systems.

I'll know this is complete when:
- The Spy has the plans in their inventory
- No guard or security entity has detected the Spy's presence

This combines:
- Object acquisition (world state check)
- Stealth maintenance (event monitoring for detection)

Both conditions must be true. Sound right?
```

### Relationship Consensus Goal
```
Writer: "The two rivals need to agree to work together and actually
trust each other"

System: The Rival A and Rival B must reach genuine agreement to
collaborate and develop mutual trust.

I'll know this is complete when:
- Both rivals explicitly agree to work together
- Both rivals report trusting each other
- Their relationship score shows positive rapport

This uses:
- Consensus detection (they must both say yes)
- Subjective assessment (they self-report trust)
- Objective measurement (relationship state check)

All three must be satisfied. Is that what you're after?
```

## Natural Language Generation

The system generates confirmations that:

1. **Restate the goal** in clear language
2. **Explain completion conditions** without jargon
3. **Identify who's responsible**
4. **Describe how completion is tracked**
5. **Mention any nuances** (thresholds, special conditions)

### Generation Prompt Template
```
You are explaining a simulation goal back to a fiction writer in clear,
natural language. Avoid technical jargon.

Goal type: {goal_type}
Description: {original_description}
Parameters: {json.dumps(parameters)}
Evaluation strategy: {evaluation_strategy}

Generate a friendly confirmation that:
1. Restates what needs to happen
2. Explains how completion is detected
3. Mentions any thresholds or special conditions
4. Asks if this matches their intent

Format:
"[Who] must [what]. I'll know this is complete when [conditions].

Is that what you meant?"

Keep it conversational and writer-friendly.
```

## Conversation State Management

The refinement process maintains conversation state:

```python
class GoalRefinementSession:
    def __init__(self, goal_description, assignment, priority):
        self.conversation = [
            {
                "role": "system",
                "content": GOAL_TRANSLATION_SYSTEM_PROMPT
            },
            {
                "role": "user",
                "content": f"Goal: {goal_description}\nAssigned to: {assignment}\nPriority: {priority}"
            }
        ]
        self.current_goal = None
        self.iterations = 0

    def refine(self, user_feedback):
        """Process writer feedback and update goal"""
        self.conversation.append({
            "role": "user",
            "content": user_feedback
        })

        response = llm.chat(
            messages=self.conversation,
            response_format=GoalTranslationResponse
        )

        self.current_goal = response.structured_goal
        self.conversation.append({
            "role": "assistant",
            "content": response.confirmation
        })

        return response.confirmation

    def confirm(self):
        """Finalize goal after writer approval"""
        return self.current_goal
```

## Error Handling

### Unrecognized Goal Type
```
Writer: "The quantum flux must stabilize the temporal vortex"

System: I'm not sure I understand this goal. It doesn't match any of
the goal types I know about.

Could you rephrase it? For example:
- "Someone must do something" (who and what?)
- "Something must be prevented"
- "A certain state must be achieved"

What's the core objective here?
```

### Conflicting Requirements
```
Writer: "The hero must defeat the villain without fighting"

System: I'm seeing a contradiction - "defeat the villain" usually means
combat, but you said "without fighting."

Did you mean:
- Defeat through non-violent means (outsmarting, trapping, etc.)?
- Drive the villain away without a fight?
- Something else?

Help me understand what "defeat without fighting" looks like.
```

## Integration with Scenario Definition

Goals are refined as part of scenario creation:

```
SCENARIO BUILDER

1. Basic Info
   Name: "The Monster Attack"
   Description: "A village under siege"

2. Characters (3 defined) ✓

3. Goals ← YOU ARE HERE

   Goal 1: "Save the children from the evil monster"
   ✓ Confirmed - Rescue mission

   [+ Add another goal]

4. Scene Setup
5. Review & Test
```

## Testing and Validation

Writers can test goals before running full simulation:

```
GOAL TEST

Goal: "Save the children from the evil monster"

Test scenario:
  5 children in the village square
  Monster 10 meters away, health: 100
  Hero and Guardian present

After 5 turns:
  3 children moved to safe house
  Monster driven to 60 meters away
  2 children still in square

Goal progress: 60%
  ✓ Monster is far enough (>50m)
  ✗ Only 60% of children saved (need 80%)

This matches your goal definition. Proceed?
[Yes] [Adjust goal] [Adjust test scenario]
```

## Implementation Considerations

### LLM Provider
- Must support structured output (JSON mode)
- Conversation/chat completion endpoint required
- Local (Ollama) or cloud (OpenAI) compatible
- Fast iteration important for good UX

### Prompt Engineering
- Clear goal type definitions with examples
- Explicit instructions for confirmation generation
- Handling of edge cases and ambiguity
- Consistent format for structured responses

### Cost Management
- Each refinement iteration = 1 LLM call
- Average 2-3 iterations per goal
- 3-5 goals per scenario
- Total: ~10-15 LLM calls per scenario setup
- Local LLM: Free, ~1-2 seconds per call
- OpenAI: ~$0.02-0.05 per scenario

### Performance
- Async LLM calls where possible
- Cache goal type catalog
- Stream responses for better UX
- Timeout handling for slow LLMs

## Future Enhancements

### Goal Templates Library
```
"Rescue mission" → Pre-fill RescueGoal parameters
"Negotiate agreement" → Pre-fill ConsensusGoal parameters
"Stealth operation" → Pre-fill combined acquisition + avoidance
```

### Learning from Corrections
Track common refinement patterns to improve initial translations

### Multi-Language Support
Natural language interface works in any language LLM supports

### Goal Composition
"Combine these two goals into one" - merge related objectives

### Visual Goal Builder
For writers who prefer form-based input alongside conversational