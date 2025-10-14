package memory

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/google/uuid"
)

// Store manages memory storage and retrieval.
type Store struct {
	memories []Memory
	embedder Embedder
}

// NewStore creates a new memory store with the given embedder.
func NewStore(embedder Embedder) *Store {
	return &Store{
		memories: make([]Memory, 0),
		embedder: embedder,
	}
}

// Add adds a new memory to the store.
func (s *Store) Add(mem Memory) string {
	// Generate ID if not provided
	if mem.ID == "" {
		mem.ID = uuid.New().String()
	}

	// Ensure metadata map exists
	if mem.Metadata == nil {
		mem.Metadata = make(map[string]string)
	}

	s.memories = append(s.memories, mem)
	return mem.ID
}

// Embed generates an embedding for the given text.
func (s *Store) Embed(ctx context.Context, text string) ([]float32, error) {
	return s.embedder.Embed(ctx, text)
}

// Search performs vector similarity search with filtering.
func (s *Store) Search(ctx context.Context, queryEmbedding []float32, filter Filter, topK int) []Memory {
	// 1. Filter by metadata
	candidates := make([]Memory, 0)
	for _, mem := range s.memories {
		if filter.Matches(&mem) {
			candidates = append(candidates, mem)
		}
	}

	if len(candidates) == 0 {
		return []Memory{}
	}

	// 2. Compute cosine similarity scores
	type scoredMemory struct {
		memory Memory
		score  float32
	}

	scored := make([]scoredMemory, len(candidates))
	for i, mem := range candidates {
		score := cosineSimilarity(queryEmbedding, mem.Embedding)
		scored[i] = scoredMemory{
			memory: mem,
			score:  score,
		}
	}

	// 3. Sort by score (highest first)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// 4. Return top K
	resultCount := topK
	if resultCount > len(scored) {
		resultCount = len(scored)
	}

	results := make([]Memory, resultCount)
	for i := 0; i < resultCount; i++ {
		results[i] = scored[i].memory
		results[i].Score = scored[i].score
	}

	return results
}

// SearchByCanonicalQuery searches using a fixed text query.
// This is used for pre-seeded memories indexed under specific queries.
func (s *Store) SearchByCanonicalQuery(ctx context.Context, query string, filter Filter, topK int) ([]Memory, error) {
	// Embed the query
	queryEmbedding, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Search with the embedding
	return s.Search(ctx, queryEmbedding, filter, topK), nil
}

// Count returns the total number of memories in the store.
func (s *Store) Count() int {
	return len(s.memories)
}

// CountByFilter returns the number of memories matching the filter.
func (s *Store) CountByFilter(filter Filter) int {
	count := 0
	for _, mem := range s.memories {
		if filter.Matches(&mem) {
			count++
		}
	}
	return count
}

// cosineSimilarity computes the cosine similarity between two vectors.
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}
