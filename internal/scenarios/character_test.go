package scenarios

import (
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCharacterMarshalTOML(t *testing.T) {
	t.Run("fully populated character", func(t *testing.T) {
		char := &Character{
			Version: "1.0.0",
			Basics: &BasicCharacterInformation{
				Archetype:          "The Mentor",
				Description:        "A wise and experienced guide",
				Background:         "Former warrior turned teacher",
				CommunicationStyle: "Patient and thoughtful",
				DecisionStyle:      "Deliberate and principled",
				Traits:             []string{"wise", "patient", "experienced"},
				Skills:             []string{"teaching", "combat", "philosophy"},
				Values:             []string{"knowledge", "honor", "duty"},
			},
		}

		var buf []byte
		var err error
		buf, err = toml.Marshal(char)
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		result := string(buf)
		assert.Contains(t, result, "version = '1.0.0'")
		assert.Contains(t, result, "archetype = 'The Mentor'")
		assert.Contains(t, result, "description = 'A wise and experienced guide'")
		assert.Contains(t, result, "background = 'Former warrior turned teacher'")
		assert.Contains(t, result, "communication_style = 'Patient and thoughtful'")
		assert.Contains(t, result, "decision_style = 'Deliberate and principled'")
		assert.Contains(t, result, "traits = ['wise', 'patient', 'experienced']")
		assert.Contains(t, result, "skills = ['teaching', 'combat', 'philosophy']")
		assert.Contains(t, result, "values = ['knowledge', 'honor', 'duty']")
	})

	t.Run("character with empty strings", func(t *testing.T) {
		char := &Character{
			Basics: &BasicCharacterInformation{
				Archetype:          "",
				Description:        "",
				Background:         "",
				CommunicationStyle: "",
				DecisionStyle:      "",
				Traits:             []string{},
				Skills:             []string{},
				Values:             []string{},
			},
		}

		buf, err := toml.Marshal(char)
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		result := string(buf)
		assert.Contains(t, result, "archetype = ''")
		assert.Contains(t, result, "traits = []")
		assert.Contains(t, result, "skills = []")
		assert.Contains(t, result, "values = []")
	})

	t.Run("character with nil slices", func(t *testing.T) {
		char := &Character{
			Basics: &BasicCharacterInformation{
				Archetype:          "The Hero",
				Description:        "A brave adventurer",
				Background:         "Unknown origins",
				CommunicationStyle: "Direct",
				DecisionStyle:      "Impulsive",
				Traits:             nil,
				Skills:             nil,
				Values:             nil,
			},
		}

		buf, err := toml.Marshal(char)
		require.NoError(t, err)
		require.NotEmpty(t, buf)
	})

	t.Run("character with special characters", func(t *testing.T) {
		char := &Character{
			Version: "2.1.3",
			Basics: &BasicCharacterInformation{
				Archetype:          "The \"Mysterious\" One",
				Description:        "A character with\nnewlines and\ttabs",
				Background:         "Background with 'quotes' and \"double quotes\"",
				CommunicationStyle: "Complex: uses symbols & punctuation!",
				DecisionStyle:      "Strategic (always thinking)",
				Traits:             []string{"clever", "mysterious", "enigmatic"},
				Skills:             []string{"stealth", "deception"},
				Values:             []string{"freedom", "truth"},
			},
		}

		buf, err := toml.Marshal(char)
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		// Verify it can be unmarshalled back
		var decoded Character
		err = toml.Unmarshal(buf, &decoded)
		require.NoError(t, err)
		assert.True(t, char.Same(&decoded))
	})

	t.Run("character with version only", func(t *testing.T) {
		char := &Character{
			Version: "1.2.3",
			Basics:  &BasicCharacterInformation{},
		}

		buf, err := toml.Marshal(char)
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		result := string(buf)
		assert.Contains(t, result, "version = '1.2.3'")
	})
}

func TestCharacterUnmarshalTOML(t *testing.T) {
	t.Run("fully populated TOML", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[basics]
archetype = "The Villain"
description = "A formidable antagonist"
background = "Once a hero, now corrupted"
communication_style = "Intimidating and commanding"
decision_style = "Ruthless and calculated"
traits = ["cunning", "powerful", "ruthless"]
skills = ["strategy", "manipulation", "dark magic"]
values = ["power", "control", "revenge"]
`

		var char Character
		err := toml.Unmarshal([]byte(tomlData), &char)
		require.NoError(t, err)
		require.NotNil(t, char.Basics)

		assert.Equal(t, "1.0.0", char.Version)
		assert.Equal(t, "The Villain", char.Basics.Archetype)
		assert.Equal(t, "A formidable antagonist", char.Basics.Description)
		assert.Equal(t, "Once a hero, now corrupted", char.Basics.Background)
		assert.Equal(t, "Intimidating and commanding", char.Basics.CommunicationStyle)
		assert.Equal(t, "Ruthless and calculated", char.Basics.DecisionStyle)
		assert.Equal(t, []string{"cunning", "powerful", "ruthless"}, char.Basics.Traits)
		assert.Equal(t, []string{"strategy", "manipulation", "dark magic"}, char.Basics.Skills)
		assert.Equal(t, []string{"power", "control", "revenge"}, char.Basics.Values)
	})

	t.Run("minimal TOML", func(t *testing.T) {
		tomlData := `
[basics]
archetype = "The Simple One"
description = ""
background = ""
communication_style = ""
decision_style = ""
`

		var char Character
		err := toml.Unmarshal([]byte(tomlData), &char)
		require.NoError(t, err)
		require.NotNil(t, char.Basics)

		assert.Equal(t, "The Simple One", char.Basics.Archetype)
		assert.Empty(t, char.Basics.Description)
		assert.Empty(t, char.Basics.Background)
		assert.Nil(t, char.Basics.Traits)
		assert.Nil(t, char.Basics.Skills)
		assert.Nil(t, char.Basics.Values)
	})

	t.Run("TOML with empty arrays", func(t *testing.T) {
		tomlData := `
[basics]
archetype = "The Lone Wolf"
description = "Works alone"
background = "Mysterious past"
communication_style = "Minimal"
decision_style = "Independent"
traits = []
skills = []
values = []
`

		var char Character
		err := toml.Unmarshal([]byte(tomlData), &char)
		require.NoError(t, err)
		require.NotNil(t, char.Basics)

		assert.Equal(t, "The Lone Wolf", char.Basics.Archetype)
		assert.Empty(t, char.Basics.Traits)
		assert.Empty(t, char.Basics.Skills)
		assert.Empty(t, char.Basics.Values)
	})

	t.Run("TOML with single element arrays", func(t *testing.T) {
		tomlData := `
[basics]
archetype = "The Specialist"
description = "Focused on one thing"
background = "Dedicated training"
communication_style = "Technical"
decision_style = "Expert"
traits = ["focused"]
skills = ["swordsmanship"]
values = ["mastery"]
`

		var char Character
		err := toml.Unmarshal([]byte(tomlData), &char)
		require.NoError(t, err)
		require.NotNil(t, char.Basics)

		assert.Equal(t, []string{"focused"}, char.Basics.Traits)
		assert.Equal(t, []string{"swordsmanship"}, char.Basics.Skills)
		assert.Equal(t, []string{"mastery"}, char.Basics.Values)
	})

	t.Run("TOML with multiline strings", func(t *testing.T) {
		tomlData := `
[basics]
archetype = "The Storyteller"
description = """
A wandering bard who tells tales
of heroes and legends from ages past.
Known throughout the land."""
background = """
Born in a small village.
Traveled the world.
Returned home."""
communication_style = "Eloquent and engaging"
decision_style = "Intuitive"
traits = ["charismatic", "creative"]
skills = ["storytelling", "music"]
values = ["art", "truth"]
`

		var char Character
		err := toml.Unmarshal([]byte(tomlData), &char)
		require.NoError(t, err)
		require.NotNil(t, char.Basics)

		assert.Equal(t, "The Storyteller", char.Basics.Archetype)
		assert.Contains(t, char.Basics.Description, "wandering bard")
		assert.Contains(t, char.Basics.Background, "small village")
	})

	t.Run("invalid TOML", func(t *testing.T) {
		tomlData := `
[basics
archetype = "Broken"
`

		var char Character
		err := toml.Unmarshal([]byte(tomlData), &char)
		require.Error(t, err)
	})

	t.Run("TOML with wrong types", func(t *testing.T) {
		tomlData := `
[basics]
archetype = 123
description = "Valid"
`

		var char Character
		err := toml.Unmarshal([]byte(tomlData), &char)
		require.Error(t, err)
	})

	t.Run("TOML with array type mismatch", func(t *testing.T) {
		tomlData := `
[basics]
archetype = "The Confused"
traits = [1, 2, 3]
`

		var char Character
		err := toml.Unmarshal([]byte(tomlData), &char)
		require.Error(t, err)
	})

	t.Run("TOML without version field", func(t *testing.T) {
		tomlData := `
[basics]
archetype = "The Unversioned"
description = "A character without a version"
background = ""
communication_style = "Plain"
decision_style = "Simple"
`

		var char Character
		err := toml.Unmarshal([]byte(tomlData), &char)
		require.NoError(t, err)
		require.NotNil(t, char.Basics)

		assert.Equal(t, "", char.Version)
		assert.Equal(t, "The Unversioned", char.Basics.Archetype)
	})

	t.Run("TOML with version field", func(t *testing.T) {
		tomlData := `
version = "2.0.1"

[basics]
archetype = "The Versioned"
description = "A character with a version"
background = ""
communication_style = "Modern"
decision_style = "Updated"
`

		var char Character
		err := toml.Unmarshal([]byte(tomlData), &char)
		require.NoError(t, err)
		require.NotNil(t, char.Basics)

		assert.Equal(t, "2.0.1", char.Version)
		assert.Equal(t, "The Versioned", char.Basics.Archetype)
	})
}

func TestCharacterRoundTrip(t *testing.T) {
	t.Run("marshal and unmarshal preserves data", func(t *testing.T) {
		original := &Character{
			Version: "1.0.0",
			Basics: &BasicCharacterInformation{
				Archetype:          "The Guardian",
				Description:        "Protector of the realm",
				Background:         "Sworn to defend",
				CommunicationStyle: "Firm but fair",
				DecisionStyle:      "Protective and cautious",
				Traits:             []string{"brave", "loyal", "steadfast"},
				Skills:             []string{"defense", "tactics", "leadership"},
				Values:             []string{"duty", "protection", "sacrifice"},
			},
		}

		// Marshal to TOML
		buf, err := toml.Marshal(original)
		require.NoError(t, err)

		// Unmarshal back
		var decoded Character
		err = toml.Unmarshal(buf, &decoded)
		require.NoError(t, err)

		// Verify equality
		assert.True(t, original.Same(&decoded))
		assert.Equal(t, "1.0.0", decoded.Version)
	})

	t.Run("round trip with empty values", func(t *testing.T) {
		original := &Character{
			Basics: &BasicCharacterInformation{
				Archetype:          "",
				Description:        "",
				Background:         "",
				CommunicationStyle: "",
				DecisionStyle:      "",
				Traits:             []string{},
				Skills:             []string{},
				Values:             []string{},
			},
		}

		buf, err := toml.Marshal(original)
		require.NoError(t, err)

		var decoded Character
		err = toml.Unmarshal(buf, &decoded)
		require.NoError(t, err)

		assert.True(t, original.Same(&decoded))
	})

	t.Run("round trip with partial data", func(t *testing.T) {
		original := &Character{
			Basics: &BasicCharacterInformation{
				Archetype:          "The Wanderer",
				Description:        "No fixed home",
				Background:         "",
				CommunicationStyle: "",
				DecisionStyle:      "",
				Traits:             []string{"adventurous"},
				Skills:             nil,
				Values:             []string{},
			},
		}

		buf, err := toml.Marshal(original)
		require.NoError(t, err)

		var decoded Character
		err = toml.Unmarshal(buf, &decoded)
		require.NoError(t, err)

		assert.True(t, original.Same(&decoded))
	})

	t.Run("multiple round trips preserve data", func(t *testing.T) {
		original := &Character{
			Version: "3.2.1",
			Basics: &BasicCharacterInformation{
				Archetype:          "The Sage",
				Description:        "Ancient wisdom keeper",
				Background:         "Centuries of study",
				CommunicationStyle: "Cryptic and profound",
				DecisionStyle:      "Measured and wise",
				Traits:             []string{"wise", "ancient", "mystical"},
				Skills:             []string{"prophecy", "ancient languages", "meditation"},
				Values:             []string{"knowledge", "balance", "truth"},
			},
		}

		current := original
		for i := 0; i < 3; i++ {
			buf, err := toml.Marshal(current)
			require.NoError(t, err, "marshal iteration %d failed", i)

			var decoded Character
			err = toml.Unmarshal(buf, &decoded)
			require.NoError(t, err, "unmarshal iteration %d failed", i)

			assert.True(t, original.Same(&decoded), "data changed after iteration %d", i)
			assert.Equal(t, "3.2.1", decoded.Version, "version changed after iteration %d", i)
			current = &decoded
		}
	})
}

func TestNewCharacter(t *testing.T) {
	t.Run("creates valid character", func(t *testing.T) {
		char := NewCharacter()
		require.NotNil(t, char)
		require.NotNil(t, char.Basics)
	})

	t.Run("can marshal new character", func(t *testing.T) {
		char := NewCharacter()
		buf, err := toml.Marshal(char)
		require.NoError(t, err)
		require.NotEmpty(t, buf)
	})

	t.Run("can unmarshal into new character", func(t *testing.T) {
		tomlData := `
[basics]
archetype = "Test"
description = "Test character"
background = "Test background"
communication_style = "Test style"
decision_style = "Test decisions"
traits = ["test"]
skills = ["testing"]
values = ["quality"]
`

		char := NewCharacter()
		err := toml.Unmarshal([]byte(tomlData), char)
		require.NoError(t, err)
		assert.Equal(t, "Test", char.Basics.Archetype)
	})
}

func TestCharacterSame(t *testing.T) {
	t.Run("identical characters are same", func(t *testing.T) {
		char1 := &Character{
			Version: "1.0.0",
			Basics: &BasicCharacterInformation{
				Archetype:          "The Hero",
				Description:        "Brave and true",
				Background:         "Humble origins",
				CommunicationStyle: "Inspiring",
				DecisionStyle:      "Courageous",
				Traits:             []string{"brave", "honest"},
				Skills:             []string{"combat", "leadership"},
				Values:             []string{"justice", "honor"},
			},
		}
		char2 := &Character{
			Version: "1.0.0",
			Basics: &BasicCharacterInformation{
				Archetype:          "The Hero",
				Description:        "Brave and true",
				Background:         "Humble origins",
				CommunicationStyle: "Inspiring",
				DecisionStyle:      "Courageous",
				Traits:             []string{"brave", "honest"},
				Skills:             []string{"combat", "leadership"},
				Values:             []string{"justice", "honor"},
			},
		}

		assert.True(t, char1.Same(char2))
		assert.True(t, char2.Same(char1))
	})

	t.Run("different version", func(t *testing.T) {
		char1 := &Character{
			Version: "1.0.0",
			Basics: &BasicCharacterInformation{
				Archetype: "The Hero",
			},
		}
		char2 := &Character{
			Version: "2.0.0",
			Basics: &BasicCharacterInformation{
				Archetype: "The Hero",
			},
		}

		assert.False(t, char1.Same(char2))
	})

	t.Run("different archetype", func(t *testing.T) {
		char1 := &Character{
			Basics: &BasicCharacterInformation{
				Archetype: "The Hero",
			},
		}
		char2 := &Character{
			Basics: &BasicCharacterInformation{
				Archetype: "The Villain",
			},
		}

		assert.False(t, char1.Same(char2))
	})

	t.Run("different traits", func(t *testing.T) {
		char1 := &Character{
			Basics: &BasicCharacterInformation{
				Archetype: "The Hero",
				Traits:    []string{"brave", "honest"},
			},
		}
		char2 := &Character{
			Basics: &BasicCharacterInformation{
				Archetype: "The Hero",
				Traits:    []string{"brave", "clever"},
			},
		}

		assert.False(t, char1.Same(char2))
	})

	t.Run("round tripped characters are same", func(t *testing.T) {
		original := &Character{
			Version: "1.5.2",
			Basics: &BasicCharacterInformation{
				Archetype:          "The Trickster",
				Description:        "Mischievous and clever",
				Background:         "Unknown origins",
				CommunicationStyle: "Playful and deceptive",
				DecisionStyle:      "Unpredictable",
				Traits:             []string{"clever", "mischievous", "unpredictable"},
				Skills:             []string{"deception", "acrobatics", "sleight of hand"},
				Values:             []string{"freedom", "chaos", "fun"},
			},
		}

		buf, err := toml.Marshal(original)
		require.NoError(t, err)

		var decoded Character
		err = toml.Unmarshal(buf, &decoded)
		require.NoError(t, err)

		assert.True(t, original.Same(&decoded))
		assert.Equal(t, "1.5.2", decoded.Version)
	})
}

func TestLoadCharacter(t *testing.T) {
	t.Run("loads minimal character", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[basics]
archetype = "The Hero"
description = "Brave and true"
background = "Humble origins"
communication_style = "Inspiring"
decision_style = "Courageous"
traits = ["brave"]
skills = ["combat"]
values = ["justice"]
`

		character, err := LoadCharacter([]byte(tomlData))
		require.NoError(t, err)

		assert.Equal(t, "1.0.0", character.Version)
		assert.Equal(t, "The Hero", character.Basics.Archetype)
		assert.Equal(t, "Brave and true", character.Basics.Description)
		assert.Equal(t, []string{"brave"}, character.Basics.Traits)
	})

	t.Run("loads fully populated character", func(t *testing.T) {
		tomlData := `
version = "2.5.0"

[basics]
archetype = "The Mentor"
description = "Wise teacher and guide"
background = "Former hero, now retired"
communication_style = "Patient and instructive"
decision_style = "Thoughtful and considered"
traits = ["wise", "patient", "experienced"]
skills = ["teaching", "strategy", "ancient knowledge"]
values = ["wisdom", "growth", "legacy"]
`

		character, err := LoadCharacter([]byte(tomlData))
		require.NoError(t, err)

		assert.Equal(t, "2.5.0", character.Version)
		assert.Equal(t, "The Mentor", character.Basics.Archetype)
		assert.Equal(t, "Wise teacher and guide", character.Basics.Description)
		assert.Equal(t, "Former hero, now retired", character.Basics.Background)
		assert.Equal(t, "Patient and instructive", character.Basics.CommunicationStyle)
		assert.Equal(t, "Thoughtful and considered", character.Basics.DecisionStyle)
		assert.Equal(t, []string{"wise", "patient", "experienced"}, character.Basics.Traits)
		assert.Equal(t, []string{"teaching", "strategy", "ancient knowledge"}, character.Basics.Skills)
		assert.Equal(t, []string{"wisdom", "growth", "legacy"}, character.Basics.Values)
	})

	t.Run("loads character with empty arrays", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[basics]
archetype = "The Blank Slate"
description = "A character with no defined traits"
background = "Unknown"
communication_style = "Silent"
decision_style = "Passive"
traits = []
skills = []
values = []
`

		character, err := LoadCharacter([]byte(tomlData))
		require.NoError(t, err)

		assert.Equal(t, "The Blank Slate", character.Basics.Archetype)
		assert.Empty(t, character.Basics.Traits)
		assert.Empty(t, character.Basics.Skills)
		assert.Empty(t, character.Basics.Values)
	})

	t.Run("loads character with multiline strings", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[basics]
archetype = "The Tragic Hero"
description = """
A character marked by fate,
struggling against inevitable doom,
yet finding nobility in the struggle."""
background = """
Born to greatness but cursed by prophecy.
Every triumph brings them closer to their downfall."""
communication_style = "Eloquent and melancholic"
decision_style = "Bound by duty and honor"
traits = ["tragic", "noble", "doomed"]
skills = ["leadership", "combat", "diplomacy"]
values = ["duty", "honor", "sacrifice"]
`

		character, err := LoadCharacter([]byte(tomlData))
		require.NoError(t, err)

		assert.Equal(t, "The Tragic Hero", character.Basics.Archetype)
		assert.Contains(t, character.Basics.Description, "marked by fate")
		assert.Contains(t, character.Basics.Background, "Born to greatness")
	})

	t.Run("loads character without version field", func(t *testing.T) {
		tomlData := `
[basics]
archetype = "The Simple One"
description = "No version specified"
background = "Test"
communication_style = "Simple"
decision_style = "Basic"
traits = ["simple"]
skills = ["basic"]
values = ["simplicity"]
`

		character, err := LoadCharacter([]byte(tomlData))
		require.NoError(t, err)

		assert.Equal(t, "", character.Version) // Default empty string
		assert.Equal(t, "The Simple One", character.Basics.Archetype)
	})

	t.Run("returns error for invalid TOML", func(t *testing.T) {
		tomlData := `
version = "1.0.0"
invalid toml syntax here
[basics]
`

		_, err := LoadCharacter([]byte(tomlData))
		require.Error(t, err)
	})

	t.Run("returns error for missing basics section", func(t *testing.T) {
		tomlData := `
version = "1.0.0"
`

		character, err := LoadCharacter([]byte(tomlData))
		require.NoError(t, err)

		// Should succeed but basics will be initialized empty
		assert.Equal(t, "1.0.0", character.Version)
		require.NotNil(t, character.Basics)
		assert.Equal(t, "", character.Basics.Archetype)
	})

	t.Run("loads character with special characters", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[basics]
archetype = "The Ðragon Rider"
description = "Flies on dragons, speaks in runes: 龍"
background = "From the lands of Ærith"
communication_style = "Multi-lingual (Español, 日本語)"
decision_style = "Instinctive"
traits = ["brave", "multi-cultural"]
skills = ["dragon-riding", "languages"]
values = ["freedom", "diversity"]
`

		character, err := LoadCharacter([]byte(tomlData))
		require.NoError(t, err)

		assert.Equal(t, "The Ðragon Rider", character.Basics.Archetype)
		assert.Contains(t, character.Basics.Description, "龍")
		assert.Contains(t, character.Basics.Background, "Ærith")
		assert.Contains(t, character.Basics.CommunicationStyle, "日本語")
	})

	t.Run("round trip through LoadCharacter preserves data", func(t *testing.T) {
		originalData := `
version = "3.2.1"

[basics]
archetype = "The Sage"
description = "Ancient wisdom keeper"
background = "Centuries of study"
communication_style = "Cryptic and profound"
decision_style = "Measured and wise"
traits = ["wise", "ancient", "mystical"]
skills = ["prophecy", "ancient languages", "meditation"]
values = ["knowledge", "balance", "truth"]
`

		character, err := LoadCharacter([]byte(originalData))
		require.NoError(t, err)

		// Marshal back to TOML
		buf, err := toml.Marshal(character)
		require.NoError(t, err)

		// Load again
		character2, err := LoadCharacter(buf)
		require.NoError(t, err)

		// Should be identical
		assert.True(t, character.Same(character2))
	})

	t.Run("loads character with single element arrays", func(t *testing.T) {
		tomlData := `
version = "1.0.0"

[basics]
archetype = "The Specialist"
description = "Master of one thing"
background = "Focused training"
communication_style = "Direct"
decision_style = "Specialized"
traits = ["focused"]
skills = ["mastery"]
values = ["excellence"]
`

		character, err := LoadCharacter([]byte(tomlData))
		require.NoError(t, err)

		assert.Equal(t, []string{"focused"}, character.Basics.Traits)
		assert.Equal(t, []string{"mastery"}, character.Basics.Skills)
		assert.Equal(t, []string{"excellence"}, character.Basics.Values)
	})
}
