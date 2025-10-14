# Simulation Instrumentation & Chronicle System

## Overview

The simulation chronicle is a detailed, multi-layered record of everything that happens during a simulation. It captures not just observable events, but also the internal thoughts, emotions, reasoning, and perceptions of each agent. The goal is to provide writers with rich raw material for inspiration, not to generate final content.

## Design Principles

1. **Inspiration, Not Generation**: System produces material to spark ideas, not finished prose
2. **Transparency**: Expose the "why" behind every action and decision
3. **Multi-Perspective**: See the same moment through different characters' eyes
4. **Searchable**: Find specific themes, emotions, or turning points
5. **Layered Detail**: From high-level summary to granular internal dialogue
6. **Privacy-Aware**: Track what each character knows vs. what's hidden

## Chronicle Layers

The chronicle captures events at multiple levels of detail:

### Layer 1: Observable Events (Public)
What actually happened in the scene that characters could witness:
- Physical actions (movement, manipulation)
- Spoken dialogue
- Visible gestures and expressions
- Environmental changes
- Time progression

**Example:**
```
[5:30 PM] Alex and Jordan sitting in coffee shop
[5:31 PM] Alex speaks: "Hey! So I've been dying to try that new Mediterranean place..."
[5:32 PM] Jordan speaks: "Mediterranean sounds nice, but I don't know..."
[5:33 PM] Alex speaks: "You make good points about the parking..."
```

### Layer 2: Internal State (Private)
What each character was experiencing internally:
- Emotional states and shifts
- Focus of attention
- Physical sensations (fatigue, tension, etc.)
- Goal awareness and priority
- Certainty/confidence levels

**Example:**
```
[5:31 PM] Alex (internal):
  Emotion: neutral → happy (7/10)
  Focus: Jordan
  Energy: 100/100
  Active goals: [Agree on restaurant]
  Certainty: High (0.9)

[5:32 PM] Jordan (internal):
  Emotion: neutral (6/10)
  Focus: Alex
  Energy: 100/100
  Active goals: [Agree on restaurant]
  Certainty: Medium (0.7)
  Thought: "Worried about cost..."
```

### Layer 3: Internal Monologue (Private)
The character's actual thoughts and reasoning:
- LLM reasoning chains
- Internal reactions to events
- Unspoken feelings
- Mental planning
- Memory associations

**Example:**
```
[5:32 PM] Jordan (thinking):
  "Alex suggested a new place downtown. I'm cautious about new restaurants
  and worried about price. Downtown parking can be expensive too. I should
  ask practical questions about cost and location before agreeing. What
  about that Italian place we know?"

[5:33 PM] Alex (thinking):
  "Jordan raised valid concerns about price and parking. I want to try the
  new place, but I also want us both to be happy. Maybe I can address the
  concerns but also stay open to compromise. The goal is agreement, not
  winning."
```

### Layer 4: Perception Details (Private)
What each character noticed or missed:
- What they observed and how
- Perception confidence levels
- Details that stood out
- Things they failed to notice
- Interpretation and biases

**Example:**
```
[5:31 PM] Jordan perceives:
  Visual: Alex leaning forward (excitement detected)
  Auditory: Enthusiastic tone, describing food in detail
  Interpretation: "Alex really wants to try this place"
  Missed: Alex's willingness to compromise (not yet revealed)

[5:32 PM] Alex perceives:
  Visual: Jordan's neutral expression
  Auditory: Cautious tone, asking practical questions
  Interpretation: "Jordan is worried about money and convenience"
  Noticed: Jordan didn't say no, left door open
```

### Layer 5: Memory Activations (Private)
What memories influenced each decision:
- Retrieved memories
- Relevance scores
- How memories shaped response
- New memories being formed

**Example:**
```
[5:33 PM] Alex memory activity:
  Retrieved: "I suggested Olive & Thyme to Jordan" (confidence: 1.0)
  Retrieved: None from "negotiating restaurant choice" (no prior experiences)
  Forming: "Jordan concerned about price/parking, suggested Maple Street
           alternative. I offered compromise."
  Importance: 0.8 (high - relates to active goal)
```

### Layer 6: Decision Traces (Private)
The full reasoning process behind each action:
- Available options considered
- Personality traits that influenced choice
- Goal alignment evaluation
- Risk/reward assessment
- Final decision rationale

**Example:**
```
[5:33 PM] Alex decision process:

  Options considered:
  1. Insist on Olive & Thyme (low probability - conflicts with values)
  2. Concede to Maple Street (medium - safe but sacrifices novelty)
  3. Offer compromise/choice (high - aligns with "open to compromise")

  Personality influence:
  - Trait "enthusiastic" → still want to try new place
  - Trait "adventurous" → push for novelty
  - Decision style "open to compromise" → offer genuine choice
  - Value "quality" → both options meet this

  Goal evaluation:
  - Primary goal: Agreement (priority 1)
  - Insisting risks disagreement
  - Compromise most likely to achieve goal

  Selected: Offer compromise, give Jordan choice
  Confidence: High (0.85)

  Reasoning: "Jordan's concerns are valid. Addressing them while staying
             flexible maximizes chance of agreement. Goal is we both feel
             good about it, not that I get my way."
```

## Chronicle Structure

### Event Record Format

Each chronicle entry captures comprehensive context:

```json
{
  "event_id": "evt_003",
  "timestamp": 120,
  "simulation_time": "5:32 PM",
  "turn_number": 3,
  "narrative_mode": "NARRATIVE",

  "observable": {
    "type": "speech",
    "actor": "Alex",
    "target": "Jordan",
    "action": "speak",
    "content": "You make good points about the parking...",
    "volume": "normal",
    "visible_to": ["Jordan"],
    "audible_to": ["Jordan"]
  },

  "actor_internal": {
    "character": "Alex",
    "state_before": {
      "emotion": "happy",
      "intensity": 7,
      "focus": "Jordan",
      "condition": 100,
      "certainty": 0.9
    },
    "state_after": {
      "emotion": "happy",
      "intensity": 6,
      "focus": "Jordan",
      "condition": 100,
      "certainty": 0.9
    },
    "thoughts": "Jordan raised valid concerns about price and parking. I want to try the new place, but I also want us both to be happy. Maybe I can address the concerns but also stay open to compromise. The goal is agreement, not winning.",

    "perception": {
      "observed": "Jordan's cautious response about cost and parking",
      "interpreted_as": "practical concerns, not rejection",
      "confidence": 0.8,
      "emotional_read": "Jordan is being careful, not hostile"
    },

    "memories_activated": [
      {
        "content": "I suggested Olive & Thyme to Jordan",
        "timestamp": 0,
        "relevance": 1.0
      }
    ],

    "decision_process": {
      "options_considered": [
        "Insist on Olive & Thyme",
        "Concede to Maple Street",
        "Offer compromise"
      ],
      "personality_factors": {
        "traits": ["enthusiastic", "open to compromise"],
        "values": ["novelty", "quality"],
        "decision_style": "Quick to suggest but open to persuasion"
      },
      "goal_alignment": {
        "primary_goal": "Agree on restaurant",
        "chosen_action_supports": 0.85
      },
      "selected_action": "Offer compromise and choice",
      "confidence": 0.85,
      "reasoning": "Jordan's concerns are valid. Addressing them while staying flexible maximizes chance of agreement."
    },

    "memory_formation": {
      "content": "Jordan expressed concerns about price and parking at Olive & Thyme, suggested Italian place. I offered compromise.",
      "importance": 0.8,
      "emotional_valence": 0.2,
      "participants": ["Jordan"],
      "tags": ["negotiation", "compromise", "restaurant_choice"]
    }
  },

  "observer_reactions": [
    {
      "character": "Jordan",
      "perceived": "Alex acknowledged my concerns and offered genuine choice",
      "emotional_response": "neutral → slightly positive",
      "interpretation": "Alex is being flexible and considerate",
      "influenced_by_traits": ["practical", "cautious"]
    }
  ],

  "goal_impact": {
    "goal_id": "goal_001",
    "progress_before": 0.3,
    "progress_after": 0.6,
    "evaluation": "Compromise offer moves toward agreement"
  },

  "narrative_significance": {
    "dramatic_tension": 0.4,
    "character_development": ["Alex shows flexibility"],
    "relationship_change": "Alex-Jordan trust increased slightly",
    "themes": ["compromise", "friendship", "decision-making"]
  }
}
```

### Chronicle Organization

Chronicles are organized for easy exploration:

```
simulation_chronicle/
├── metadata.json                 # Simulation info, duration, participants
├── timeline.json                 # Chronological event sequence
├── by_character/
│   ├── alex_pov.json            # Alex's complete experience
│   ├── jordan_pov.json          # Jordan's complete experience
│   └── omniscient.json          # All internal states visible
├── by_turn/
│   ├── turn_001.json            # All events in turn 1
│   ├── turn_002.json
│   └── ...
├── by_theme/
│   ├── negotiations.json        # All negotiation moments
│   ├── compromises.json         # Compromise events
│   └── emotional_shifts.json    # Significant emotion changes
├── dialogue_only.json           # Just observable speech
├── decisions.json               # Decision traces only
└── summary.json                 # High-level overview
```

## Writer Query Interface

Writers can explore the chronicle through flexible queries:

### Query Examples

**Find emotional turning points:**
```
query: "Show me when Jordan's emotion changed from neutral to happy"
result: Turn 4, after Alex offered compromise
context: Full event with internal monologue
```

**Track a theme:**
```
query: "Find all moments involving 'compromise'"
result: Turns 3-4
context: Both characters' thoughts about compromise
```

**Character perspective:**
```
query: "Show me Jordan's internal experience of turn 2"
result: {
  "heard": "Alex's enthusiastic pitch for Olive & Thyme",
  "thought": "Worried about price and parking",
  "felt": "Slight anxiety about unfamiliar choice",
  "decided": "Ask practical questions, suggest alternative",
  "said": "How expensive is it? What about Maple Street?"
}
```

**Decision analysis:**
```
query: "Why did Alex offer a compromise in turn 3?"
result: {
  "reasoning": "Jordan's concerns were valid...",
  "personality": "Open to compromise trait activated",
  "goal_driven": "Agreement more important than getting my way",
  "alternatives_rejected": ["Insist on choice", "Give in completely"]
}
```

**Relationship dynamics:**
```
query: "How did the Alex-Jordan relationship evolve?"
result: Timeline of trust/rapport changes with supporting events
```

## Export Formats

### 1. Full Chronicle (JSON)
Complete machine-readable record with all layers

### 2. Narrative View
Formatted as readable prose with expandable internal thoughts:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
COFFEE SHOP - 5:31 PM

Alex leaned forward, eyes bright with excitement. "Hey! So I've been
dying to try that new Mediterranean place that opened downtown—Olive &
Thyme. I heard they have this amazing grilled octopus and everything's
super fresh. Plus it's supposed to be really healthy. What do you think?"

  💭 Alex (thinking): "I need to start the conversation about dinner.
     Given my enthusiasm for food and love of trying new places, I
     should suggest something interesting."

  😊 Emotion: neutral → happy (7/10)

Jordan considered this for a moment, eyebrows slightly raised.

  💭 Jordan (thinking): "Alex suggested a new place downtown. I'm
     cautious about new restaurants and worried about price..."

  😐 Emotion: neutral (6/10)

"Mediterranean sounds nice, but I don't know... how expensive is it?
And downtown—isn't parking a nightmare there? What about that Italian
place on Maple Street we've been to before? I know it's good and it's
pretty affordable."

  💭 Jordan (thinking): "I should ask practical questions before
     agreeing. Downtown parking can be expensive too."
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### 3. Character Study
Focus on single character's complete arc:

```
ALEX'S JOURNEY - "Dinner Plans" Simulation

INITIAL STATE:
  Emotion: neutral (5/10)
  Goals: Agree on restaurant
  Traits influencing behavior: enthusiastic, adventurous, open to compromise

TURN 1: Enthusiastic Proposal
  Said: Pitched Olive & Thyme with passion
  Thought: "I should suggest something exciting and healthy"
  Felt: Excited (7/10)
  Goal progress: 20% (initiated conversation)

TURN 2: Receiving Pushback
  Heard: Jordan's concerns about price and parking
  Thought: "Jordan suggested familiar option instead"
  Felt: Still happy but slightly less intense (7/10)
  Noticed: Jordan didn't reject outright, just expressed concerns

TURN 3: Finding Compromise
  Thought: "Jordan's concerns are valid. I want us both happy."
  Decided: Offer genuine choice rather than insist
  Said: "What would make you happier?"
  Felt: Slightly less excited but more aligned (6/10)
  Personality trait activated: "open to compromise"
  Goal progress: 60% (near agreement)

OUTCOME:
  Agreement reached on Alex's original choice (Olive & Thyme)
  But achieved through genuine compromise and respect
  Final emotion: happy (6/10) - satisfied with process and result

CHARACTER INSIGHTS:
  - Values novelty but also values relationship harmony
  - Willing to sacrifice preference for mutual satisfaction
  - Shows flexibility when presented with valid concerns
  - Goal-oriented but not rigid
```

### 4. Dialogue Script
Clean dialogue with stage directions:

```
INT. COFFEE SHOP - EVENING

ALEX and JORDAN sit across from each other. Coffee cups between them.

ALEX
(leaning forward, excited)
Hey! So I've been dying to try that new
Mediterranean place that opened downtown—
Olive & Thyme. I heard they have this
amazing grilled octopus and everything's
super fresh. Plus it's supposed to be
really healthy. What do you think?

JORDAN
(considering, cautious)
Mediterranean sounds nice, but I don't
know... how expensive is it? And downtown—
isn't parking a nightmare there?
(beat)
What about that Italian place on Maple
Street we've been to before? I know it's
good and it's pretty affordable.

ALEX
(nodding, acknowledging concerns)
You make good points about the parking.
How about this—what if we try Olive &
Thyme tonight since it's early and parking
might be easier, and if it's too expensive
we can stick with Maple Street in the future?
(earnest)
Or, honestly, Maple Street is great too.
Which would make you happier?

JORDAN
(warming, smile appearing)
You know what, you're right—it is early
so parking should be okay. And if it's
too pricey we can just keep it in mind
for special occasions.
(decided)
Let's try Olive & Thyme. I'm curious
about that octopus now!
```

### 5. Decision Tree Visualization
Branching possibilities and actual path taken:

```
Turn 3: Alex's Decision Point
│
├─ Option 1: Insist on Olive & Thyme
│   Probability: 15%
│   Personality alignment: Low (conflicts with "open to compromise")
│   Goal risk: High (might cause disagreement)
│   └─ REJECTED
│
├─ Option 2: Concede to Maple Street
│   Probability: 25%
│   Personality alignment: Medium (sacrifices "adventurous")
│   Goal risk: Low (safe agreement)
│   └─ REJECTED (misses opportunity for win-win)
│
└─ Option 3: Offer Compromise ✓ CHOSEN
    Probability: 60%
    Personality alignment: High (matches "open to compromise")
    Goal risk: Low (high chance of agreement)
    Result: Jordan agreed → Goal achieved
```

### 6. Insights Report
Thematic and analytical summary:

```
SIMULATION INSIGHTS: "Dinner Plans"

DURATION: 4 turns (~3 minutes)
MODE: Narrative throughout
OUTCOME: Success - Goal achieved

THEMES DETECTED:
1. Compromise and Negotiation (3 occurrences)
2. Practical vs. Adventurous (2 occurrences)
3. Friendship and Consideration (4 occurrences)

EMOTIONAL ARCS:
Alex: neutral → happy → happy → happy (stable positive)
Jordan: neutral → neutral → neutral → happy (gradual warming)

DECISION PATTERNS:
Alex: Quick proposer, willing compromiser
Jordan: Careful evaluator, persuadable with logic

RELATIONSHIP DYNAMICS:
- Started: neutral acquaintance
- Ended: positive rapport (trust increased)
- Key moment: Alex's genuine offer of choice (turn 3)

CHARACTER REVELATIONS:
Alex: Values novelty but relationships more
Jordan: Cautious but not inflexible
Both: Prioritize mutual satisfaction over individual preference

POTENTIAL STORY BEATS:
1. Initial excitement vs. practical concerns
2. Recognition of other's needs
3. Genuine compromise offered
4. Agreement through mutual respect

WRITER PROMPTS:
- What if Jordan hadn't been persuaded?
- How would this dynamic shift in higher-stakes situation?
- What does Alex's compromise reveal about their character?
```

## Privacy and Knowledge Tracking

The chronicle distinguishes between:

### Public Knowledge (Omniscient)
- All observable events
- Full internal states of all characters
- All decision reasoning
- Available to writer only, not to characters

### Character Knowledge (Limited)
- What each character observed
- What they were told
- What they inferred (correctly or not)
- Their own internal state only

**Example:**
```json
{
  "event": "Alex offers compromise",

  "omniscient_knowledge": {
    "alex_thinking": "I want us both happy, not just to get my way",
    "jordan_thinking": "Alex is being really considerate",
    "truth": "Both genuinely want agreement"
  },

  "alex_knows": {
    "observable": "Jordan expressed concerns about price/parking",
    "inferred": "Jordan prefers familiar places",
    "unknown": "Jordan's internal thought process"
  },

  "jordan_knows": {
    "observable": "Alex offered genuine choice",
    "inferred": "Alex values my input",
    "unknown": "How much Alex wanted Olive & Thyme initially"
  }
}
```

This enables writers to:
- Understand dramatic irony opportunities
- See where characters misunderstand each other
- Track secrets and hidden information
- Explore what-if scenarios based on different knowledge states

## Real-Time Chronicle Building

During simulation execution:

```
[Turn 1] Alex's action
  └─ Observable: Speech captured
  └─ Internal: Thought process logged
  └─ Decision: Reasoning saved
  └─ Memory: Formation tracked

[Turn 2] Jordan's action
  └─ Observable: Speech captured
  └─ Internal: Emotional response logged
  └─ Perception: How they interpreted Alex's words
  └─ Decision: Why they chose to push back

[Chronicle Update] After each turn
  └─ Event added to timeline
  └─ Character POVs updated
  └─ Theme tags applied
  └─ Goal progress logged
  └─ Relationship dynamics updated
```

## Implementation Considerations

### Storage
- SQLite for structured event data
- JSON files for export
- Full-text search indexing for queries
- Vector embeddings for semantic search

### Performance
- Async logging (don't block simulation)
- Batch writes for efficiency
- Incremental chronicle building
- On-demand export generation

### Privacy
- Clear separation of omniscient vs. character knowledge
- Ability to export character-specific views
- Filtering sensitive information if needed

## Use Cases for Writers

1. **Direct Inspiration**: Read dialogue and internal thoughts
2. **Character Studies**: Understand motivations deeply
3. **Plot Mining**: Find unexpected twists or conflicts
4. **Dialogue Refinement**: See natural conversation flow
5. **Emotional Beats**: Track feeling changes for pacing
6. **Theme Discovery**: Identify emergent patterns
7. **Alternative Paths**: Explore rejected decision options
8. **Relationship Mapping**: Understand character dynamics
9. **Pacing Analysis**: See where tension built or relaxed
10. **Voice Development**: Study character-specific language patterns

The chronicle is a creative tool, not a replacement for writing. It generates the raw ore that writers refine into gold.