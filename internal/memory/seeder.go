package memory

import (
	"context"
	"fmt"
	"strings"

	"github.com/poiesic/wonda/internal/scenarios"
)

// SeedCharacter pre-seeds the memory store with character knowledge.
// Stores content under multiple canonical queries for reliable retrieval.
func SeedCharacter(ctx context.Context, store *Store, agentName string, char *scenarios.Character) error {
	// Identity: Core character information
	identityContent := buildIdentityContent(char)
	if identityContent != "" {
		identityQueries := []string{
			"who am I?",
			"what is my identity?",
			"describe myself",
		}

		for _, query := range identityQueries {
			embedding, err := store.Embed(ctx, query)
			if err != nil {
				return fmt.Errorf("failed to embed identity query: %w", err)
			}

			store.Add(Memory{
				Content:   identityContent,
				Embedding: embedding,
				Metadata: map[string]string{
					"agent":      agentName,
					"type":       "character",
					"category":   "identity",
					"indexed_by": query,
				},
			})
		}
	}

	// Background: Personal history
	if char.Basics.Background != "" {
		backgroundQueries := []string{
			"what is my background?",
			"what is my history?",
		}

		// Chunk background if it's long
		chunks := ChunkText(char.Basics.Background, 300)

		for _, chunk := range chunks {
			for _, query := range backgroundQueries {
				embedding, err := store.Embed(ctx, query)
				if err != nil {
					return fmt.Errorf("failed to embed background query: %w", err)
				}

				store.Add(Memory{
					Content:   chunk,
					Embedding: embedding,
					Metadata: map[string]string{
						"agent":      agentName,
						"type":       "character",
						"category":   "background",
						"indexed_by": query,
					},
				})
			}
		}
	}

	// Communication Style
	if char.Basics.CommunicationStyle != "" {
		commQueries := []string{
			"how do I communicate?",
			"how do I speak?",
			"what is my communication style?",
		}

		for _, query := range commQueries {
			embedding, err := store.Embed(ctx, query)
			if err != nil {
				return fmt.Errorf("failed to embed communication query: %w", err)
			}

			store.Add(Memory{
				Content:   char.Basics.CommunicationStyle,
				Embedding: embedding,
				Metadata: map[string]string{
					"agent":      agentName,
					"type":       "character",
					"category":   "communication",
					"indexed_by": query,
				},
			})
		}
	}

	// Decision Style
	if char.Basics.DecisionStyle != "" {
		decisionQueries := []string{
			"how do I make decisions?",
			"what is my decision style?",
		}

		for _, query := range decisionQueries {
			embedding, err := store.Embed(ctx, query)
			if err != nil {
				return fmt.Errorf("failed to embed decision query: %w", err)
			}

			store.Add(Memory{
				Content:   char.Basics.DecisionStyle,
				Embedding: embedding,
				Metadata: map[string]string{
					"agent":      agentName,
					"type":       "character",
					"category":   "decision_style",
					"indexed_by": query,
				},
			})
		}
	}

	// Traits
	if len(char.Basics.Traits) > 0 {
		traitsContent := fmt.Sprintf("Your key traits: %s", strings.Join(char.Basics.Traits, ", "))
		traitsQueries := []string{
			"what are my traits?",
			"describe my personality",
		}

		for _, query := range traitsQueries {
			embedding, err := store.Embed(ctx, query)
			if err != nil {
				return fmt.Errorf("failed to embed traits query: %w", err)
			}

			store.Add(Memory{
				Content:   traitsContent,
				Embedding: embedding,
				Metadata: map[string]string{
					"agent":      agentName,
					"type":       "character",
					"category":   "traits",
					"indexed_by": query,
				},
			})
		}
	}

	// Skills
	if len(char.Basics.Skills) > 0 {
		skillsContent := fmt.Sprintf("Your skills: %s", strings.Join(char.Basics.Skills, ", "))
		skillsQueries := []string{
			"what am I good at?",
			"what are my skills?",
		}

		for _, query := range skillsQueries {
			embedding, err := store.Embed(ctx, query)
			if err != nil {
				return fmt.Errorf("failed to embed skills query: %w", err)
			}

			store.Add(Memory{
				Content:   skillsContent,
				Embedding: embedding,
				Metadata: map[string]string{
					"agent":      agentName,
					"type":       "character",
					"category":   "skills",
					"indexed_by": query,
				},
			})
		}
	}

	// Values
	if len(char.Basics.Values) > 0 {
		valuesContent := fmt.Sprintf("Your values: %s", strings.Join(char.Basics.Values, ", "))
		valuesQueries := []string{
			"what do I value?",
			"what are my principles?",
		}

		for _, query := range valuesQueries {
			embedding, err := store.Embed(ctx, query)
			if err != nil {
				return fmt.Errorf("failed to embed values query: %w", err)
			}

			store.Add(Memory{
				Content:   valuesContent,
				Embedding: embedding,
				Metadata: map[string]string{
					"agent":      agentName,
					"type":       "character",
					"category":   "values",
					"indexed_by": query,
				},
			})
		}
	}

	return nil
}

// buildIdentityContent creates a concise identity string from character basics.
func buildIdentityContent(char *scenarios.Character) string {
	parts := make([]string, 0)

	if char.Basics.Archetype != "" {
		parts = append(parts, fmt.Sprintf("You are %s", char.Basics.Archetype))
	}

	if char.Basics.Description != "" {
		parts = append(parts, char.Basics.Description)
	}

	return strings.Join(parts, ". ")
}

// SeedOtherCharacter pre-seeds knowledge about another character.
// Stores what this agent knows about another agent.
func SeedOtherCharacter(ctx context.Context, store *Store, agentName string, targetName string, char *scenarios.Character) error {
	// Build content about the other character
	content := buildIdentityContent(char)
	if content == "" {
		return nil
	}

	// Store under queries about the target
	queries := []string{
		fmt.Sprintf("who is %s?", targetName),
		fmt.Sprintf("what do I know about %s?", targetName),
		fmt.Sprintf("describe %s", targetName),
	}

	for _, query := range queries {
		embedding, err := store.Embed(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to embed character knowledge query: %w", err)
		}

		store.Add(Memory{
			Content:   content,
			Embedding: embedding,
			Metadata: map[string]string{
				"agent":      agentName,
				"type":       "character_knowledge",
				"about":      targetName,
				"indexed_by": query,
			},
		})
	}

	return nil
}

// SeedScenario pre-seeds the memory store with scenario context.
// This information is shared across all agents.
func SeedScenario(ctx context.Context, store *Store, scenario *scenarios.Scenario) error {
	// Location
	if scenario.Basics.Location != "" {
		locationQueries := []string{
			"where am I?",
			"what is the location?",
			"describe the scene",
		}

		for _, query := range locationQueries {
			embedding, err := store.Embed(ctx, query)
			if err != nil {
				return fmt.Errorf("failed to embed location query: %w", err)
			}

			store.Add(Memory{
				Content:   fmt.Sprintf("Location: %s", scenario.Basics.Location),
				Embedding: embedding,
				Metadata: map[string]string{
					"type":       "scene",
					"category":   "location",
					"indexed_by": query,
				},
			})
		}
	}

	// Atmosphere
	if scenario.Basics.Atmosphere != "" {
		atmosphereQueries := []string{
			"what's the atmosphere?",
			"what's the mood?",
			"describe the atmosphere",
		}

		for _, query := range atmosphereQueries {
			embedding, err := store.Embed(ctx, query)
			if err != nil {
				return fmt.Errorf("failed to embed atmosphere query: %w", err)
			}

			store.Add(Memory{
				Content:   fmt.Sprintf("Atmosphere: %s", scenario.Basics.Atmosphere),
				Embedding: embedding,
				Metadata: map[string]string{
					"type":       "scene",
					"category":   "atmosphere",
					"indexed_by": query,
				},
			})
		}
	}

	// Time of Day
	if scenario.Basics.TOD != "" {
		timeQueries := []string{
			"what time is it?",
			"when is this happening?",
		}

		for _, query := range timeQueries {
			embedding, err := store.Embed(ctx, query)
			if err != nil {
				return fmt.Errorf("failed to embed time query: %w", err)
			}

			store.Add(Memory{
				Content:   fmt.Sprintf("Time: %s", scenario.Basics.TOD),
				Embedding: embedding,
				Metadata: map[string]string{
					"type":       "scene",
					"category":   "time",
					"indexed_by": query,
				},
			})
		}
	}

	// Scenario description/context
	if scenario.Basics.Description != "" {
		contextQueries := []string{
			"what is happening?",
			"what's the situation?",
		}

		for _, query := range contextQueries {
			embedding, err := store.Embed(ctx, query)
			if err != nil {
				return fmt.Errorf("failed to embed context query: %w", err)
			}

			store.Add(Memory{
				Content:   scenario.Basics.Description,
				Embedding: embedding,
				Metadata: map[string]string{
					"type":       "scene",
					"category":   "context",
					"indexed_by": query,
				},
			})
		}
	}

	return nil
}
