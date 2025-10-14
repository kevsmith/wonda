package memory

import "fmt"

// Memory represents a single memory entry with its content and vector embedding.
type Memory struct {
	ID        string            // Unique identifier
	Content   string            // The actual text content
	Embedding []float32         // Vector representation (768d for gtr-t5-base)
	Score     float32           // Relevance score (populated during search)
	Metadata  map[string]string // Structured tags for filtering
}

// Filter defines criteria for filtering memories during search.
type Filter struct {
	Agent    string // Filter by agent name
	Type     string // Filter by memory type (character, scene, episodic, etc.)
	Category string // Filter by category (identity, background, dialogue, etc.)
	About    string // For character_knowledge, who the memory is about
	MinTurn  int    // Minimum turn number (0 = no filter)
	MaxTurn  int    // Maximum turn number (0 = no filter)
}

// Matches returns true if the memory matches all non-empty filter criteria.
func (f *Filter) Matches(m *Memory) bool {
	if f.Agent != "" && m.Metadata["agent"] != f.Agent {
		return false
	}

	if f.Type != "" && m.Metadata["type"] != f.Type {
		return false
	}

	if f.Category != "" && m.Metadata["category"] != f.Category {
		return false
	}

	if f.About != "" && m.Metadata["about"] != f.About {
		return false
	}

	// Turn filtering (if metadata has "turn" field)
	if turnStr, ok := m.Metadata["turn"]; ok {
		// Parse turn number
		var turn int
		if _, err := fmt.Sscanf(turnStr, "%d", &turn); err == nil {
			if f.MinTurn > 0 && turn < f.MinTurn {
				return false
			}
			if f.MaxTurn > 0 && turn > f.MaxTurn {
				return false
			}
		}
	}

	return true
}
