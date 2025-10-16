package scenarios

import (
	"fmt"
	"os"
	"slices"

	"github.com/pelletier/go-toml/v2"
	"github.com/poiesic/wonda/internal/config"
)

type ExternalCharacterInfo struct {
	Archetype          string   `toml:"archetype"`
	Description        string   `toml:"description"`
	CommunicationStyle string   `toml:"communication_style"`
	PositiveTraits     []string `toml:"positive_traits"`
	NegativeTraits     []string `toml:"negative_traits"`
	UniqueSkills       []string `toml:"unique_skills"`
}

type InternalCharacterInfo struct {
	Background    string   `toml:"background"`
	DecisionStyle string   `toml:"decision_style"`
	Secrets       []string `toml:"secrets"`
}

type Character struct {
	External *ExternalCharacterInfo `toml:"external"`
	Internal *InternalCharacterInfo `toml:"internal"`
	Version  string                 `toml:"version"`
}

func NewCharacter() *Character {
	return &Character{
		External: &ExternalCharacterInfo{},
		Internal: &InternalCharacterInfo{},
	}
}

// LoadCharacter creates and populates a Character from TOML data.
func LoadCharacter(data []byte) (*Character, error) {
	c := NewCharacter()
	if err := toml.Unmarshal(data, c); err != nil {
		return nil, err
	}

	// Validate version
	if err := config.ValidateVersion("character", c.Version); err != nil {
		return nil, err
	}

	return c, nil
}

// LoadCharacterFromFile loads a character definition from a file path.
func LoadCharacterFromFile(path string) (*Character, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	character, err := LoadCharacter(data)
	if err != nil {
		return nil, err
	}
	if err := character.Validate(); err != nil {
		return nil, fmt.Errorf("character validation failed: %w", err)
	}
	return character, nil
}

// Validate checks that all required fields are present and valid.
func (c *Character) Validate() error {
	// External validations
	if c.External == nil {
		return fmt.Errorf("external section is required")
	}
	if c.External.Archetype == "" {
		return fmt.Errorf("external.archetype is required")
	}
	if len(c.External.Description) < 10 || len(c.External.Description) > 1000 {
		return fmt.Errorf("external.description must be 10-1000 characters (got %d)", len(c.External.Description))
	}
	if len(c.External.CommunicationStyle) < 10 || len(c.External.CommunicationStyle) > 500 {
		return fmt.Errorf("external.communication_style must be 10-500 characters (got %d)", len(c.External.CommunicationStyle))
	}
	if len(c.External.PositiveTraits) == 0 {
		return fmt.Errorf("external.positive_traits must have at least 1 item")
	}
	if len(c.External.NegativeTraits) == 0 {
		return fmt.Errorf("external.negative_traits must have at least 1 item")
	}

	// Internal validations
	if c.Internal == nil {
		return fmt.Errorf("internal section is required")
	}
	if len(c.Internal.DecisionStyle) < 10 || len(c.Internal.DecisionStyle) > 500 {
		return fmt.Errorf("internal.decision_style must be 10-500 characters (got %d)", len(c.Internal.DecisionStyle))
	}
	if len(c.Internal.Background) > 2000 {
		return fmt.Errorf("internal.background must be at most 2000 characters (got %d)", len(c.Internal.Background))
	}

	return nil
}

func (c *Character) Same(other *Character) bool {
	if c.Version != other.Version {
		return false
	}

	// Compare external fields
	if c.External.Archetype != other.External.Archetype ||
		c.External.Description != other.External.Description ||
		c.External.CommunicationStyle != other.External.CommunicationStyle {
		return false
	}
	if !slices.Equal(c.External.PositiveTraits, other.External.PositiveTraits) ||
		!slices.Equal(c.External.NegativeTraits, other.External.NegativeTraits) ||
		!slices.Equal(c.External.UniqueSkills, other.External.UniqueSkills) {
		return false
	}

	// Compare internal fields
	if c.Internal.Background != other.Internal.Background ||
		c.Internal.DecisionStyle != other.Internal.DecisionStyle {
		return false
	}
	if !slices.Equal(c.Internal.Secrets, other.Internal.Secrets) {
		return false
	}

	return true
}
