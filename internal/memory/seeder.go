package memory

import (
	"context"
	"fmt"
	"strings"

	"github.com/poiesic/wonda/internal/scenarios"
)

// SeedCharacter pre-seeds the memory store with character knowledge.
// Only seeds information NOT in the system prompt (background, unique_skills).
// Core identity, traits, communication style, decision style, and secrets
// are provided directly in the agent's system prompt.
func SeedCharacter(ctx context.Context, store *Store, agentName string, char *scenarios.Character) error {
	// Background: Personal history (chunked if long)
	if char.Internal.Background != "" {
		backgroundQueries := []string{
			"what is my background?",
			"what is my history?",
		}

		// Chunk background if it's long
		chunks := ChunkText(char.Internal.Background, 300)

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

	// Unique Skills (when relevant to query)
	if len(char.External.UniqueSkills) > 0 {
		skillsContent := fmt.Sprintf("Your skills: %s", strings.Join(char.External.UniqueSkills, ", "))
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

	return nil
}

// buildExternalIdentity creates identity string from external (observable) character info.
func buildExternalIdentity(targetName string, char *scenarios.Character) string {
	parts := make([]string, 0)

	// Start with the agent's actual name and archetype
	if char.External.Archetype != "" {
		parts = append(parts, fmt.Sprintf("%s is %s", targetName, char.External.Archetype))
	} else {
		parts = append(parts, targetName)
	}

	// Add description (observable)
	if char.External.Description != "" {
		parts = append(parts, char.External.Description)
	}

	// Add communication style (observable when they speak)
	if char.External.CommunicationStyle != "" {
		parts = append(parts, fmt.Sprintf("Communication style: %s", char.External.CommunicationStyle))
	}

	// Add positive traits (observable)
	if len(char.External.PositiveTraits) > 0 {
		parts = append(parts, fmt.Sprintf("Positive traits: %s", strings.Join(char.External.PositiveTraits, ", ")))
	}

	// Add some negative traits (the obvious/observable ones)
	if len(char.External.NegativeTraits) > 0 {
		parts = append(parts, fmt.Sprintf("Notable flaws: %s", strings.Join(char.External.NegativeTraits, ", ")))
	}

	// Add unique skills (when demonstrated)
	if len(char.External.UniqueSkills) > 0 {
		parts = append(parts, fmt.Sprintf("Skills: %s", strings.Join(char.External.UniqueSkills, ", ")))
	}

	return strings.Join(parts, ". ")
}

// SeedOtherCharacter pre-seeds knowledge about another character.
// Only includes external/observable information (not secrets or internal thoughts).
func SeedOtherCharacter(ctx context.Context, store *Store, agentName string, targetName string, char *scenarios.Character) error {
	// Build content about the other character using only external info
	content := buildExternalIdentity(targetName, char)
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
