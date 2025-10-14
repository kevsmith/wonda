# RAG Memory System

## Overview
The RAG (Retrieval-Augmented Generation) memory system provides agents with persistent, searchable long-term memory that extends beyond their LLM context window.

## Memory Architecture

### Memory Types

#### Episodic Memory
- Specific events and experiences
- Timestamped with simulation time
- Includes: participants, location, actions, outcomes
- Example: "Sarah betrayed me in the alley at turn 15"

#### Semantic Memory
- Facts and knowledge learned during simulation
- Relationships between concepts
- World state understanding
- Example: "The red door leads to the basement"

#### Procedural Memory
- How to accomplish tasks
- Successful action sequences
- Learned strategies
- Example: "Negotiating with Bob requires appealing to greed"

#### Emotional Memory
- Feelings associated with entities/events
- Emotional responses to stimuli
- Trust/fear associations
- Example: "The warehouse makes me anxious"

### Memory Formation

#### Encoding Pipeline
1. **Experience occurs** → Short-term memory (LLM context)
2. **Importance evaluation**:
   - Relevance to goals (0-1)
   - Emotional impact (0-1)
   - Novelty factor (0-1)
   - Combined importance score
3. **Vector encoding** with metadata:
   - Timestamp
   - Participants
   - Location
   - Emotional valence
   - Related goals
4. **Storage** in vector database
5. **Decay scheduling** based on importance

#### Importance Calculation
```
importance = (goal_relevance * 0.4) + (emotional_impact * 0.3) + (novelty * 0.3)
```

High importance memories:
- Persist longer
- Retrieved more readily
- Influence decisions more strongly

## MCP Memory Server Interface

### Core Tools

#### Retrieval Tools
- `recall_episodes(query, time_range?, importance_threshold?)`
  - Search specific events and experiences
  - Returns temporally ordered results

- `recall_facts(topic)`
  - Retrieve learned information
  - Returns confidence-scored facts

- `recall_interactions(character_name)`
  - Get relationship history
  - Returns chronological interaction log

- `check_similar_situations(current_context)`
  - Find analogous past experiences
  - Returns similarity-scored memories

#### Memory Management Tools
- `form_memory(content, importance, type)`
  - Explicitly create new memory
  - Used for significant observations

- `reinforce_memory(memory_id)`
  - Strengthen important memories
  - Prevents decay

- `associate_memories(memory_id_1, memory_id_2)`
  - Create explicit links
  - Improves retrieval

### Response Format
```json
{
  "memories": [
    {
      "content": "Description of memory",
      "timestamp": 1234,
      "importance": 0.8,
      "confidence": 0.9,
      "participants": ["Alice", "Bob"],
      "location": "town square",
      "emotional_valence": -0.3,
      "related_goals": ["survive", "find_allies"]
    }
  ],
  "retrieval_confidence": 0.85,
  "query_expansion_used": ["synonyms", "related_concepts"]
}
```

## Retrieval Strategies

### Query Enhancement
- **Synonym expansion**: Include related terms
- **Context injection**: Use current scene to enhance query
- **Temporal relevance**: Boost recent memories when appropriate
- **Emotional priming**: Consider current emotional state

### Pre-population Strategy
Before each agent turn, automatically retrieve:
- Memories involving visible characters
- Location-specific memories
- Goal-relevant experiences
- Recent high-importance events

This creates a "memory context" that primes decision-making.

## Prompting for Memory Use

### Mandatory Memory Consultation
Agent prompts enforce memory checking:
```
Before deciding your action, you MUST:
1. Check memories for relevant past experiences
2. Recall your history with present characters
3. Consider similar situations you've faced
4. State "no relevant memories" if none found
```

### Memory-Aware Reasoning Chain
```
Observation → Memory Retrieval → Integration → Goal Assessment → Decision
```

### Confidence Gating
```
if uncertainty > threshold:
    required: query_memory()
    required: justify_decision_with_memories()
```

## Memory Decay and Reinforcement

### Decay Model
- Base decay rate varies by importance
- Reinforcement through retrieval
- Emotional memories decay slower
- Goal-critical memories preserved longer

### Reinforcement Triggers
- Memory successfully retrieved
- Memory influences decision
- Emotional resonance with current situation
- Explicit reinforcement via tool

## Semantic Distortion Integration

Memory system supports distortion filters at multiple points:

### Formation Distortions
- Modify memories as they're encoded
- Example: Paranoia adds threatening intent to neutral interactions

### Retrieval Distortions
- Alter memories during retrieval
- Example: Depression dampens positive memories

### Association Distortions
- Change how memories connect
- Example: Trauma creates strong links between triggers

## Shared Memory Pool

### Common Knowledge
- Facts known to all agents
- Environmental constants
- Scenario background information
- Cultural/social norms

### Discovery Propagation
- Information spread between agents
- Conversation-based memory transfer
- Observational learning
- Rumor and misinformation modeling

## Performance Considerations

### Optimization Strategies
- Hierarchical memory organization
- Importance-based indexing
- Temporal partitioning
- Embedding cache for frequent queries

### Scalability
- Memory pruning for long simulations
- Compression of old memories
- Batch retrieval for efficiency
- Async memory operations

## Integration with Agent Architecture

The memory system integrates with agents through:
1. **MCP tool calls** during perception phase
2. **Automatic pre-population** before decisions
3. **Memory formation** after significant events
4. **Distortion filters** when configured

See [Agent Architecture](./agent-architecture.md) for integration details.