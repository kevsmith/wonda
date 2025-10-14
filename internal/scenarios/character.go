package scenarios

import (
	"os"
	"slices"

	"github.com/pelletier/go-toml/v2"
)

type BasicCharacterInformation struct {
	Archetype          string   `toml:"archetype"`
	Description        string   `toml:"description"`
	Background         string   `toml:"background"`
	CommunicationStyle string   `toml:"communication_style"`
	DecisionStyle      string   `toml:"decision_style"`
	Traits             []string `toml:"traits"`
	Skills             []string `toml:"skills"`
	Values             []string `toml:"values"`
}

type Character struct {
	Basics  *BasicCharacterInformation `toml:"basics"`
	Version string                     `toml:"version"`
}

func NewCharacter() *Character {
	return &Character{
		Basics: &BasicCharacterInformation{},
	}
}

// LoadCharacter creates and populates a Character from TOML data.
func LoadCharacter(data []byte) (*Character, error) {
	c := NewCharacter()
	if err := toml.Unmarshal(data, c); err != nil {
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
	return LoadCharacter(data)
}

func (c *Character) Same(other *Character) bool {
	if c.Version != other.Version {
		return false
	}
	if c.Basics.Archetype != other.Basics.Archetype || c.Basics.Description != other.Basics.Description ||
		c.Basics.Background != other.Basics.Background || c.Basics.CommunicationStyle != other.Basics.CommunicationStyle ||
		c.Basics.DecisionStyle != other.Basics.DecisionStyle {
		return false
	}
	if !slices.Equal(c.Basics.Traits, other.Basics.Traits) || !slices.Equal(c.Basics.Skills, other.Basics.Skills) ||
		!slices.Equal(c.Basics.Values, other.Basics.Values) {
		return false
	}
	return true
}
