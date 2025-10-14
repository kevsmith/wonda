# Memory System

**Version**: 1.0.0
**Status**: Implemented (MVP)
**Last Updated**: 2025-10-14

## Overview

The Wonda memory system enables agents to discover their identity and recall experiences through retrieval-augmented generation (RAG). Rather than injecting large character definitions directly into prompts, agents query their memories to understand who they are, where they are, and what has happened.

This "gifted memories" approach is inspired by the Blade Runner concept of implanting memories into replicants - agents are seeded with pre-defined character knowledge and accumulate episodic memories during simulation.

## Design Principles

1. **Discovery over Declaration**: Agents discover their identity through memory queries rather than receiving it in system prompts
2. **Efficiency**: Only retrieve relevant context instead of passing entire character definitions every turn
3. **Realism**: Agents must actively query memories, creating natural forgetting when they don't
4. **In-Process Performance**: Leverage in-process storage to minimize RAG latency
5. **Creative Tool, Not Content Generator**: Memory supports the chronicle as raw creative material for writers to refine

## Key Concepts

**Pre-seeding** (✓ Implemented): Loading memories into the store at simulation initialization. Character definitions, scenario context, and inter-character knowledge are embedded and stored before any agent takes a turn.

**Pre-population** (✗ Not Implemented): Automatically retrieving relevant memories and injecting them into each agent's prompt context before they decide what to do. Currently, agents must explicitly call MCP tools like `query_self()` or `query_scene()` to access even their own identity.

**Episodic Capture** (✓ Implemented): Automatically storing dialogue as memories after each turn so agents can recall what was said earlier in the simulation.

## Memory Types

### Pre-Seeded Memories (Gifted) ✓ Implemented

**Character Memories** - Loaded at simulation initialization from character definitions:
- **Identity**: Name, role, archetype, description
- **Background**: Personal history, formative experiences
- **Communication**: Speech patterns and interaction style
- **Decision Style**: How the agent makes choices
- **Traits**: Personality characteristics
- **Skills**: Capabilities and expertise
- **Values**: Beliefs and principles

**Scenario Context** - Shared across all agents:
- **Location**: Where the scene takes place
- **Atmosphere**: Emotional/social ambiance
- **Time**: Time of day, temporal context
- **Context**: Situation and circumstances

**Character Knowledge** - What each agent knows about other agents:
- Identity and archetype of other characters
- Observable characteristics
- Known relationships

### Episodic Memories (Earned) ✓ Implemented (Partial)

**Dialogue** - Captured automatically during simulation:
- Everything said by all agents
- Timestamped with turn number
- Indexed by speaker and content
- Searchable through flexible semantic queries

**Future Types** (not yet implemented):
- Observations from perception tools
- Actions taken (proposals, votes, movements)
- Goal progress updates
- Emotional responses to events

## Architecture

### Memory Storage

```go
type Memory struct {
    ID        string              // Unique identifier
    Content   string              // The actual text content
    Embedding []float32           // Vector representation (768d)
    Score     float32             // Relevance score (populated during search)
    Metadata  map[string]string   // Type, agent, category, turn, etc.
}

type Store struct {
    memories  []Memory
    embedder  Embedder
}
```

**Storage Characteristics:**
- In-memory only (no persistence between runs)
- Simple slice-based storage
- No external dependencies (SQLite, vector DB, etc.)

### Embedding Model

**Current Model**: `nomic-ai/nomic-embed-text-v1.5-GGUF`

**Specifications**:
- 768 dimensions
- 8192 token context length
- Available via LM Studio and Ollama
- Fast local generation

**Provider Support**:
- Ollama format: `POST /api/embeddings` with `prompt` field
- OpenAI format: `POST /v1/embeddings` with `input` field
- Automatic detection and fallback between formats

**Note**: The current model (nomic-embed-text) is NOT vec2text compatible. For future cognitive distortion features, we may need to switch to `sentence-transformers/gtr-t5-base`.

### Validation

At simulation startup:
1. Check provider availability
2. Validate embedding model exists
3. Test embedding generation
4. Verify embedding dimensions (must be 768)

Startup fails with clear error message if validation fails, including instructions to install the model.

## Memory Seeding

### Canonical Query Pattern

Pre-seeded memories are stored multiple times under different query embeddings that represent the same semantic intent. This ensures reliable retrieval even when agents phrase questions differently.

```go
// Example: Identity content stored under multiple query vectors
content := "You are a pragmatic project manager..."

queries := []string{
    "who am I?",
    "what is my identity?",
    "describe myself",
}

for _, query := range queries {
    embedding := embedder.Embed(query)
    store.Add(Memory{
        Content: content,
        Embedding: embedding,
        Metadata: {
            "agent": "Alex",
            "type": "character",
            "category": "identity",
            "indexed_by": query,
        },
    })
}
```

### Character Memory Seeding

Each agent receives pre-seeded memories for their character:

| Category | Canonical Queries | Content Source |
|----------|------------------|----------------|
| Identity | "who am I?", "what is my identity?", "describe myself" | Archetype + Description |
| Background | "what is my background?", "what is my history?" | Background field (chunked if >300 chars) |
| Communication | "how do I communicate?", "how do I speak?", "what is my communication style?" | CommunicationStyle field |
| Decision Style | "how do I make decisions?", "what is my decision style?" | DecisionStyle field |
| Traits | "what are my traits?", "describe my personality" | Traits list |
| Skills | "what am I good at?", "what are my skills?" | Skills list |
| Values | "what do I value?", "what are my principles?" | Values list |

### Scenario Context Seeding

Scenario information is seeded once and available to all agents:

| Category | Canonical Queries | Content Source |
|----------|------------------|----------------|
| Location | "where am I?", "what is the location?", "describe the scene" | Scenario.Location |
| Atmosphere | "what's the atmosphere?", "what's the mood?", "describe the atmosphere" | Scenario.Atmosphere |
| Time | "what time is it?", "when is this happening?" | Scenario.TOD |
| Context | "what is happening?", "what's the situation?" | Scenario.Description |

### Other Character Seeding

For each agent, knowledge about other agents is pre-seeded:

```go
queries := []string{
    "who is {name}?",
    "what do I know about {name}?",
    "describe {name}",
}
```

Content includes the other character's identity (archetype + description).

### Episodic Memory Capture

Dialogue is automatically captured after each agent's turn:

```go
content := fmt.Sprintf("%s said: %s", agentName, dialogue)
embedding := store.Embed(content)

store.Add(Memory{
    Content: content,
    Embedding: embedding,
    Metadata: {
        "type": "episodic",
        "category": "dialogue",
        "turn": turnNumber,
        "speaker": agentName,
    },
})
```

## MCP Tool Interface

Agents access memories through MCP tools during their turns.

### Fixed-Query Tools (Character Knowledge)

These tools use fixed canonical queries for reliable retrieval:

**`query_self()`**
- Description: "Retrieve your core identity - who you are, your personality, background"
- Canonical query: "who am I?"
- Filter: `{agent: self, type: "character", category: "identity"}`
- Returns: Top 5 identity memories

**`query_background()`**
- Description: "Retrieve your personal history and background"
- Canonical query: "what is my background?"
- Filter: `{agent: self, type: "character", category: "background"}`
- Returns: Top 5 background memories

**`query_communication_style()`**
- Description: "Learn how you communicate and interact with others"
- Canonical query: "how do I communicate?"
- Filter: `{agent: self, type: "character", category: "communication"}`
- Returns: Top 3 communication memories

**`query_scene()`**
- Description: "Understand where you are and the current atmosphere"
- Canonical query: "where am I?"
- Filter: `{type: "scene"}`
- Returns: Top 5 scene memories (location, atmosphere, time, context)

### Parameterized Query Tools

**`query_character(name: string)`**
- Description: "Learn about another agent in the simulation"
- Query: `"who is {name}?"`
- Filter: `{agent: self, type: "character_knowledge", about: name}`
- Returns: Top 3 memories about the specified character

### Flexible Query Tools (Episodic Memory)

**`query_memory(query: string)`**
- Description: "Search your memories of what has happened during the simulation"
- Query: User-provided (e.g., "what did Alice say about restaurants?")
- Filter: `{type: "episodic"}`
- Returns: Top 5 semantically relevant episodic memories with turn numbers

### Response Format

All memory tools return structured responses:

```json
{
  "memories": [
    {
      "content": "Memory text...",
      "relevance": 0.87
    },
    {
      "content": "Another memory...",
      "relevance": 0.82
    }
  ]
}
```

For episodic memories, turn numbers are also included:

```json
{
  "query": "what did Alice say about restaurants?",
  "memories": [
    {
      "content": "Alice said: I love Italian food...",
      "relevance": 0.91,
      "turn": "3"
    }
  ]
}
```

## Vector Search

### Search Algorithm

The current implementation uses brute-force cosine similarity:

```go
func (s *Store) Search(queryEmbedding []float32, filter Filter, topK int) []Memory {
    // 1. Filter candidates by metadata
    candidates := []Memory{}
    for _, mem := range s.memories {
        if filter.Matches(&mem) {
            candidates = append(candidates, mem)
        }
    }

    // 2. Compute cosine similarity for all candidates
    scores := []struct{idx int; score float32}{}
    for i, mem := range candidates {
        scores[i] = {
            idx: i,
            score: cosineSimilarity(queryEmbedding, mem.Embedding),
        }
    }

    // 3. Sort by score (highest first)
    sort.Slice(scores, func(i, j int) bool {
        return scores[i].score > scores[j].score
    })

    // 4. Return top K
    results := make([]Memory, min(topK, len(scores)))
    for i := 0; i < len(results); i++ {
        results[i] = candidates[scores[i].idx]
        results[i].Score = scores[i].score
    }

    return results
}
```

**Cosine Similarity**:
```go
func cosineSimilarity(a, b []float32) float32 {
    var dotProduct, normA, normB float32
    for i := range a {
        dotProduct += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }
    return dotProduct / (sqrt(normA) * sqrt(normB))
}
```

### Performance Characteristics

**Brute-force approach**:
- Time complexity: O(n × d) where n=memories, d=dimensions
- For 1,000 memories × 768 dimensions: ~1-5ms on modern CPU
- For 10,000 memories: ~10-50ms (acceptable for MVP)
- No indexing overhead
- Simple and predictable

**When to optimize**:
- If memory count exceeds 10,000
- If per-turn latency becomes noticeable (>100ms)
- Consider HNSW indexing or product quantization

## Agent Prompt Strategy

### Minimal System Prompt

Instead of large character definitions, agents receive minimal prompts that encourage memory tool usage:

```
You are an agent in a simulation. You have access to memory tools:

Character Knowledge (discover who you are):
- query_self(): Your core identity and personality
- query_background(): Your personal history
- query_communication_style(): How you speak and interact

Situational Awareness:
- query_scene(): Understand where you are
- query_character(name): Learn about other agents

Memory:
- query_memory(question): Recall what has happened

Use these tools to understand yourself and your situation before acting.
```

**Token Efficiency**:
- Old approach: ~2000-3000 tokens per turn (full character definition)
- New approach: ~100-200 tokens base + ~200-500 tokens retrieved content
- Net savings: ~50-80% reduction in input tokens

### Agent Behavior Pattern

**Turn 1 (Discovery)**:
```
Agent: <tool_call>query_self()</tool_call>
→ Returns: "You are Alex, a pragmatic project manager..."

Agent: <tool_call>query_scene()</tool_call>
→ Returns: "Location: Alex's apartment, living room..."

Agent: Hi everyone, I'm Alex. Should we start discussing dinner plans?
```

**Turn 3 (Recall)**:
```
Agent: <tool_call>query_memory("what did Jordan say about food?")</tool_call>
→ Returns: "Jordan said: I love Italian food..."

Agent: Given your preference for Italian, how about we try Bella's?
```

## Metadata Filtering

The `Filter` struct enables precise memory selection:

```go
type Filter struct {
    Agent    string // Filter by agent name
    Type     string // character, scene, episodic, character_knowledge
    Category string // identity, background, dialogue, etc.
    About    string // For character_knowledge - target agent name
    MinTurn  int    // Temporal filtering (0 = no filter)
    MaxTurn  int    // Temporal filtering (0 = no filter)
}
```

**Examples**:
```go
// Get agent's own identity
Filter{Agent: "Alex", Type: "character", Category: "identity"}

// Get recent dialogue (last 5 turns)
Filter{Type: "episodic", Category: "dialogue", MinTurn: currentTurn-5}

// Get what Alex knows about Jordan
Filter{Agent: "Alex", Type: "character_knowledge", About: "Jordan"}

// Get all scene context
Filter{Type: "scene"}
```

## Text Chunking

Long text fields (e.g., background) are chunked before embedding to stay within token limits and improve retrieval precision.

**Strategy**: Sentence-based chunking with maximum character limit (default 300 chars)

```go
func ChunkText(text string, maxChars int) []string {
    // 1. Split into sentences at .!? boundaries
    // 2. Accumulate sentences into chunks up to maxChars
    // 3. Start new chunk when adding next sentence would exceed limit
    // 4. Return array of chunks
}
```

Each chunk is stored as a separate memory with the same metadata, allowing fine-grained retrieval.

## Integration with Simulation

### Initialization Flow

```
┌─────────────────────────────────────────────────────────┐
│                  Simulation.Initialize()                 │
└────────────────────┬────────────────────────────────────┘
                     │
                     ↓
         ┌───────────────────────────┐
         │   Validate Embeddings     │
         │   (nomic-embed-text)      │
         └───────────┬───────────────┘
                     │
                     ↓
         ┌───────────────────────────┐
         │   Initialize Memory Store │
         │   (with OllamaEmbedder)   │
         └───────────┬───────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
        ↓                         ↓
┌──────────────────┐    ┌──────────────────┐
│ SeedScenario()   │    │ For each agent:  │
│ - Location       │    │ - SeedCharacter()│
│ - Atmosphere     │    │ - SeedOther      │
│ - Time           │    │   Characters()   │
│ - Context        │    │                  │
└────────┬─────────┘    └────────┬─────────┘
         │                       │
         └───────────┬───────────┘
                     │
                     ↓
         ┌───────────────────────────┐
         │ Register MCP Memory Tools │
         │ - query_self              │
         │ - query_background        │
         │ - query_communication     │
         │ - query_scene             │
         │ - query_character         │
         │ - query_memory            │
         └───────────────────────────┘
```

### Turn Execution Flow

```
┌─────────────────────────────────────────────────────────┐
│                     Agent Turn                           │
└────────────────────┬────────────────────────────────────┘
                     │
                     ↓
         ┌───────────────────────────┐
         │   Agent Decision Phase    │
         │   - LLM calls tools       │
         │   - query_self()          │
         │   - query_scene()         │
         │   - query_memory()        │
         └───────────┬───────────────┘
                     │
                     ↓
         ┌───────────────────────────┐
         │   Vector Search           │
         │   - Embed query           │
         │   - Filter candidates     │
         │   - Cosine similarity     │
         │   - Return top K          │
         └───────────┬───────────────┘
                     │
                     ↓
         ┌───────────────────────────┐
         │   Agent Takes Action      │
         │   (speaks, proposes, etc.)│
         └───────────┬───────────────┘
                     │
                     ↓
         ┌───────────────────────────┐
         │   Capture Episodic Memory │
         │   - Format dialogue        │
         │   - Embed content         │
         │   - Store with turn #     │
         └───────────────────────────┘
```

## Future Enhancements

The following features are designed but not yet implemented:

### Memory Importance and Decay

**Importance Scoring**:
```
importance = (goal_relevance × 0.4) + (emotional_impact × 0.3) + (novelty × 0.3)
```

**Decay Model**:
- Base decay rate varies by importance
- Reinforcement through retrieval
- Emotional memories decay slower
- Goal-critical memories preserved longer

**Reinforcement Triggers**:
- Memory successfully retrieved
- Memory influences decision
- Emotional resonance with current situation

### Extended Memory Types

**Emotional Memory**:
- Feelings associated with entities/events
- Emotional responses to stimuli
- Trust/fear associations
- Example: "The warehouse makes me anxious"

**Procedural Memory**:
- How to accomplish tasks
- Successful action sequences
- Learned strategies
- Example: "Negotiating with Bob requires appealing to greed"

**Semantic Memory**:
- Facts and knowledge learned during simulation
- Relationships between concepts
- World state understanding
- Example: "The red door leads to the basement"

### Advanced MCP Tools

**Memory Formation**:
- `form_memory(content, importance, type)` - Explicitly create memory
- `reinforce_memory(memory_id)` - Strengthen important memories
- `associate_memories(id1, id2)` - Create explicit links

**Enhanced Retrieval**:
- `recall_episodes(query, time_range?, importance_threshold?)` - Temporally ordered events
- `recall_facts(topic)` - Confidence-scored knowledge
- `recall_interactions(character)` - Relationship history
- `check_similar_situations(context)` - Analogous past experiences

### Retrieval Enhancements

**Query Enhancement**:
- Synonym expansion for related terms
- Context injection from current scene
- Temporal relevance boosting
- Emotional priming based on agent state

**Automatic Pre-population**:
Before each agent turn, automatically retrieve and inject relevant memories into the prompt context:
- Memories involving visible characters
- Location-specific memories
- Goal-relevant experiences
- Recent high-importance events

This would create a "memory context" that primes decision-making without requiring explicit tool calls.

**Note**: Currently agents must explicitly call memory tools (`query_self()`, `query_scene()`, etc.) to retrieve information. Automatic pre-population would reduce the cognitive load on agents and improve response time by eliminating these tool calls.

### Semantic Distortion Filters

Memory system extension points for cognitive distortions:

**Formation Distortions**:
- Modify memories as they're encoded
- Example: Paranoia adds threatening intent to neutral interactions

**Retrieval Distortions**:
- Alter memories during retrieval
- Example: Depression dampens positive memories

**Association Distortions**:
- Change how memories connect
- Example: Trauma creates strong links between triggers

**Implementation** (requires vec2text-compatible embeddings):
```go
type DistortionFilter interface {
    Apply(text string) string
}

func (s *Store) SearchWithDistortion(query string, filter DistortionFilter) []Memory {
    embedding := s.Embed(query)
    results := s.Search(embedding, ...)

    // Invert vectors to text via vec2text
    for i, mem := range results {
        text := s.inverter.Invert(mem.Embedding)
        distortedText := filter.Apply(text)
        results[i].Content = distortedText
    }

    return results
}
```

### Performance Optimizations

**HNSW Indexing** (for >100k memories):
- Hierarchical Navigable Small World index
- O(log n) search instead of O(n)
- Trade-off: build/update overhead

**Memory Pruning**:
- Compression of old, low-importance memories
- Batch operations for efficiency
- Async memory operations

**Query Alias System**:
Store content once, index under multiple query vectors:
```go
type QueryAlias struct {
    QueryEmbedding []float32
    TargetIDs      []string  // → Memory IDs
}
```

### Persistence

**Disk Storage**:
```go
func (s *Store) SaveToFile(path string) error
func (s *Store) LoadFromFile(path string) error
```

**SQLite Integration**:
- Use SQLite with vector extension
- Persist memories between simulation runs
- Enable multi-session character continuity

### Prompt Integration

**Mandatory Memory Consultation**:
```
Before deciding your action, you MUST:
1. Check memories for relevant past experiences
2. Recall your history with present characters
3. Consider similar situations you've faced
4. State "no relevant memories" if none found
```

**Memory-Aware Reasoning Chain**:
```
Observation → Memory Retrieval → Integration → Goal Assessment → Decision
```

**Confidence Gating**:
```
if uncertainty > threshold:
    required: query_memory()
    required: justify_decision_with_memories()
```

## Testing Strategy

### Unit Tests

- `embedder_test.go`: Ollama/LM Studio integration, embedding generation
- `chunker_test.go`: Text chunking logic, sentence splitting
- `store_test.go`: Vector search, cosine similarity, filtering
- `seeder_test.go`: Character/scenario seeding, canonical queries

### Integration Tests

- Full memory retrieval flow (seed → query → retrieve)
- Canonical query reliability (same content retrieved regardless of phrasing)
- Episodic capture across multiple turns
- Metadata filtering accuracy

### End-to-End Tests

- Run dinner-planning scenario with memory system
- Verify agents discover identity on turn 1
- Verify agents recall episodic memories
- Compare token usage vs baseline
- Verify behavior quality maintained

## Performance Targets

**Current MVP Performance**:
- Embedding generation: <100ms per call (via LM Studio/Ollama)
- Vector search: <10ms for <10k memories
- Memory seeding: <1s for full character/scenario
- Per-turn overhead: <200ms total memory operations

**Optimization Threshold**:
- Consider HNSW indexing if memory count exceeds 10,000
- Consider caching if embedding generation exceeds 200ms consistently

## References

- [Agent Architecture](./agent-architecture.md) - How agents integrate with memory
- [MCP Interface](./mcp-interface.md) - MCP protocol details
- [Simulation Execution](./simulation-execution.md) - Simulation lifecycle
- [Character Definition](./character-definition.md) - Character data structure
- [Scenario Definition](./scenario-definition.md) - Scenario data structure

## Decision Log

**2025-01-14**: Decided on simple memory duplication (store under multiple query vectors) instead of alias system for MVP simplicity.

**2025-01-14**: Fixed canonical queries for character knowledge tools to solve semantic mismatch between questions and declarative content.

**2025-01-14**: Initially planned to use gtr-t5-base for vec2text compatibility, but implemented with nomic-embed-text for easier MVP deployment via LM Studio.

**2025-01-14**: In-memory storage for MVP, defer persistence to future phase.

**2025-10-14**: Merged memory-architecture.md and rag-memory-system.md into single coherent document reflecting actual implementation.
