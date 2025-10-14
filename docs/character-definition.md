# Character Definition

## Overview
Characters are templates that define roles within scenarios. When a simulation begins, agents are instantiated from these character definitions, binding them with LLM providers, models, and runtime state.

## Character Structure

### Core Components

A character definition consists of:

```toml
[character]
# Identity
name = "string"                    # Character name or archetype
description = "string"             # What/who this character is (10-1000 chars)
background = "string"              # History, context, secrets (optional, max 2000 chars)

# Personality
traits = ["string"]                # Behavioral characteristics (min 1)
communication_style = "string"     # How they speak and interact (10-500 chars)
decision_style = "string"          # How they make choices (10-500 chars)
skills = ["string"]                # Expertise areas (optional)
values = ["string"]                # Beliefs and principles (optional)

# Initial State (optional defaults)
initial_position = "string"        # Starting location
initial_condition = 100            # Health/energy (0-100, default 100)
initial_emotion = "neutral"        # Starting emotion (default "neutral")
initial_emotion_intensity = 5      # 0-10 (default 5)
```

## Detailed Field Descriptions

### Identity Fields

**name**
- Character name or archetype descriptor
- Examples: "Detective Sarah Chen", "Scrappy youth", "Kindly teacher"
- Used for agent instantiation and narrative generation

**description** (10-1000 characters)
- Core definition of who/what the character is
- Includes role, basic personality, motivations
- Example: "A world-weary detective who's seen too much corruption but still believes justice is possible. Methodical and observant, she misses nothing."

**background** (optional, max 2000 characters)
- Detailed history and context
- Can include secrets unknown to other characters
- Informs decision-making and reactions
- Example: "Grew up in the same neighborhood she now patrols. Her brother was killed by gang violence when she was 16, driving her to become a cop. Has an informant network built over 15 years on the force."

### Personality Fields

**traits** (minimum 1 required)
- Concise behavioral descriptors
- Shape how agent interprets situations and chooses actions
- Examples: `["analytical", "cynical", "protective", "stubborn"]`
- Common traits:
  - Emotional: `anxious`, `confident`, `empathetic`, `cold`
  - Social: `charming`, `awkward`, `manipulative`, `honest`
  - Cognitive: `analytical`, `impulsive`, `creative`, `methodical`
  - Moral: `principled`, `pragmatic`, `ruthless`, `compassionate`

**communication_style** (10-500 characters)
- How the character speaks and interacts
- Affects dialogue generation and social actions
- Include: tone, vocabulary level, directness, emotional expression
- Example: "Speaks in short, direct sentences. Rarely shows emotion but occasional dry humor. Asks pointed questions to expose lies. Uses cop jargon naturally."

**decision_style** (10-500 characters)
- How the character makes choices under pressure
- Guides LLM reasoning during critical moments
- Include: risk tolerance, time preference, information needs
- Example: "Methodical - gathers evidence before acting. Willing to take calculated risks but not reckless. Prioritizes protecting innocents even at personal cost."

**skills** (optional)
- Areas of expertise or capability
- Informs what actions are credible for this character
- Examples: `["forensics", "interrogation", "street_knowledge", "marksmanship"]`

**values** (optional)
- Core beliefs and principles
- Create internal conflicts and drive goal prioritization
- Examples: `["justice", "loyalty", "protecting_the_innocent", "truth"]`

### Initial State Fields (Optional)

These provide default starting conditions when agents are instantiated:

**initial_position**
- Starting location in the scene
- If not specified, scenario determines placement

**initial_condition** (0-100, default 100)
- Starting health/energy level
- Useful for injured or exhausted characters

**initial_emotion** (default "neutral")
- Starting emotional state
- Options: `neutral`, `angry`, `afraid`, `happy`, `sad`

**initial_emotion_intensity** (0-10, default 5)
- How strongly they feel the initial emotion

## Example Characters

### Minimal Character
```toml
name = "Street Vendor"
description = "Elderly vendor who's worked this corner for 30 years. Sees everything, says little."
traits = ["observant", "cautious", "streetwise"]
communication_style = "Few words, thick accent. Speaks in questions rather than statements."
decision_style = "Self-preservation first. Helps when it doesn't cost him anything."
```

### Rich Character
```toml
name = "Detective Sarah Chen"
description = "Homicide detective, 15 years on the force. Brilliant investigator struggling with burnout and moral compromise in a corrupt department."
background = "Grew up in Chinatown, lost her brother to gang violence at age 16. Joined the force to make a difference but has watched good cops get pushed out or broken down. Recently discovered her captain is on the take. Drinks too much. Divorced, no kids. Lives for the job because it's all she has left."
traits = ["analytical", "cynical", "determined", "protective", "stubborn"]
communication_style = "Direct and professional with civilians. Sarcastic with colleagues. Asks probing questions. Uses silence strategically in interrogations. Occasional Cantonese phrases when stressed."
decision_style = "Methodical evidence-gatherer but will break rules for the right reasons. Prioritizes protecting innocents over following orders. Willing to take personal risks. Trusts her gut but backs it up with facts."
skills = ["forensics", "interrogation", "street_knowledge", "marksmanship", "cantonese", "reading_people"]
values = ["justice", "protecting_innocents", "truth", "loyalty_to_partners"]
initial_position = "crime_scene"
initial_condition = 75
initial_emotion = "tired"
initial_emotion_intensity = 6
```

### Antagonist Character
```toml
name = "Marcus Webb"
description = "Charismatic cult leader who genuinely believes he's saving humanity. Brilliant manipulator with a martyr complex."
background = "Former PhD student in sociology who had a breakdown after his research on mass movements was rejected. Spent two years homeless before emerging with a 'vision' of societal collapse. Built his following through online communities, targeting vulnerable people. Truly believes the end is near and he's gathering the chosen ones."
traits = ["charismatic", "delusional", "manipulative", "paranoid", "eloquent"]
communication_style = "Speaks in grand, philosophical terms. Uses religious and apocalyptic imagery. Makes intense eye contact. Voice shifts from gentle to commanding. Quotes obscure texts."
decision_style = "Sees every action through the lens of his mission. Ends justify means. Paranoid about betrayal. Will sacrifice individuals for the 'greater good' of his vision."
skills = ["psychology", "public_speaking", "manipulation", "philosophy", "social_media"]
values = ["the_mission", "loyalty_from_followers", "being_right", "control"]
initial_emotion = "confident"
initial_emotion_intensity = 9
```

## Character Assignment in Scenarios

Characters are defined within scenario files and assigned goals:

```toml
[scenario]
name = "The Standoff"

[[characters]]
name = "Detective Chen"
description = "Veteran detective trying to talk down the hostage taker"
traits = ["calm", "experienced", "empathetic"]
communication_style = "Steady and reassuring. Uses active listening."
decision_style = "Patient. Prioritizes preserving life above all."

[[characters]]
name = "Armed Suspect"
description = "Desperate man who's cornered and terrified"
traits = ["desperate", "frightened", "volatile"]
communication_style = "Erratic. Voice cracks. Repeats himself."
decision_style = "Panic-driven. Looking for any way out."

[[goals]]
description = "Get everyone out alive"
priority = 1
assignment = ["Detective Chen"]

[[goals]]
description = "Escape without going to prison"
priority = 1
assignment = ["Armed Suspect"]
```

## Agent Instantiation

When a simulation begins:

1. **Character template** → **Agent instance**
2. Agent adds:
   - LLM provider configuration
   - Model and parameters (temperature, etc.)
   - Runtime state tracking
   - Memory systems (context + RAG)
   - Assigned goals from scenario

3. Character traits, personality, and background inform:
   - LLM system prompts
   - Decision-making weights
   - Action filtering
   - Perception biases (via distortion filters)

## Design Principles

**1. Fiction-First**
- Optimized for narrative quality, not simulation accuracy
- Rich personality over mechanical stats
- Secrets and internal conflicts drive drama

**2. Minimal Core, Rich Optional**
- Only 5 required fields (name, description, traits, communication_style, decision_style)
- Everything else is optional enrichment
- Simple characters work fine, detailed ones shine

**3. LLM-Friendly**
- Natural language descriptions, not numeric stats
- Traits and styles guide LLM behavior
- Background provides context for reasoning

**4. Reusability**
- Characters can be templates ("Detective archetype")
- Or specific individuals ("Detective Sarah Chen")
- Scenarios can clone and modify existing characters

## Validation Rules

**Required Fields:**
- name (1-100 characters)
- description (10-1000 characters)
- traits (at least 1)
- communication_style (10-500 characters)
- decision_style (10-500 characters)

**Optional Field Limits:**
- background (max 2000 characters)
- skills, values, traits (no max, but keep reasonable for LLM context)

**State Ranges:**
- initial_condition: 0-100
- initial_emotion_intensity: 0-10
- initial_emotion: must be one of [neutral, angry, afraid, happy, sad]

## Future Extensions

The character definition is designed for expansion:
- **Relationship templates**: Pre-existing relationships between characters
- **Character arcs**: Growth/change trajectories
- **Dialogue samples**: Example lines to guide voice
- **Appearance details**: For visual descriptions
- **Secrets**: Information hidden from other characters
- **Triggers**: Specific stimuli that provoke strong reactions