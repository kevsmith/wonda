# MVP Prompts

## Overview

This document contains the LLM prompt templates used in the MVP. All prompts use Jinja2 templating syntax.

## Character Field Reference

These fields come from the character definition (see `character-definition.md`):

**Required:**
- `name` - Character name
- `description` - Who/what this character is
- `traits` - List of behavioral characteristics
- `communication_style` - How they speak and interact
- `decision_style` - How they make choices

**Optional:**
- `background` - History and context
- `skills` - List of expertise areas
- `values` - List of beliefs and principles

## Agent State Fields

These fields come from runtime agent state (see `agent-architecture.md`):

**Minimal MVP State:**
- `emotion` - Current emotion (angry, afraid, happy, sad, neutral)
- `emotion_intensity` - 0-10 scale
- `goal_opinion` - Agent's assessment of goal progress (0.0-1.0)

## Dialogue Context Fields

**Runtime Data:**
- `dialogue_history` - List of recent messages
- `assigned_goals` - List of goals agent is pursuing
- `turn_number` - Current turn number

---

## Prompt 1: Agent Turn Decision

**Purpose:** Generate agent's decision for their turn, including internal thoughts, action, and goal assessment.

**Template:**

```jinja2
You are {{ name }}, {{ description }}

{% if background %}
BACKGROUND:
{{ background }}
{% endif %}

PERSONALITY:
Traits: {{ traits|join(", ") }}
Communication Style: {{ communication_style }}
Decision Style: {{ decision_style }}
{% if skills %}Skills: {{ skills|join(", ") }}{% endif %}
{% if values %}Values: {{ values|join(", ") }}{% endif %}

CURRENT STATE:
Emotion: {{ emotion }} (intensity {{ emotion_intensity }}/10)

ACTIVE GOALS:
{% for goal in assigned_goals %}
- [{{ goal.id }}] {{ goal.description }}
  Priority: {{ goal.priority }}/5 (1 is highest)
  {% if goal.assigned_to|length > 1 %}Shared with: {{ goal.assigned_to|reject('equalto', name)|join(', ') }}{% endif %}
  Your current assessment: {{ (goal_opinion * 100)|int }}% complete
{% endfor %}

{% if dialogue_history %}
RECENT CONVERSATION:
{% for message in dialogue_history %}
{% if message.speaker == name %}You{% else %}{{ message.speaker }}{% endif %}: "{{ message.content }}"
{% endfor %}
{% else %}
CONVERSATION:
This conversation is just beginning. You are about to speak first.
{% endif %}

AVAILABLE ACTIONS:
You can use these tools to interact:

1. speak(message)
   - Say something to the other person
   - Use this for dialogue and expressing ideas

2. listen()
   - Hear what the other person just said
   - Use this if you want to confirm what you heard

3. assess_goal(goal_id, completion, reasoning)
   - Report your assessment of a goal's progress
   - completion: 0.0 to 1.0 (0% to 100%)
   - reasoning: explain why you think this
   - Use this when you think a goal's status has changed

INSTRUCTIONS:
Respond with your decision for this turn. You must provide:

1. INTERNAL THOUGHTS: What are you thinking right now? Consider:
   - Your personality ({{ traits|slice(3)|join(', ') }})
   - Your goals and their current progress
   - What just happened in the conversation
   - How you're feeling ({{ emotion }})

2. CHOSEN ACTION: What will you do?
   - Select ONE action from the available actions above
   - Provide all required parameters

3. EMOTIONAL UPDATE: How do you feel after taking this action?
   - emotion: angry, afraid, happy, sad, or neutral
   - intensity: 0-10

4. GOAL ASSESSMENT: Do any of your goals need status updates?
   - If yes, include assess_goal call
   - If no, explain briefly why not

Remember:
- Stay true to your personality ({{ communication_style }})
- Make decisions based on your style ({{ decision_style }})
{% if values %}- Consider your values: {{ values|join(', ') }}{% endif %}
- Be authentic to who you are as {{ name }}

FORMAT YOUR RESPONSE AS JSON:
{
  "internal_thoughts": "your thinking here",
  "action": {
    "type": "speak|listen|assess_goal",
    "parameters": {
      // action-specific parameters
    }
  },
  "emotional_update": {
    "emotion": "angry|afraid|happy|sad|neutral",
    "intensity": 0-10
  },
  "goal_updates": [
    {
      "goal_id": "goal_001",
      "completion": 0.0-1.0,
      "reasoning": "why you assess it this way"
    }
    // ... more if needed
  ]
}
```

---

## Prompt 2: Perception Distortion (Stretch Goal)

**Purpose:** Apply semantic distortion to perception data (e.g., paranoia filter).

**Template:**

```jinja2
You are applying a {{ distortion_type }} filter to {{ character_name }}'s perception.

CHARACTER CONTEXT:
Name: {{ name }}
Traits: {{ traits|join(", ") }}
Current Emotion: {{ emotion }} ({{ emotion_intensity }}/10)
{% if background %}Background: {{ background }}{% endif %}

DISTORTION TYPE: {{ distortion_type }}
{% if distortion_type == "paranoia" %}
Effect: Interpret neutral or ambiguous statements as potentially threatening, manipulative, or hiding ulterior motives. Add suspicious interpretations.
{% elif distortion_type == "depression" %}
Effect: Dampen positive statements and emphasize negative aspects. Make hopeful things seem futile.
{% elif distortion_type == "mania" %}
Effect: Amplify excitement and possibilities. See connections and opportunities everywhere.
{% endif %}

DISTORTION INTENSITY: {{ distortion_intensity }}/10
- 0-3: Subtle reinterpretation
- 4-7: Moderate distortion, clearly different from original
- 8-10: Severe distortion, significantly altered meaning

ORIGINAL PERCEPTION:
{{ original_perception }}

INSTRUCTIONS:
Reinterpret the original perception through the lens of {{ distortion_type }}.

1. Keep the factual content (what was actually said/done)
2. Add the distorted interpretation in [brackets]
3. The distortion should reflect {{ character_name }}'s current state
4. Intensity {{ distortion_intensity }}/10 means {% if distortion_intensity <= 3 %}subtle hints{% elif distortion_intensity <= 7 %}clear reinterpretation{% else %}severe distortion{% endif %}

FORMAT YOUR RESPONSE AS JSON:
{
  "distorted_perception": "Original content with [distorted interpretation added]",
  "distortion_applied": "brief note on what changed",
  "original_preserved": true/false
}

EXAMPLE (paranoia, intensity 6):
Original: "What about that Italian place on Maple Street?"
Output: "What about that Italian place on Maple Street? [Why are they pushing so hard for that specific place? Do they know someone there? Are they trying to control where we go?]"
```

---

## Prompt Usage Notes

### Agent Turn Decision Prompt

**When to use:**
- Every agent turn
- After each agent receives perception data

**Data to populate:**
- Character definition fields (loaded from scenario)
- Current agent state (tracked by simulation)
- Dialogue history (recent N messages)
- Assigned goals (from scenario)

**Expected output:**
Structured JSON with:
- Internal thoughts (logged to chronicle)
- Action selection (routed to MCP server)
- Emotional update (updates agent state)
- Goal assessments (sent to goal monitor)

**Error handling:**
- If JSON parsing fails, retry with clarification
- If action is invalid, ask agent to choose again
- If no action selected, default to `listen()`

### Perception Distortion Prompt

**When to use:**
- After perception is generated but before agent processes it
- Only if agent has distortion filter configured

**Data to populate:**
- Character context (who is perceiving)
- Original perception (what actually happened)
- Distortion type and intensity (from character config)

**Expected output:**
Structured JSON with:
- Distorted version of perception
- Note on what changed
- Whether original meaning preserved

**Error handling:**
- If distortion fails, fall back to original perception
- Log failure but don't block simulation

---

## Prompt Engineering Notes

### Agent Turn Decision

**Key design choices:**

1. **Structured output:** JSON format ensures parseable responses
2. **Personality reinforcement:** Remind agent of traits/style multiple times
3. **Goal visibility:** Show progress assessment to encourage evaluation
4. **Available actions:** Explicitly list tools to guide valid choices
5. **Emotional continuity:** Current emotion influences next decision

**Testing considerations:**
- Verify agents stay in character across multiple turns
- Check that goal assessments are reasonable
- Ensure emotional changes make sense
- Validate JSON parsing success rate

### Perception Distortion

**Key design choices:**

1. **Bracketed additions:** Clear distinction between fact and distortion
2. **Intensity scaling:** Explicit guidance on how much to distort
3. **Character awareness:** Distortion reflects character's traits/state
4. **Preservation option:** Can choose to preserve original meaning

**Testing considerations:**
- Verify distortion is applied consistently
- Check intensity scaling works as intended
- Ensure factual content isn't lost
- Validate impact on agent decisions

---

## Prompt Iterations

These prompts will likely need refinement based on:

1. **LLM model behavior:** Different models may need different guidance
2. **Agent personality expression:** May need stronger personality cues
3. **Goal assessment accuracy:** May need examples or criteria
4. **JSON reliability:** May need schema examples or error recovery

**Improvement process:**
1. Run simulation with current prompts
2. Review chronicle for issues (out of character, invalid actions, poor assessments)
3. Adjust prompts to address issues
4. Test again with same scenario
5. Document changes and rationale

---

## Template Variables Reference

### Character Definition
```python
{
  "name": str,
  "description": str,
  "background": str | None,
  "traits": List[str],
  "communication_style": str,
  "decision_style": str,
  "skills": List[str] | None,
  "values": List[str] | None
}
```

### Agent State
```python
{
  "emotion": "angry" | "afraid" | "happy" | "sad" | "neutral",
  "emotion_intensity": int,  # 0-10
  "goal_opinion": float  # 0.0-1.0
}
```

### Dialogue History
```python
[
  {
    "speaker": str,  # Character name
    "content": str,  # What they said
    "turn": int
  },
  ...
]
```

### Goals
```python
[
  {
    "id": str,
    "description": str,
    "priority": int,  # 1-5
    "assigned_to": List[str]  # Character names
  },
  ...
]
```

### Distortion Config
```python
{
  "distortion_type": "paranoia" | "depression" | "mania" | ...,
  "distortion_intensity": int,  # 0-10
  "character_name": str,
  "original_perception": str
}
```